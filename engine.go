package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// EntryState is the life cycle of one playlist entry.
type EntryState int

const (
	StateQueued EntryState = iota
	StateResolving
	StateDownloading
	StateMerging
	StateDone
	StateDuplicate
	StateFailed
)

func (s EntryState) String() string {
	switch s {
	case StateQueued:
		return "queued"
	case StateResolving:
		return "resolving"
	case StateDownloading:
		return "downloading"
	case StateMerging:
		return "merging"
	case StateDone:
		return "done"
	case StateDuplicate:
		return "duplicate"
	case StateFailed:
		return "failed"
	}
	return "unknown"
}

// finished reports whether the state is terminal.
func (s EntryState) finished() bool {
	return s == StateDone || s == StateDuplicate || s == StateFailed
}

// EntryInfo is the immutable description of an entry handed to renderers.
type EntryInfo struct {
	ID       int // global, sequential
	Playlist int // index into the playlist list
	Index    int // position inside the playlist (1 based)
	VideoID  string
	Title    string
	Channel  string
}

// Events emitted by the Engine. Renderers (TUI, plain log) consume them.
type (
	// EvPlaylistListed is sent once the flat listing of a playlist is known.
	EvPlaylistListed struct {
		Playlist int
		Title    string
		Entries  []EntryInfo
	}
	// EvPlaylistFailed is sent when a playlist could not be listed.
	EvPlaylistFailed struct {
		Playlist int
		Err      string
	}
	// EvEntryState is sent on every state transition of an entry.
	EvEntryState struct {
		Entry    int
		State    EntryState
		TargetID string // video actually downloaded (official or original)
		Official bool   // true when TargetID is the official video
		Err      string // set for StateFailed
	}
	// EvProgress carries a download progress tick.
	EvProgress struct {
		Entry      int
		Percent    float64
		Downloaded int64
		Total      int64
		Speed      float64 // bytes per second, 0 when unknown
		ETA        int     // seconds, -1 when unknown
	}
	// EvLog is one line of output. Entry is -1 for playlist level lines.
	EvLog struct {
		Playlist int
		Entry    int
		Line     string
	}
	// EvIdle is sent when the queue is empty and no worker is busy.
	EvIdle struct{}
)

// PlaylistSource is one URL from downloads.conf.
type PlaylistSource struct {
	Index int
	URL   string
}

type engineEntry struct {
	info     EntryInfo
	raw      PlaylistEntry
	state    EntryState
	targetID string
	official bool
	err      string
}

// Engine lists playlists, resolves official videos and runs yt-dlp per
// entry with a bounded pool of workers. All observable state is reported
// through Events().
type Engine struct {
	cfg      Config
	ytDlp    string
	sources  []PlaylistSource
	resolver *OfficialResolver

	events chan interface{}
	queue  *taskQueue
	logger *runLogger

	mu       sync.Mutex
	entries  []*engineEntry
	claims   map[int]map[string]int // playlist -> target id -> entry id
	pending  int                    // queued + in flight
	listing  bool                   // producer still listing playlists
	idleSent bool
}

// NewEngine builds an engine. logPath may be empty to disable the log file.
func NewEngine(cfg Config, ytDlp string, urls []string, logPath string) (*Engine, error) {
	logger, err := newRunLogger(logPath)
	if err != nil {
		return nil, err
	}
	e := &Engine{
		cfg:    cfg,
		ytDlp:  ytDlp,
		events: make(chan interface{}, 1024),
		queue:  newTaskQueue(),
		logger: logger,
		claims: make(map[int]map[string]int),
	}
	for i, u := range urls {
		e.sources = append(e.sources, PlaylistSource{Index: i, URL: u})
	}
	if cfg.DownloadOfficialMusicVideo {
		e.resolver = newOfficialResolver(nil)
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
	workers := e.cfg.MaxConcurrentDownloads
	if workers < 1 {
		workers = 1
	}

	e.mu.Lock()
	e.listing = true
	e.mu.Unlock()

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
	e.checkIdle()

	<-ctx.Done()
	e.queue.Close()
	wg.Wait()
	e.logger.Close()
	close(e.events)
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
	e.log(src.Index, -1, "listing %s", src.URL)

	entries, err := listPlaylistEntriesCtx(ctx, e.ytDlp, src.URL, e.cfg.ExtraArgs)
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

	if e.resolver != nil && isAutoGenerated(en.raw) {
		e.setState(en, StateResolving, "")
		official, reason := e.resolver.ResolveLog(en.info.VideoID, func(format string, a ...interface{}) {
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

	outPattern := filepath.Join(e.cfg.OutputDir, e.cfg.OutputTemplate)
	args := buildDownloadArgs(e.cfg, applyPlaylistFields(outPattern, en.raw),
		"--newline",
		"--progress-template", progressTemplate,
		"--no-playlist",
		"https://www.youtube.com/watch?v="+en.targetID,
	)

	err := e.runYtDlp(ctx, en, args)
	if err != nil {
		e.setState(en, StateFailed, err.Error())
		return
	}
	e.setState(en, StateDone, "")
}

// runYtDlp runs yt-dlp with stdout and stderr captured line by line.
func (e *Engine) runYtDlp(ctx context.Context, en *engineEntry, args []string) error {
	pl, eid := en.info.Playlist, en.info.ID

	cmd := exec.CommandContext(ctx, e.ytDlp, args...)
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		pw.Close()
		return fmt.Errorf("cannot start yt-dlp: %w", err)
	}

	lastError := ""
	scanDone := make(chan struct{})
	go func() {
		defer close(scanDone)
		scanner := bufio.NewScanner(pr)
		scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
		for scanner.Scan() {
			line := strings.TrimRight(scanner.Text(), "\r")
			if line == "" {
				continue
			}
			if p, ok := parseProgressLine(line); ok {
				p.Entry = eid
				e.emit(p)
				continue
			}
			if isPostProcessLine(line) {
				e.mu.Lock()
				if en.state == StateDownloading {
					en.state = StateMerging
					e.mu.Unlock()
					e.emit(EvEntryState{Entry: eid, State: StateMerging, TargetID: en.targetID, Official: en.official})
				} else {
					e.mu.Unlock()
				}
			}
			if strings.HasPrefix(line, "ERROR:") {
				lastError = strings.TrimSpace(strings.TrimPrefix(line, "ERROR:"))
			}
			e.log(pl, eid, "%s", line)
		}
	}()

	waitErr := cmd.Wait()
	pw.Close()
	<-scanDone

	if waitErr != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("cancelled")
		}
		if lastError != "" {
			return fmt.Errorf("%s", lastError)
		}
		return waitErr
	}
	return nil
}

func (e *Engine) finishOne() {
	e.mu.Lock()
	e.pending--
	e.mu.Unlock()
	e.checkIdle()
}

func (e *Engine) checkIdle() {
	e.mu.Lock()
	idle := !e.listing && e.pending == 0 && !e.idleSent
	if idle {
		e.idleSent = true
	}
	e.mu.Unlock()
	if idle {
		e.emit(EvIdle{})
	}
}

// --- queue ----------------------------------------------------------------

// taskQueue is an unbounded FIFO of entry ids with blocking Pop.
type taskQueue struct {
	mu     sync.Mutex
	cond   *sync.Cond
	items  []int
	closed bool
}

func newTaskQueue() *taskQueue {
	q := &taskQueue{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *taskQueue) Push(id int) {
	q.mu.Lock()
	q.items = append(q.items, id)
	q.mu.Unlock()
	q.cond.Signal()
}

// Pop blocks until an item is available, the queue is closed or ctx ends.
func (q *taskQueue) Pop(ctx context.Context) (int, bool) {
	stop := context.AfterFunc(ctx, func() {
		q.mu.Lock()
		q.mu.Unlock()
		q.cond.Broadcast()
	})
	defer stop()

	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == 0 && !q.closed && ctx.Err() == nil {
		q.cond.Wait()
	}
	if len(q.items) == 0 {
		return 0, false
	}
	id := q.items[0]
	q.items = q.items[1:]
	return id, true
}

func (q *taskQueue) Close() {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()
	q.cond.Broadcast()
}

// --- yt-dlp output parsing ------------------------------------------------

// progressTemplate makes yt-dlp print one machine readable line per tick.
const progressTemplate = "download:MVD|%(progress.downloaded_bytes)s|%(progress.total_bytes)s|%(progress.total_bytes_estimate)s|%(progress.speed)s|%(progress.eta)s"

// parseProgressLine decodes a line produced by progressTemplate.
func parseProgressLine(line string) (EvProgress, bool) {
	if !strings.HasPrefix(line, "MVD|") {
		return EvProgress{}, false
	}
	parts := strings.Split(line, "|")
	if len(parts) != 6 {
		return EvProgress{}, false
	}
	num := func(s string) float64 {
		s = strings.TrimSpace(s)
		if s == "" || s == "NA" || s == "None" {
			return -1
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return -1
		}
		return f
	}
	downloaded := num(parts[1])
	total := num(parts[2])
	if total <= 0 {
		total = num(parts[3])
	}
	speed := num(parts[4])
	eta := num(parts[5])

	p := EvProgress{ETA: -1}
	if downloaded >= 0 {
		p.Downloaded = int64(downloaded)
	}
	if total > 0 {
		p.Total = int64(total)
		p.Percent = float64(p.Downloaded) / total * 100
		if p.Percent > 100 {
			p.Percent = 100
		}
	}
	if speed > 0 {
		p.Speed = speed
	}
	if eta >= 0 {
		p.ETA = int(eta)
	}
	return p, true
}

// isPostProcessLine reports whether a yt-dlp line marks the start of
// post-processing (merging, conversion, metadata).
func isPostProcessLine(line string) bool {
	for _, p := range []string{"[Merger]", "[ExtractAudio]", "[VideoConvertor]", "[VideoRemuxer]", "[Metadata]", "[EmbedThumbnail]"} {
		if strings.HasPrefix(line, p) {
			return true
		}
	}
	return false
}

// listPlaylistEntriesCtx is listPlaylistEntries with cancellation.
func listPlaylistEntriesCtx(ctx context.Context, ytDlpPath, playlistURL string, extraArgs []string) ([]PlaylistEntry, error) {
	args := []string{"--flat-playlist", "-j", "--no-warnings"}
	args = append(args, extraArgs...)
	args = append(args, playlistURL)

	cmd := exec.CommandContext(ctx, ytDlpPath, args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("yt-dlp --flat-playlist failed: %s", lastLine(msg))
	}
	return parsePlaylistEntries(strings.NewReader(string(out)))
}

func lastLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.LastIndex(s, "\n"); i >= 0 {
		return strings.TrimSpace(s[i+1:])
	}
	return s
}

// --- log file -------------------------------------------------------------

// runLogger appends every line to a log file so nothing is lost when the
// TUI redraws the screen.
type runLogger struct {
	path string
	mu   sync.Mutex
	f    *os.File
}

func newRunLogger(path string) (*runLogger, error) {
	l := &runLogger{path: path}
	if path == "" {
		return l, nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot open log file %s: %w", path, err)
	}
	l.f = f
	fmt.Fprintf(f, "\n===== MVD run started %s =====\n", time.Now().Format(time.RFC3339))
	return l, nil
}

func (l *runLogger) Write(playlist, entry int, line string) {
	if l.f == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	tag := fmt.Sprintf("P%d", playlist+1)
	if entry >= 0 {
		tag = fmt.Sprintf("P%d/E%d", playlist+1, entry)
	}
	fmt.Fprintf(l.f, "%s [%s] %s\n", time.Now().Format("15:04:05"), tag, line)
}

func (l *runLogger) Close() {
	if l.f != nil {
		l.f.Close()
	}
}
