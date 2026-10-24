package engine

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
	"youtube-downloader/libs/mvd-core/ytdlp/fetch"

	"youtube-downloader/libs/mvd-core/ytdlp"
)

// defaultRetryWait is the wait before the first retry cycle; each further cycle waits
// that long again, once more (the second twice as long, the third three times).
const defaultRetryWait = 30 * time.Second

// maxDownloaders is the most downloads that can ever be in flight: the pool is this big,
// and the throttle holds it to the number the person asks for, which can change while
// the run goes on.
const maxDownloaders = 64

// defaultRetryCycles is how many retry cycles a failed entry gets.
const defaultRetryCycles = 30

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
	// OfficialOnly downloads only the songs that have an official video; the others end
	// as StateNotFound. NotFoundFile, when set, is a CSV file that receives them (the
	// playlist, the artist, the title and the address), rewritten each time the engine
	// goes quiet.
	OfficialOnly bool
	NotFoundFile string
	// PlanOnly runs the listing, naming and picking stages and stops there: each song
	// ends as StatePlanned (or not found, or failed) and Plan() says what would be
	// downloaded. Nothing is downloaded.
	PlanOnly bool
	// OfficialLookupOff builds the run with the lookup of official videos switched off; it
	// can be switched on later (see Reconfigure). Without a Resolver there is no lookup.
	OfficialLookupOff bool
	// HoldForReview runs the listing, naming and picking stages and keeps going: each song
	// waits as StatePlanned until Release sends it to the download stage, so a song
	// the person has looked at downloads while the others are still being looked up or
	// looked at. Unlike PlanOnly the run goes on, ready for Release.
	HoldForReview bool
	// Plan, when set, is what to download: the playlists are not listed and no version is
	// looked for, each entry goes straight to the download stage with the video it names.
	Plan []PlannedEntry
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
	// CheckDownload, when set, is called with the path of each file a replacement video
	// produced. A problem it returns (the sound plays on one side only, say) sends the
	// song to the search again, without that video; an error only means it could not be
	// checked. The file itself is never changed.
	CheckDownload func(ctx context.Context, path string) (problem string, err error)
	Workers       int    // the most videos in flight across all playlists; fewer while the service limits requests
	LogPath       string // "" disables the log file
	Resolver      OfficialResolver
	// Tracks, when set, lists the playlists of another service and finds the
	// YouTube video of each track.
	Tracks TrackSource
	// AutoRetry turns on the immediate one-off retry and the retry cycles.
	AutoRetry bool
	// RetryCycles is how many times a failed download is retried once the backlog has
	// drained. 0 uses 30.
	RetryCycles int
	// RetryWait is the wait before the first cycle, multiplied by the number of the cycle:
	// with 10 seconds the first retry waits 10, the second 20, and so on. 0 uses 30 seconds.
	RetryWait time.Duration
	// StartTimeout is how long a download may take to begin fetching data before it is
	// stopped and counted as a failure the retry cycles take up. 0 waits as long as it takes.
	StartTimeout time.Duration
}

type engineEntry struct {
	info     EntryInfo
	raw      ytdlp.PlaylistEntry
	state    EntryState
	targetID string
	// identified is what the naming stage learned, for the picking stage.
	identified *Identification
	// ownOfficial is true when the upload itself is the official video.
	ownOfficial bool
	// named is the song as a music database named it.
	named    *namedSong
	official bool
	better   bool   // replaced by an upload of better quality, not the official video
	reason   string // why the version was chosen, as the lookup put it
	// fixed is true for an entry that came from a plan: the version is not looked for again.
	fixed bool
	// rejected are the videos that were downloaded and found wanting, never taken again.
	rejected []string
	// userChose is true when the person picked the video themselves: it is kept as it is.
	userChose bool
	track     *Track // set when the entry is a song with no video yet
	err       string
	deferred  bool // failed with an error a later retry may fix, eligible for a cycle
	sweeps    int  // retry cycles already spent
	// saved are the files the last successful download wrote, so a swap can take them away.
	saved []string
}

// namedSong is a song as a music database named it.
type namedSong struct{ artist, title, album string }

// notFoundRow is a song with no official video, as the CSV file lists it.
type notFoundRow struct{ playlist, artist, title, url string }

// Engine lists playlists, resolves official videos and runs yt-dlp per
// entry with a bounded pool of workers. All observable state is reported
// through Events().
type Engine struct {
	opts    Options
	// throttle keeps the downloads at or below Workers, fewer while the service limits requests.
	throttle *downloadThrottle
	// live holds the options that can change while the run goes on (see Reconfigure).
	liveMu sync.RWMutex
	live   LiveSettings
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

	removed        map[int]bool   // playlists the person took out of the run
	playlistTitles map[int]string // the titles of the playlists listed, by index
	notFound       []notFoundRow  // songs with no official video, when only those are wanted
}

// New builds an engine. It opens the log file right away.
func New(opts Options) (*Engine, error) {
	logger, err := newRunLogger(opts.LogPath)
	if err != nil {
		return nil, err
	}
	e := &Engine{
		opts:           opts,
		events:         make(chan interface{}, 1024),
		nameQueue:      newTaskQueue(),
		pickQueue:      newTaskQueue(),
		queue:          newTaskQueue(),
		logger:         logger,
		drained:        make(chan struct{}, 1),
		wake:           make(chan struct{}, 1),
		claims:         make(map[int]map[string]int),
		playlistTitles: make(map[int]string),
		removed:        make(map[int]bool),
		throttle:       newDownloadThrottle(opts.Workers),
		live:           liveOf(opts),
	}
	for i, u := range opts.URLs {
		e.sources = append(e.sources, PlaylistSource{Index: i, URL: u})
		e.toList = append(e.toList, i)
	}
	e.planSources()
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
	downloading := pool(maxDownloaders, e.queue, func(ctx context.Context, id int) {
		e.process(ctx, id)
		e.finishOne()
	})

	e.seedPlan()

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
	e.writeNotFound()
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

// coordinate waits for the queue to drain and then either runs a retry cycle (after a
// wait that grows with the cycle) or emits the final idle event.
func (e *Engine) coordinate(ctx context.Context) {
	wait := e.settings().RetryWait
	if wait <= 0 {
		wait = defaultRetryWait
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-e.drained:
			ids, cycle := e.sweepCandidates()
			if len(ids) == 0 {
				e.markIdle()
				continue
			}
			e.mu.Lock()
			pl := e.entries[ids[0]].info.Playlist
			e.mu.Unlock()
			cooldown := wait * time.Duration(cycle)
			e.log(pl, -1, "retry %d of %d: waiting %s before retrying %d failed download(s)", cycle, e.retryCycles(), cooldown, len(ids))
			if !sleepCtx(ctx, cooldown) {
				return
			}
			for _, id := range ids {
				e.autoRetry(id)
			}
		}
	}
}

// retryCycles is how many retry cycles an entry gets.
func (e *Engine) retryCycles() int { return orDefault(e.settings().RetryCycles, defaultRetryCycles) }

// Retry re-queues a failed entry. It returns false when the entry is not
// in a failed state.
func (e *Engine) Retry(entryID int) bool {
	e.mu.Lock()
	if entryID < 0 || entryID >= len(e.entries) {
		e.mu.Unlock()
		return false
	}
	en := e.entries[entryID]
	if en.state != StateFailed || e.removed[en.info.Playlist] {
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
	if e.isRemoved(src.Index) {
		e.log(src.Index, -1, "removed while it was being listed")
		return
	}
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
	e.playlistTitles[src.Index] = title
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
	ev := EvEntryState{Entry: en.info.ID, State: st, TargetID: en.targetID, Official: en.official, Better: en.better, Err: err, OwnOfficial: en.ownOfficial, Reason: en.reason}
	if en.named != nil {
		ev.Artist, ev.Title, ev.Album = en.named.artist, en.named.title, en.named.album
	}
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
	if e.removed[en.info.Playlist] {
		e.mu.Unlock()
		e.dropRemoved(en)
		return
	}
	if en.fixed {
		// The version came from a plan, and stays whatever the lookup would say now.
		e.mu.Unlock()
		e.queue.Push(id)
		return
	}
	en.targetID = en.info.VideoID
	en.official = false
	en.better = false
	en.identified = nil
	en.ownOfficial = false
	en.named = nil
	e.mu.Unlock()

	pl, eid := en.info.Playlist, en.info.ID
	if ctx.Err() != nil {
		e.cancelEntry(en)
		return
	}

	switch {
	case en.track != nil:
		e.pickQueue.Push(id) // a song of another service is found by its own search
	case e.resolver() != nil && e.resolver().Wanted(en.info.Title, en.raw.Channel, en.raw.Uploader):
		e.setState(en, StateResolving, "")
		channel := en.raw.Channel
		if channel == "" {
			channel = en.raw.Uploader
		}
		identification := e.resolver().Identify(en.info.VideoID, en.raw.Title, channel, int(en.raw.Duration), e.entryLog(pl, eid))
		if identification.Done {
			e.applyResolution(en, identification.Resolution)
			e.advance(en)
			return
		}
		e.mu.Lock()
		en.identified = &identification
		if identification.Artist != "" && identification.Title != "" {
			en.named = &namedSong{artist: identification.Artist, title: identification.Title, album: identification.Album}
		}
		e.mu.Unlock()
		e.pickQueue.Push(id)
	default:
		// Nothing to look up. With a resolver that means the upload says it is the
		// official video, which is then what is wanted.
		e.mu.Lock()
		en.ownOfficial = e.resolver() != nil
		e.mu.Unlock()
		e.advance(en)
	}
}

// pickStage is the second stage: it finds the version of the video to download, the
// official one or else the best quality.
func (e *Engine) pickStage(ctx context.Context, id int) {
	e.mu.Lock()
	en := e.entries[id]
	identification := en.identified
	gone := e.removed[en.info.Playlist]
	e.mu.Unlock()
	if gone {
		e.dropRemoved(en)
		return
	}

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
	case identification != nil && e.resolver() != nil:
		e.applyResolution(en, e.resolver().Pick(*identification, e.entryLog(pl, eid)))
	}
	e.advance(en)
}

// advance sends an entry to the download stage, or ends it when only official videos
// are wanted and this song has none.
func (e *Engine) advance(en *engineEntry) {
	e.mu.Lock()
	official := en.official || en.ownOfficial
	e.mu.Unlock()

	if e.settings().OfficialOnly && !official {
		e.markNotFound(en)
		e.finishOne()
		return
	}
	if e.opts.PlanOnly || e.opts.HoldForReview {
		e.setState(en, StatePlanned, "")
		e.finishOne()
		return
	}
	e.queue.Push(en.info.ID)
}

// markNotFound ends an entry that has no official video, and notes it for the CSV file.
func (e *Engine) markNotFound(en *engineEntry) {
	e.log(en.info.Playlist, en.info.ID, "no official video found: not downloaded")
	e.mu.Lock()
	row := notFoundRow{playlist: e.playlistTitles[en.info.Playlist]}
	switch {
	case en.named != nil:
		row.artist, row.title = en.named.artist, en.named.title
	case en.track != nil:
		row.artist, row.title = en.track.Artist, en.track.Title
	default:
		channel := en.raw.Channel
		if channel == "" {
			channel = en.raw.Uploader
		}
		row.artist, row.title = strings.TrimSuffix(channel, " - Topic"), en.raw.Title
	}
	if en.track == nil && en.info.VideoID != "" {
		row.url = "https://www.youtube.com/watch?v=" + en.info.VideoID
	}
	e.notFound = append(e.notFound, row)
	e.mu.Unlock()
	e.setState(en, StateNotFound, "")
}

// writeNotFound writes the songs with no official video to the CSV file, all of those
// of this run so far.
func (e *Engine) writeNotFound() {
	if e.opts.NotFoundFile == "" {
		return
	}
	e.mu.Lock()
	rows := append([]notFoundRow(nil), e.notFound...)
	e.mu.Unlock()
	if len(rows) == 0 {
		return
	}
	if err := os.MkdirAll(filepath.Dir(e.opts.NotFoundFile), 0o755); err != nil {
		return
	}
	file, err := os.Create(e.opts.NotFoundFile)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	// A byte order mark lets Excel read the accents of the titles.
	_, _ = file.WriteString("\ufeff")
	w := csv.NewWriter(file)
	_ = w.Write([]string{"playlist", "artist", "title", "url"})
	for _, row := range rows {
		_ = w.Write([]string{row.playlist, row.artist, row.title, row.url})
	}
	w.Flush()
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
		e.mu.Lock()
		en.reason = res.Reason
		if res.OwnOfficial {
			en.ownOfficial = true
		}
		e.mu.Unlock()
		e.log(pl, eid, "kept original (%s)", res.Reason)
		return
	}
	e.mu.Lock()
	en.reason = res.Reason
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

	if e.isRemoved(pl) {
		e.setState(en, StateSkipped, "")
		return
	}
	e.mu.Lock()
	rejected := en.state == StateSkipped
	e.mu.Unlock()
	if rejected {
		return // accepted, and then rejected before its turn came
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

	immediateUsed := false
	for {
		err := e.throttledDownload(ctx, en)
		if err == nil {
			e.setState(en, StateDone, "")
			return
		}
		if ctx.Err() != nil {
			e.setState(en, StateFailed, "cancelled")
			return
		}

		// The video came out with a defect: another one is looked for.
		var bad *rejectedDownload
		if errors.As(err, &bad) {
			e.searchAgain(en, bad)
			if !e.claimTarget(pl, en.targetID, eid) {
				e.log(pl, eid, "%s already downloaded for this playlist, skipping duplicate", en.targetID)
				e.setState(en, StateDuplicate, "")
				return
			}
			continue
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
			continue
		}

		if !e.settings().AutoRetry {
			e.setState(en, StateFailed, err.Error())
			return
		}

		switch classify(err.Error()) {
		case ClassImmediate:
			if !immediateUsed {
				immediateUsed = true
				e.log(pl, eid, "retrying now (%v)", err)
				continue
			}
			// An error nobody has classified may be a glitch that a later cycle clears.
			e.mu.Lock()
			en.deferred = true
			e.mu.Unlock()
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

// throttledDownload is download under the throttle: it waits for a downloader the
// current limit allows, and tells the throttle how the download went.
func (e *Engine) throttledDownload(ctx context.Context, en *engineEntry) error {
	release, ok := e.throttle.acquire(ctx)
	if !ok {
		return ctx.Err()
	}
	defer release()
	// The song shows as downloading only once a downloader is free, so a song waiting its
	// turn (or a pause) is still queued.
	e.mu.Lock()
	downloading := en.state == StateDownloading || en.state == StateMerging
	e.mu.Unlock()
	if !downloading {
		e.setState(en, StateDownloading, "")
	}

	err := e.download(ctx, en)
	var change string
	switch {
	case err == nil:
		change = e.throttle.succeeded()
	case isRateLimited(err.Error()):
		change = e.throttle.rateLimited()
	}
	if change != "" {
		e.log(en.info.Playlist, -1, "%s", change)
	}
	return err
}

// download runs yt-dlp for the entry's current target, turning its output
// into progress, merging and log events.
func (e *Engine) download(ctx context.Context, en *engineEntry) error {
	pl, eid := en.info.Playlist, en.info.ID

	e.mu.Lock()
	target := en.targetID
	e.mu.Unlock()

	live := e.settings()
	template := fetch.ApplyPlaylistFields(live.OutputTemplate, en.raw)
	options := fetch.DownloadOptions{
		Format:              live.Quality,
		OutputTemplate:      filepath.Join(live.OutputDir, template),
		MergeOutputFormat:   live.MergeOutputFormat,
		ConcurrentFragments: live.ConcurrentFragments,
		ExtraArgs:           live.ExtraArgs,
	}
	if e.opts.PartialDir != "" {
		options.OutputTemplate, options.HomeDir, options.TempDir = template, live.OutputDir, e.opts.PartialDir
	}
	trailing := []string{"--newline", "--progress-template", fetch.ProgressTemplate, "--no-playlist"}
	// yt-dlp writes the path of every finished file here, for CheckDownload. Only a
	// replacement is checked: the playlist's own upload is what there is to fall back to.
	// It is also how the files are known, so a swap of the video can take them away.
	check := e.opts.CheckDownload != nil && target != en.info.VideoID && en.info.VideoID != "" && !en.userChose
	pathsFile := ""
	if f, err := os.CreateTemp("", "mvd-paths-*.txt"); err == nil {
		pathsFile = f.Name()
		_ = f.Close()
		defer func() { _ = os.Remove(pathsFile) }()
		trailing = append(trailing, "--print-to-file", "after_move:filepath", pathsFile)
	}
	// yt-dlp says how good the chosen version is in a file of its own, as it begins: it is read at
	// the first progress tick, so what yt-dlp prints stays what it was.
	qualityFile := ""
	if f, err := os.CreateTemp("", "mvd-quality-*.txt"); err == nil {
		qualityFile = f.Name()
		_ = f.Close()
		defer func() { _ = os.Remove(qualityFile) }()
		trailing = append(trailing, "--print-to-file", fetch.QualityPrint, qualityFile)
	}
	args := fetch.DownloadArgs(options, append(trailing, "https://www.youtube.com/watch?v="+target)...)

	qualityKnown := false
	err := fetch.DownloadWithStartTimeout(ctx, e.opts.YtDlp, args, live.StartTimeout, func(line string) {
		if p, ok := fetch.ParseProgressLine(line); ok {
			if !qualityKnown && qualityFile != "" {
				if data, err := os.ReadFile(qualityFile); err == nil {
					if height, kbps, found := fetch.ParseQualityLine(string(data)); found {
						qualityKnown = true
						e.emit(EvQuality{Entry: eid, Height: height, AudioKbps: kbps})
					}
				}
			}
			e.emit(EvProgress{Entry: eid, Percent: p.Percent, Downloaded: p.Downloaded, Total: p.Total, Speed: p.Speed, ETA: p.ETA, Parts: partsInFlight(p, live.ConcurrentFragments)})
			return
		}
		if fetch.IsPostProcessLine(line) {
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
	if err == nil && pathsFile != "" {
		if data, readErr := os.ReadFile(pathsFile); readErr == nil {
			var files []string
			for _, path := range strings.Split(string(data), "\n") {
				if path = strings.TrimSpace(path); path != "" {
					files = append(files, path)
				}
			}
			e.mu.Lock()
			en.saved = files
			e.mu.Unlock()
		}
		if check {
			return e.checkDownloaded(ctx, pl, eid, target, pathsFile)
		}
	}
	return err
}

// rejectedDownload is the answer of download when the file came out but is not one to
// keep: the file is removed and another version is looked for.
type rejectedDownload struct {
	problem string
	files   []string
}

func (r *rejectedDownload) Error() string { return r.problem }

// checkDownloaded hands each file yt-dlp reported to Options.CheckDownload. A problem
// with any of them rejects the download.
func (e *Engine) checkDownloaded(ctx context.Context, pl, eid int, target, pathsFile string) error {
	data, err := os.ReadFile(pathsFile)
	if err != nil {
		return nil
	}
	var files []string
	problem := ""
	for _, path := range strings.Split(string(data), "\n") {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		files = append(files, path)
		found, err := e.opts.CheckDownload(ctx, path)
		switch {
		case err != nil:
			e.log(pl, eid, "[check] %s: %v", filepath.Base(path), err)
		case found != "" && problem == "":
			problem = found
		}
	}
	if problem == "" {
		return nil
	}
	return &rejectedDownload{problem: fmt.Sprintf("%s: %s", target, problem), files: files}
}

// searchAgain takes a rejected replacement out of the running: its files are removed,
// and the song gets another version (the next best one the search finds, else the
// playlist's own upload). It reports false when there is nothing left to try.
func (e *Engine) searchAgain(en *engineEntry, bad *rejectedDownload) {
	pl, eid := en.info.Playlist, en.info.ID
	for _, path := range bad.files {
		_ = os.Remove(path)
	}

	e.mu.Lock()
	en.rejected = append(en.rejected, en.targetID)
	rejected := append([]string(nil), en.rejected...)
	identified := en.identified
	e.mu.Unlock()
	e.log(pl, eid, "%s was removed (%s); looking for another version", en.targetID, bad.problem)

	var res Resolution
	if picker, ok := e.resolver().(RePicker); ok && identified != nil {
		res = picker.PickWithout(*identified, rejected, e.entryLog(pl, eid))
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	if res.VideoID != "" && !slices.Contains(en.rejected, res.VideoID) {
		en.targetID, en.official, en.better, en.reason = res.VideoID, res.Official, !res.Official, res.Reason
		return
	}
	en.targetID, en.official, en.better = en.info.VideoID, false, false
	en.reason = "the other versions of the song play on one side only; the playlist's own upload is kept"
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
	e.writeNotFound()
	if !already {
		e.emit(EvIdle{})
	}
}

// sweepCandidates returns the failed entries still owed a retry cycle, and the number of
// the cycle they are about to get: the lowest count among them plus one, so the wait
// grows with the attempts and starts over for entries that fail later.
func (e *Engine) sweepCandidates() ([]int, int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.settings().AutoRetry {
		return nil, 0
	}
	var ids []int
	cycle := 0
	for _, en := range e.entries {
		if en.deferred && en.state == StateFailed && en.sweeps < e.retryCycles() && !e.removed[en.info.Playlist] {
			ids = append(ids, en.info.ID)
			if cycle == 0 || en.sweeps+1 < cycle {
				cycle = en.sweeps + 1
			}
		}
	}
	return ids, cycle
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
	e.log(en.info.Playlist, id, "retrying after the wait")
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

// partsInFlight is how many parts of a file are being fetched at once: when the part
// comes in fragments, as many as are allowed at once and are left, else the one part.
func partsInFlight(p fetch.Progress, concurrentFragments int) int {
	if p.FragmentCount <= 1 || concurrentFragments <= 1 {
		return 1
	}
	left := p.FragmentCount - p.FragmentIndex + 1
	if left < 1 {
		left = 1
	}
	return min(concurrentFragments, left)
}
