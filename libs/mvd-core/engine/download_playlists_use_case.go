package engine

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"youtube-downloader/libs/mvd-core/ytdlp"
)

// defaultRetryCooldown is how long the engine waits before the deferred
// retry sweep, so rate-limit windows have time to reset.
const defaultRetryCooldown = 20 * time.Second

// maxSweeps is how many deferred retry passes a failed entry gets.
const maxSweeps = 1

// Options configures an Engine.
type Options struct {
	YtDlp               string   // path to the yt-dlp executable
	URLs                []string // playlists, in order
	OutputDir           string
	OutputTemplate      string
	Quality             string
	MergeOutputFormat   string
	ConcurrentFragments int // yt-dlp --concurrent-fragments per video; 0 skips
	ExtraArgs           []string
	Workers             int    // videos in flight across all playlists
	LogPath             string // "" disables the log file
	Resolver            OfficialResolver
	// AutoRetry turns on immediate one-off retries and the deferred sweep.
	AutoRetry bool
	// RetryCooldown is the wait before each deferred sweep. 0 uses the default.
	RetryCooldown time.Duration
}

type engineEntry struct {
	info     EntryInfo
	raw      ytdlp.PlaylistEntry
	state    EntryState
	targetID string
	official bool
	err      string
	deferred bool // failed with a transient error, eligible for the sweep
	sweeps   int  // deferred retry passes already spent
}

// Engine lists playlists, resolves official videos and runs yt-dlp per
// entry with a bounded pool of workers. All observable state is reported
// through Events().
type Engine struct {
	opts    Options
	sources []PlaylistSource

	events  chan interface{}
	queue   *taskQueue
	logger  *runLogger
	drained chan struct{} // pinged when the queue empties

	mu       sync.Mutex
	entries  []*engineEntry
	claims   map[int]map[string]int // playlist -> target id -> entry id
	pending  int                    // queued + in flight
	listing  bool                   // producer still listing playlists
	idleSent bool
}

// New builds an engine. It opens the log file right away.
func New(opts Options) (*Engine, error) {
	logger, err := newRunLogger(opts.LogPath)
	if err != nil {
		return nil, err
	}
	e := &Engine{
		opts:    opts,
		events:  make(chan interface{}, 1024),
		queue:   newTaskQueue(),
		logger:  logger,
		drained: make(chan struct{}, 1),
		claims:  make(map[int]map[string]int),
	}
	for i, u := range opts.URLs {
		e.sources = append(e.sources, PlaylistSource{Index: i, URL: u})
	}
	return e, nil
}

// Events returns the channel renderers read from.
func (e *Engine) Events() <-chan interface{} { return e.events }

// Sources returns the playlists in order.
func (e *Engine) Sources() []PlaylistSource { return e.sources }

// LogPath returns the path of the log file, or "" when disabled.
func (e *Engine) LogPath() string { return e.logger.path }

// Run lists playlists, downloads everything and keeps serving retries
// until ctx is cancelled. It returns once all workers have exited.
func (e *Engine) Run(ctx context.Context) {
	workers := e.opts.Workers
	if workers < 1 {
		workers = 1
	}

	e.mu.Lock()
	e.listing = true
	e.mu.Unlock()

	// Coordinator: decides, each time the queue empties, whether to run a
	// deferred retry sweep or declare the run idle.
	coordDone := make(chan struct{})
	go func() {
		e.coordinate(ctx)
		close(coordDone)
	}()

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				id, ok := e.queue.Pop(ctx)
				if !ok {
					return
				}
				e.process(ctx, id)
				e.finishOne()
			}
		}()
	}

	// Producer: list playlists in order, feeding the queue as each one
	// comes back so downloads start while later playlists are still listed.
	for _, src := range e.sources {
		if ctx.Err() != nil {
			break
		}
		e.listPlaylist(ctx, src)
	}

	e.mu.Lock()
	e.listing = false
	e.mu.Unlock()
	e.signalDrain()

	<-ctx.Done()
	e.queue.Close()
	wg.Wait()
	<-coordDone
	e.logger.Close()
	close(e.events)
}

// coordinate waits for the queue to drain and then either runs a deferred
// retry sweep (after a cooldown) or emits the final idle event.
func (e *Engine) coordinate(ctx context.Context) {
	cooldown := e.opts.RetryCooldown
	if cooldown <= 0 {
		cooldown = defaultRetryCooldown
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-e.drained:
			ids := e.sweepCandidates()
			if len(ids) == 0 {
				e.markIdle()
				continue
			}
			e.mu.Lock()
			pl := e.entries[ids[0]].info.Playlist
			e.mu.Unlock()
			e.log(pl, -1, "cooling down %s before retrying %d rate-limited download(s)", cooldown, len(ids))
			if !sleepCtx(ctx, cooldown) {
				return
			}
			for _, id := range ids {
				e.autoRetry(id)
			}
		}
	}
}

// Retry re-queues a failed entry. It returns false when the entry is not
// in a failed state.
func (e *Engine) Retry(entryID int) bool {
	e.mu.Lock()
	if entryID < 0 || entryID >= len(e.entries) {
		e.mu.Unlock()
		return false
	}
	en := e.entries[entryID]
	if en.state != StateFailed {
		e.mu.Unlock()
		return false
	}
	en.state = StateQueued
	en.err = ""
	e.pending++
	e.idleSent = false
	e.mu.Unlock()

	e.emit(EvEntryState{Entry: entryID, State: StateQueued, TargetID: en.targetID, Official: en.official})
	e.log(en.info.Playlist, entryID, "retry requested")
	e.queue.Push(entryID)
	return true
}

// RetryPlaylist re-queues every failed entry of a playlist and returns
// how many were queued.
func (e *Engine) RetryPlaylist(playlist int) int {
	e.mu.Lock()
	var ids []int
	for _, en := range e.entries {
		if en.info.Playlist == playlist && en.state == StateFailed {
			ids = append(ids, en.info.ID)
		}
	}
	e.mu.Unlock()
	n := 0
	for _, id := range ids {
		if e.Retry(id) {
			n++
		}
	}
	return n
}

func (e *Engine) emit(ev interface{}) {
	e.events <- ev
}

func (e *Engine) log(playlist, entry int, format string, a ...interface{}) {
	line := fmt.Sprintf(format, a...)
	e.logger.Write(playlist, entry, line)
	e.emit(EvLog{Playlist: playlist, Entry: entry, Line: line})
}

func (e *Engine) listPlaylist(ctx context.Context, src PlaylistSource) {
	e.emit(EvPlaylistListing{Playlist: src.Index, URL: src.URL})
	e.log(src.Index, -1, "listing %s", src.URL)

	entries, err := ytdlp.ListPlaylist(ctx, e.opts.YtDlp, src.URL, e.opts.ExtraArgs)
	if err != nil {
		e.log(src.Index, -1, "listing failed: %v", err)
		e.emit(EvPlaylistFailed{Playlist: src.Index, Err: err.Error()})
		return
	}
	if len(entries) == 0 {
		e.log(src.Index, -1, "playlist is empty")
		e.emit(EvPlaylistFailed{Playlist: src.Index, Err: "playlist is empty or could not be listed"})
		return
	}

	title := entries[0].PlaylistTitle
	if title == "" {
		title = entries[0].Playlist
	}
	if title == "" {
		title = src.URL
	}

	e.mu.Lock()
	infos := make([]EntryInfo, 0, len(entries))
	ids := make([]int, 0, len(entries))
	for i, pe := range entries {
		idx := pe.PlaylistIndex
		if idx == 0 {
			idx = i + 1
		}
		info := EntryInfo{
			ID:       len(e.entries),
			Playlist: src.Index,
			Index:    idx,
			VideoID:  pe.ID,
			Title:    pe.Title,
			Channel:  pe.Channel,
		}
		e.entries = append(e.entries, &engineEntry{info: info, raw: pe, state: StateQueued})
		infos = append(infos, info)
		ids = append(ids, info.ID)
	}
	e.pending += len(ids)
	e.mu.Unlock()

	e.log(src.Index, -1, "%q: %d entries", title, len(entries))
	e.emit(EvPlaylistListed{Playlist: src.Index, Title: title, Entries: infos})
	for _, id := range ids {
		e.queue.Push(id)
	}
}

func (e *Engine) setState(en *engineEntry, st EntryState, err string) {
	e.mu.Lock()
	en.state = st
	en.err = err
	ev := EvEntryState{Entry: en.info.ID, State: st, TargetID: en.targetID, Official: en.official, Err: err}
	e.mu.Unlock()
	e.emit(ev)
}

// claimTarget records that targetID is being downloaded for a playlist.
// It returns false when another entry of the same playlist already owns it.
func (e *Engine) claimTarget(playlist int, target string, entryID int) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	m := e.claims[playlist]
	if m == nil {
		m = make(map[string]int)
		e.claims[playlist] = m
	}
	if owner, ok := m[target]; ok && owner != entryID {
		return false
	}
	m[target] = entryID
	return true
}

func (e *Engine) process(ctx context.Context, id int) {
	e.mu.Lock()
	en := e.entries[id]
	en.targetID = en.info.VideoID
	en.official = false
	e.mu.Unlock()

	pl, eid := en.info.Playlist, en.info.ID

	if e.opts.Resolver != nil && e.opts.Resolver.Wanted(en.raw.Channel, en.raw.Uploader) {
		e.setState(en, StateResolving, "")
		official, reason := e.opts.Resolver.ResolveLog(en.info.VideoID, func(format string, a ...interface{}) {
			e.log(pl, eid, "%s", strings.TrimSpace(fmt.Sprintf(format, a...)))
		})
		if official != "" {
			e.mu.Lock()
			en.targetID = official
			en.official = true
			e.mu.Unlock()
			e.log(pl, eid, "%s -> %s (%s)", en.info.VideoID, official, reason)
		} else {
			e.log(pl, eid, "kept original (%s)", reason)
		}
	}

	if ctx.Err() != nil {
		e.setState(en, StateFailed, "cancelled")
		return
	}

	if !e.claimTarget(pl, en.targetID, eid) {
		e.log(pl, eid, "%s already downloaded for this playlist, skipping duplicate", en.targetID)
		e.setState(en, StateDuplicate, "")
		return
	}

	e.setState(en, StateDownloading, "")

	immediateUsed := false
	for {
		err := e.download(ctx, en)
		if err == nil {
			e.setState(en, StateDone, "")
			return
		}
		if ctx.Err() != nil {
			e.setState(en, StateFailed, "cancelled")
			return
		}

		// The official video could not be downloaded: fall back to the art
		// track itself rather than losing the entry.
		if en.official {
			e.log(pl, eid, "official video %s failed (%v); downloading the original instead", en.targetID, err)
			e.mu.Lock()
			en.targetID = en.info.VideoID
			en.official = false
			e.mu.Unlock()
			if !e.claimTarget(pl, en.targetID, eid) {
				e.log(pl, eid, "%s already downloaded for this playlist, skipping duplicate", en.targetID)
				e.setState(en, StateDuplicate, "")
				return
			}
			e.setState(en, StateDownloading, "")
			continue
		}

		if !e.opts.AutoRetry {
			e.setState(en, StateFailed, err.Error())
			return
		}

		switch classify(err.Error()) {
		case ClassImmediate:
			if !immediateUsed {
				immediateUsed = true
				e.log(pl, eid, "retrying now (%v)", err)
				e.setState(en, StateDownloading, "")
				continue
			}
			e.setState(en, StateFailed, err.Error())
			return
		case ClassDeferred:
			e.mu.Lock()
			en.deferred = true
			e.mu.Unlock()
			e.setState(en, StateFailed, err.Error())
			return
		default: // ClassPermanent
			e.setState(en, StateFailed, err.Error())
			return
		}
	}
}

// download runs yt-dlp for the entry's current target, turning its output
// into progress, merging and log events.
func (e *Engine) download(ctx context.Context, en *engineEntry) error {
	pl, eid := en.info.Playlist, en.info.ID

	e.mu.Lock()
	target := en.targetID
	e.mu.Unlock()

	outPattern := filepath.Join(e.opts.OutputDir, e.opts.OutputTemplate)
	args := ytdlp.DownloadArgs(ytdlp.DownloadOptions{
		Format:              e.opts.Quality,
		OutputTemplate:      ytdlp.ApplyPlaylistFields(outPattern, en.raw),
		MergeOutputFormat:   e.opts.MergeOutputFormat,
		ConcurrentFragments: e.opts.ConcurrentFragments,
		ExtraArgs:           e.opts.ExtraArgs,
	},
		"--newline",
		"--progress-template", ytdlp.ProgressTemplate,
		"--no-playlist",
		"https://www.youtube.com/watch?v="+target,
	)

	return ytdlp.Download(ctx, e.opts.YtDlp, args, func(line string) {
		if p, ok := ytdlp.ParseProgressLine(line); ok {
			e.emit(EvProgress{Entry: eid, Percent: p.Percent, Downloaded: p.Downloaded, Total: p.Total, Speed: p.Speed, ETA: p.ETA})
			return
		}
		if ytdlp.IsPostProcessLine(line) {
			e.mu.Lock()
			if en.state == StateDownloading {
				en.state = StateMerging
				e.mu.Unlock()
				e.emit(EvEntryState{Entry: eid, State: StateMerging, TargetID: en.targetID, Official: en.official})
			} else {
				e.mu.Unlock()
			}
		}
		e.log(pl, eid, "%s", line)
	})
}

func (e *Engine) finishOne() {
	e.mu.Lock()
	e.pending--
	e.mu.Unlock()
	e.signalDrain()
}

// signalDrain wakes the coordinator when the queue has emptied.
func (e *Engine) signalDrain() {
	e.mu.Lock()
	drained := !e.listing && e.pending == 0
	e.mu.Unlock()
	if drained {
		select {
		case e.drained <- struct{}{}:
		default:
		}
	}
}

// markIdle emits the idle event once per quiet period.
func (e *Engine) markIdle() {
	e.mu.Lock()
	already := e.idleSent
	e.idleSent = true
	e.mu.Unlock()
	if !already {
		e.emit(EvIdle{})
	}
}

// sweepCandidates returns the failed entries still eligible for a deferred
// retry pass.
func (e *Engine) sweepCandidates() []int {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.opts.AutoRetry {
		return nil
	}
	var ids []int
	for _, en := range e.entries {
		if en.deferred && en.state == StateFailed && en.sweeps < maxSweeps {
			ids = append(ids, en.info.ID)
		}
	}
	return ids
}

// autoRetry re-queues a deferred failure for the sweep, counting the pass.
func (e *Engine) autoRetry(id int) {
	e.mu.Lock()
	en := e.entries[id]
	if en.state != StateFailed {
		e.mu.Unlock()
		return
	}
	en.state = StateQueued
	en.err = ""
	en.sweeps++
	e.pending++
	e.idleSent = false
	e.mu.Unlock()

	e.emit(EvEntryState{Entry: id, State: StateQueued, TargetID: en.targetID, Official: en.official})
	e.log(en.info.Playlist, id, "auto-retry after cooldown")
	e.queue.Push(id)
}

// sleepCtx waits for d or until ctx is cancelled. It returns false on cancel.
func sleepCtx(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
