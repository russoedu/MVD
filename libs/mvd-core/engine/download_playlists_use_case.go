package engine

import (
	"context"
	"fmt"
	"os"
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
	YtDlp          string   // path to the yt-dlp executable
	URLs           []string // playlists, in order
	OutputDir      string
	OutputTemplate string
	// PartialDir, when set, is where videos are downloaded and merged before they are
	// moved to OutputDir, so a failed or cancelled download leaves nothing in the
	// library. The engine empties it when a run starts and removes it when it ends.
	PartialDir string
	// ListWorkers, NameWorkers and PickWorkers are how many playlists are listed, uploads
	// identified and versions picked at the same time, ahead of the downloads (Workers).
	// Each stage has its own queue, so a slow one holds the others back no more than it
	// must. 0 uses 4, 8 and 4.
	ListWorkers         int
	NameWorkers         int
	PickWorkers         int
	Quality             string
	MergeOutputFormat   string
	ConcurrentFragments int // yt-dlp --concurrent-fragments per video; 0 skips
	ExtraArgs           []string
	Workers             int    // videos in flight across all playlists
	LogPath             string // "" disables the log file
	Resolver            OfficialResolver
	// Tracks, when set, lists the playlists of another service and finds the
	// YouTube video of each track.
	Tracks TrackSource
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
	// identified is what the naming stage learned, for the picking stage.
	identified *Identification
	official   bool
	better     bool   // replaced by an upload of better quality, not the official video
	track      *Track // set when the entry is a song with no video yet
	err        string
	deferred   bool // failed with a transient error, eligible for the sweep
	sweeps     int  // deferred retry passes already spent
}

// Engine lists playlists, resolves official videos and runs yt-dlp per
// entry with a bounded pool of workers. All observable state is reported
// through Events().
type Engine struct {
	opts    Options
	sources []PlaylistSource

	events chan interface{}
	// The pipeline: entries go through the naming queue, the picking queue and then the
	// download queue (queue), each with its own workers.
	nameQueue *taskQueue
	pickQueue *taskQueue
	queue     *taskQueue
	logger    *runLogger
	drained   chan struct{} // pinged when the queue empties

	mu          sync.Mutex
	entries     []*engineEntry
	claims      map[int]map[string]int // playlist -> target id -> entry id
	pending     int                    // queued + in flight
	listing     bool                   // producer still listing playlists
	listsActive int                    // playlists being listed right now
	idleSent    bool
	toList      []int         // sources waiting to be listed, in the order they arrived
	wake        chan struct{} // pinged when AddSource gives a waiting producer work
	closed      bool          // Run has finished: no more events will be sent
}

// New builds an engine. It opens the log file right away.
func New(opts Options) (*Engine, error) {
	logger, err := newRunLogger(opts.LogPath)
	if err != nil {
		return nil, err
	}
	e := &Engine{
		opts:      opts,
		events:    make(chan interface{}, 1024),
		nameQueue: newTaskQueue(),
		pickQueue: newTaskQueue(),
		queue:     newTaskQueue(),
		logger:    logger,
		drained:   make(chan struct{}, 1),
		wake:      make(chan struct{}, 1),
		claims:    make(map[int]map[string]int),
	}
	for i, u := range opts.URLs {
		e.sources = append(e.sources, PlaylistSource{Index: i, URL: u})
		e.toList = append(e.toList, i)
	}
	return e, nil
}

// Events returns the channel renderers read from.
func (e *Engine) Events() <-chan interface{} { return e.events }

// Sources returns the playlists in order, including any added while running.
func (e *Engine) Sources() []PlaylistSource {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]PlaylistSource(nil), e.sources...)
}

// AddSource queues one more playlist or video for an engine that is already
// running, or one that has not started yet. It gets the next free index, so
// indices stay stable, and renderers learn of it from an EvPlaylistAdded event
// before its listing starts.
//
// It returns false, and queues nothing, once Run has finished: the events
// channel is closed by then. A caller that serves requests should stop
// accepting them before it cancels the engine, because a call that is already
// past that check can still lose the race with the close.
func (e *Engine) AddSource(url string) (PlaylistSource, bool) {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return PlaylistSource{}, false
	}
	src := PlaylistSource{Index: len(e.sources), URL: url}
	e.sources = append(e.sources, src)
	e.toList = append(e.toList, src.Index)
	// The producer has work again, so a quiet period that was announced is
	// over and the engine is not idle until this is listed and downloaded.
	// Both flags change under the lock the producer takes to look for work,
	// so no idle event can slip in between the add and the listing.
	e.idleSent = false
	e.listing = true
	e.mu.Unlock()

	e.emit(EvPlaylistAdded{Source: src})
	select {
	case e.wake <- struct{}{}:
	default:
	}
	return src, true
}

// nextToList hands the producer the next source to list. When there is none it
// marks listing as finished, under the same lock AddSource takes, unless other
// lists are still being read.
func (e *Engine) nextToList() (PlaylistSource, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.toList) == 0 {
		e.listing = e.listsActive > 0
		return PlaylistSource{}, false
	}
	idx := e.toList[0]
	e.toList = e.toList[1:]
	e.listsActive++
	e.listing = true
	return e.sources[idx], true
}

// listDone is called when one list is read, so the engine knows when listing is over.
func (e *Engine) listDone() {
	e.mu.Lock()
	e.listsActive--
	e.listing = e.listsActive > 0 || len(e.toList) > 0
	e.mu.Unlock()
	e.signalDrain()
}

// LogPath returns the path of the log file, or "" when disabled.
func (e *Engine) LogPath() string { return e.logger.path }

// Run lists playlists, downloads everything and keeps serving retries
// until ctx is cancelled. It returns once all workers have exited.
func (e *Engine) Run(ctx context.Context) {
	if dir := e.opts.PartialDir; dir != "" {
		_ = os.RemoveAll(dir) // what an earlier run that did not end well left
		_ = os.MkdirAll(dir, 0o755)
		defer func() { _ = os.RemoveAll(dir) }()
	}

	workers := atLeast(e.opts.Workers, 1)
	listWorkers := orDefault(e.opts.ListWorkers, 4)
	nameWorkers := orDefault(e.opts.NameWorkers, 8)
	pickWorkers := orDefault(e.opts.PickWorkers, 4)

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

	// A pool takes entries from a queue until it is closed and empty.
	pool := func(n int, queue *taskQueue, work func(context.Context, int)) *sync.WaitGroup {
		var wg sync.WaitGroup
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					id, ok := queue.Pop(ctx)
					if !ok {
						return
					}
					work(ctx, id)
				}
			}()
		}
		return &wg
	}
	naming := pool(nameWorkers, e.nameQueue, e.nameStage)
	picking := pool(pickWorkers, e.pickQueue, e.pickStage)
	downloading := pool(workers, e.queue, func(ctx context.Context, id int) {
		e.process(ctx, id)
		e.finishOne()
	})

	// Producer: list playlists, a few at a time, feeding the pipeline as each one
	// comes back so downloads start while later playlists are still listed.
	// When there is nothing left to list it waits, so AddSource can feed an
	// engine that is already running; it stops only when ctx is cancelled.
	slots := make(chan struct{}, listWorkers)
	var listing sync.WaitGroup
	for ctx.Err() == nil {
		select {
		case slots <- struct{}{}:
		case <-ctx.Done():
			continue
		}
		src, ok := e.nextToList()
		if !ok {
			<-slots
			e.signalDrain()
			select {
			case <-e.wake:
			case <-ctx.Done():
			}
			continue
		}
		listing.Add(1)
		go func() {
			defer listing.Done()
			defer func() { <-slots }()
			defer e.listDone()
			e.listPlaylist(ctx, src)
		}()
	}

	// Close the stages front to back, each one emptied before the next is closed, so
	// nothing is handed to a queue nobody reads.
	listing.Wait()
	e.nameQueue.Close()
	naming.Wait()
	e.pickQueue.Close()
	picking.Wait()
	e.queue.Close()
	downloading.Wait()
	<-coordDone
	e.logger.Close()
	e.mu.Lock()
	e.closed = true
	e.mu.Unlock()
	close(e.events)
}

// atLeast returns n, or floor when n is smaller.
func atLeast(n, floor int) int {
	if n < floor {
		return floor
	}
	return n
}

// orDefault returns n, or fallback when n is not set.
func orDefault(n, fallback int) int {
	if n < 1 {
		return fallback
	}
	return n
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

	e.emit(EvEntryState{Entry: entryID, State: StateQueued, TargetID: en.targetID, Official: en.official, Better: en.better})
	e.log(en.info.Playlist, entryID, "retry requested")
	e.nameQueue.Push(entryID)
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

	entries, tracks, err := e.listEntries(ctx, src)
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
		en := &engineEntry{info: info, raw: pe, state: StateQueued}
		if tracks != nil {
			en.track = &tracks[i]
		}
		e.entries = append(e.entries, en)
		infos = append(infos, info)
		ids = append(ids, info.ID)
	}
	e.pending += len(ids)
	e.mu.Unlock()

	e.log(src.Index, -1, "%q: %d entries", title, len(entries))
	e.emit(EvPlaylistListed{Playlist: src.Index, Title: title, Entries: infos})
	for _, id := range ids {
		e.nameQueue.Push(id)
	}
}

func (e *Engine) setState(en *engineEntry, st EntryState, err string) {
	e.mu.Lock()
	en.state = st
	en.err = err
	ev := EvEntryState{Entry: en.info.ID, State: st, TargetID: en.targetID, Official: en.official, Better: en.better, Err: err}
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

// nameStage is the first stage of the pipeline: it names the song of an upload, or
// passes it on when there is nothing to name.
func (e *Engine) nameStage(ctx context.Context, id int) {
	e.mu.Lock()
	en := e.entries[id]
	en.targetID = en.info.VideoID
	en.official = false
	en.better = false
	en.identified = nil
	e.mu.Unlock()

	pl, eid := en.info.Playlist, en.info.ID
	if ctx.Err() != nil {
		e.cancelEntry(en)
		return
	}

	switch {
	case en.track != nil:
		e.pickQueue.Push(id) // a song of another service is found by its own search
	case e.opts.Resolver != nil && e.opts.Resolver.Wanted(en.info.Title, en.raw.Channel, en.raw.Uploader):
		e.setState(en, StateResolving, "")
		channel := en.raw.Channel
		if channel == "" {
			channel = en.raw.Uploader
		}
		identification := e.opts.Resolver.Identify(en.info.VideoID, en.raw.Title, channel, int(en.raw.Duration), e.entryLog(pl, eid))
		if identification.Done {
			e.applyResolution(en, identification.Resolution)
			e.queue.Push(id)
			return
		}
		e.mu.Lock()
		en.identified = &identification
		e.mu.Unlock()
		e.pickQueue.Push(id)
	default:
		e.queue.Push(id)
	}
}

// pickStage is the second stage: it finds the version of the video to download, the
// official one or else the best quality.
func (e *Engine) pickStage(ctx context.Context, id int) {
	e.mu.Lock()
	en := e.entries[id]
	identification := en.identified
	e.mu.Unlock()

	pl, eid := en.info.Playlist, en.info.ID
	if ctx.Err() != nil {
		e.cancelEntry(en)
		return
	}

	switch {
	case en.track != nil:
		if !e.findTrack(ctx, en) {
			e.finishOne()
			return
		}
		e.log(pl, eid, "%s - %s -> %s (official: %v)", en.track.Artist, en.track.Title, en.targetID, en.official)
	case identification != nil && e.opts.Resolver != nil:
		e.applyResolution(en, e.opts.Resolver.Pick(*identification, e.entryLog(pl, eid)))
	}
	e.queue.Push(id)
}

// cancelEntry ends an entry that was still waiting for a stage when the run was cancelled.
func (e *Engine) cancelEntry(en *engineEntry) {
	e.setState(en, StateFailed, "cancelled")
	e.finishOne()
}

// entryLog returns a log function for the lookup of one entry.
func (e *Engine) entryLog(pl, eid int) func(format string, a ...interface{}) {
	return func(format string, a ...interface{}) {
		e.log(pl, eid, "%s", strings.TrimSpace(fmt.Sprintf(format, a...)))
	}
}

// applyResolution records the version a lookup chose for an entry.
func (e *Engine) applyResolution(en *engineEntry, res Resolution) {
	pl, eid := en.info.Playlist, en.info.ID
	if res.VideoID == "" {
		e.log(pl, eid, "kept original (%s)", res.Reason)
		return
	}
	e.mu.Lock()
	en.targetID = res.VideoID
	en.official = res.Official
	en.better = !res.Official
	e.mu.Unlock()
	e.log(pl, eid, "%s -> %s (%s)", en.info.VideoID, res.VideoID, res.Reason)
}

// process is the last stage: it downloads the version the earlier stages chose.
func (e *Engine) process(ctx context.Context, id int) {
	e.mu.Lock()
	en := e.entries[id]
	e.mu.Unlock()

	pl, eid := en.info.Playlist, en.info.ID

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
		if (en.official || en.better) && en.info.VideoID != "" {
			e.log(pl, eid, "replacement video %s failed (%v); downloading the original instead", en.targetID, err)
			e.mu.Lock()
			en.targetID = en.info.VideoID
			en.official = false
			en.better = false
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

	template := ytdlp.ApplyPlaylistFields(e.opts.OutputTemplate, en.raw)
	options := ytdlp.DownloadOptions{
		Format:              e.opts.Quality,
		OutputTemplate:      filepath.Join(e.opts.OutputDir, template),
		MergeOutputFormat:   e.opts.MergeOutputFormat,
		ConcurrentFragments: e.opts.ConcurrentFragments,
		ExtraArgs:           e.opts.ExtraArgs,
	}
	if e.opts.PartialDir != "" {
		options.OutputTemplate, options.HomeDir, options.TempDir = template, e.opts.OutputDir, e.opts.PartialDir
	}
	args := ytdlp.DownloadArgs(options,
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
				e.emit(EvEntryState{Entry: eid, State: StateMerging, TargetID: en.targetID, Official: en.official, Better: en.better})
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

	e.emit(EvEntryState{Entry: id, State: StateQueued, TargetID: en.targetID, Official: en.official, Better: en.better})
	e.log(en.info.Playlist, id, "auto-retry after cooldown")
	e.nameQueue.Push(id)
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
