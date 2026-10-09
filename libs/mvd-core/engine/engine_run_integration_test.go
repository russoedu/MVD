package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"youtube-downloader/libs/mvd-core/ytdlp"
)

// TestMain lets the test binary double as a yt-dlp stub, so engine tests
// run the same on every platform without shell scripts.
func TestMain(m *testing.M) {
	if os.Getenv("MVD_STUB_YTDLP") == "1" {
		os.Exit(stubYtDlp(os.Args[1:]))
	}
	os.Exit(m.Run())
}

func stubYtDlp(args []string) int {
	isFlat := false
	for _, a := range args {
		if a == "--flat-playlist" {
			isFlat = true
		}
	}
	url := args[len(args)-1]

	if isFlat {
		var entries []ytdlp.PlaylistEntry
		switch {
		case strings.Contains(url, "list=A"):
			entries = []ytdlp.PlaylistEntry{
				{ID: "aaaaaaaaaa1", Title: "Track One (Radio Edit)", Channel: "Artist - Topic"},
				{ID: "aaaaaaaaaa2", Title: "Track One (Extended)", Channel: "Artist - Topic"},
				{ID: "failfailfai", Title: "Broken", Channel: "Someone"},
			}
			for i := range entries {
				entries[i].Playlist, entries[i].PlaylistTitle = "Playlist A", "Playlist A"
				entries[i].PlaylistID, entries[i].PlaylistIndex, entries[i].PlaylistCount = "A", i+1, 3
			}
		case strings.Contains(url, "list=B"):
			entries = []ytdlp.PlaylistEntry{{ID: "bbbbbbbbbb1", Title: "Normal upload", Channel: "Band",
				Playlist: "Playlist B", PlaylistTitle: "Playlist B", PlaylistID: "B", PlaylistIndex: 1, PlaylistCount: 1}}
		case strings.Contains(url, "list=C"):
			entries = []ytdlp.PlaylistEntry{{ID: "cccccccccc1", Title: "Art track with a dead official video", Channel: "Other - Topic",
				Playlist: "Playlist C", PlaylistTitle: "Playlist C", PlaylistID: "C", PlaylistIndex: 1, PlaylistCount: 1}}
		case strings.Contains(url, "list=R"):
			entries = []ytdlp.PlaylistEntry{
				{ID: "trans4290001", Title: "Rate limited then ok", Channel: "Band"},
				{ID: "glitch000001", Title: "Glitch then ok", Channel: "Band"},
				{ID: "private00001", Title: "Private", Channel: "Band"},
			}
			for i := range entries {
				entries[i].Playlist, entries[i].PlaylistTitle = "Playlist R", "Playlist R"
				entries[i].PlaylistID, entries[i].PlaylistIndex, entries[i].PlaylistCount = "R", i+1, 3
			}
		default:
			fmt.Fprintln(os.Stderr, "ERROR: [youtube:tab] Unable to recognize playlist")
			return 1
		}
		for _, e := range entries {
			b, _ := json.Marshal(e)
			fmt.Println(string(b))
		}
		return 0
	}

	id := url[strings.LastIndex(url, "=")+1:]
	fmt.Printf("[youtube] %s: Downloading webpage\n", id)
	attempt := stubAttempt(id)
	switch {
	case strings.HasPrefix(id, "trans429"):
		if attempt == 1 {
			fmt.Fprintln(os.Stderr, "ERROR: [youtube] "+id+": HTTP Error 429: Too Many Requests")
			return 1
		}
	case strings.HasPrefix(id, "glitch"):
		if attempt == 1 {
			fmt.Fprintln(os.Stderr, "ERROR: [youtube] "+id+": Unable to download webpage")
			return 1
		}
	case strings.HasPrefix(id, "private"):
		fmt.Fprintln(os.Stderr, "ERROR: [youtube] "+id+": Private video. Sign in if you've been granted access")
		return 1
	case strings.Contains(id, "fail"):
		fmt.Fprintln(os.Stderr, "ERROR: [youtube] "+id+": Video unavailable")
		return 1
	}
	fmt.Println("MVD|512|1024|NA|2048.0|1")
	fmt.Println("MVD|1024|1024|NA|2048.0|0")
	fmt.Printf("[Merger] Merging formats into \"%s.mp4\"\n", id)
	return 0
}

// stubAttempt returns the 1-based attempt number for a video, persisted in
// MVD_STUB_STATE_DIR so a stub can "fail once then succeed". Without the dir
// it always reports attempt 1 (stateless), so other tests are unaffected.
func stubAttempt(id string) int {
	dir := os.Getenv("MVD_STUB_STATE_DIR")
	if dir == "" {
		return 1
	}
	p := filepath.Join(dir, id)
	n := 0
	if b, err := os.ReadFile(p); err == nil {
		n, _ = strconv.Atoi(strings.TrimSpace(string(b)))
	}
	n++
	_ = os.WriteFile(p, []byte(strconv.Itoa(n)), 0644)
	return n
}

// fakeResolver maps both art tracks to the same official video, and the normal
// upload to a better quality one.
type fakeResolver struct{}

func (fakeResolver) Wanted(title, channel, uploader string) bool {
	return strings.HasSuffix(channel, " - Topic") || title == "Normal upload"
}

func (fakeResolver) ResolveVersion(videoID, title, channel string, _ int, logf func(string, ...interface{})) Resolution {
	logf("[official] %s: looked up", videoID)
	if strings.HasPrefix(videoID, "aaaa") {
		return Resolution{VideoID: "OFFICIAL001", Official: true, Reason: `official video by "Label Records"`}
	}
	if strings.HasPrefix(videoID, "bbbb") {
		return Resolution{VideoID: "BETTER00001", Reason: "no official video found; the best quality is BETTER00001"}
	}
	if strings.HasPrefix(videoID, "cccc") {
		return Resolution{VideoID: "OFFICIALfail", Official: true, Reason: `official video by "Gone Records"`}
	}
	return Resolution{Reason: "no official video link found in description"}
}

func stubEngine(t *testing.T, official bool, urls ...string) *Engine {
	t.Helper()
	t.Setenv("MVD_STUB_YTDLP", "1")

	opts := Options{
		YtDlp:             os.Args[0],
		URLs:              urls,
		OutputDir:         t.TempDir(),
		OutputTemplate:    "%(playlist_title)s/%(playlist_index)02d - %(title)s.%(ext)s",
		Quality:           "best",
		MergeOutputFormat: "mp4",
		Workers:           2,
	}
	if official {
		opts.Resolver = fakeResolver{}
	}
	eng, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}
	return eng
}

// collect gathers events until EvIdle arrives (or timeout).
func collect(t *testing.T, eng *Engine) []interface{} {
	t.Helper()
	var seen []interface{}
	timeout := time.After(20 * time.Second)
	for {
		select {
		case ev, ok := <-eng.Events():
			if !ok {
				t.Fatal("events closed before idle")
			}
			seen = append(seen, ev)
			if _, idle := ev.(EvIdle); idle {
				return seen
			}
		case <-timeout:
			t.Fatal("timed out waiting for EvIdle")
		}
	}
}

// final returns the last state event seen per entry.
func final(events []interface{}) map[int]EvEntryState {
	out := map[int]EvEntryState{}
	for _, ev := range events {
		if st, ok := ev.(EvEntryState); ok {
			out[st.Entry] = st
		}
	}
	return out
}

func TestEngineRun(t *testing.T) {
	eng := stubEngine(t, true, "https://youtube.com/playlist?list=A", "https://youtube.com/playlist?list=B", "https://youtube.com/playlist?list=NOPE", "https://youtube.com/playlist?list=C")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { eng.Run(ctx); close(done) }()

	events := collect(t, eng)

	var listed, failedLists int
	infos := map[string]EntryInfo{}
	sawMerging, sawProgress := false, false
	for _, ev := range events {
		switch e := ev.(type) {
		case EvPlaylistListed:
			listed++
			for _, i := range e.Entries {
				infos[i.VideoID] = i
			}
			if e.Playlist == 0 && e.Title != "Playlist A" {
				t.Errorf("playlist title not applied: %q", e.Title)
			}
		case EvPlaylistFailed:
			failedLists++
		case EvEntryState:
			if e.State == StateMerging {
				sawMerging = true
			}
		case EvProgress:
			sawProgress = true
		}
	}
	if listed != 3 || failedLists != 1 {
		t.Fatalf("want 3 listed and 1 failed playlist, got %d and %d", listed, failedLists)
	}
	if !sawMerging || !sawProgress {
		t.Errorf("expected merging and progress events (merging=%v progress=%v)", sawMerging, sawProgress)
	}

	states := final(events)
	a1, a2 := states[infos["aaaaaaaaaa1"].ID], states[infos["aaaaaaaaaa2"].ID]
	// Both art tracks resolve to the same official video. Two workers race
	// for it, so either may win; the other must be a duplicate.
	if a2.State == StateDone {
		a1, a2 = a2, a1
	}
	if a1.State != StateDone || !a1.Official || a1.TargetID != "OFFICIAL001" {
		t.Errorf("one art track should be done via the official video: %+v", a1)
	}
	if a2.State != StateDuplicate || a2.TargetID != "OFFICIAL001" {
		t.Errorf("the other art track should be a duplicate: %+v", a2)
	}
	bad := states[infos["failfailfai"].ID]
	if bad.State != StateFailed || !strings.Contains(bad.Err, "Video unavailable") {
		t.Errorf("bad entry should fail with yt-dlp's error: %+v", bad)
	}
	b1 := states[infos["bbbbbbbbbb1"].ID]
	if b1.State != StateDone || b1.Official || !b1.Better || b1.TargetID != "BETTER00001" {
		t.Errorf("b1 should be replaced by the better quality upload, not counted as official: %+v", b1)
	}
	// The official video of c1 cannot be downloaded: the original is used.
	c1 := states[infos["cccccccccc1"].ID]
	if c1.State != StateDone || c1.Official || c1.TargetID != "cccccccccc1" {
		t.Errorf("c1 should fall back to the original track: %+v", c1)
	}
	sawFallbackLog := false
	for _, ev := range events {
		if lg, ok := ev.(EvLog); ok && lg.Entry == c1.Entry && strings.Contains(lg.Line, "downloading the original instead") {
			sawFallbackLog = true
		}
	}
	if !sawFallbackLog {
		t.Error("fallback to the original should be logged")
	}

	// Retry the failed entry: it fails again and the engine goes idle again.
	if !eng.Retry(bad.Entry) {
		t.Fatal("retry of a failed entry should be accepted")
	}
	if eng.Retry(b1.Entry) {
		t.Fatal("retry of a finished entry must be refused")
	}
	if st := final(collect(t, eng))[bad.Entry]; st.State != StateFailed {
		t.Errorf("retried entry should fail again, got %v", st.State)
	}
	if n := eng.RetryPlaylist(0); n != 1 {
		t.Errorf("RetryPlaylist should re-queue 1 entry, got %d", n)
	}
	collect(t, eng)

	cancel()
	for range eng.Events() {
	}
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("engine did not stop after cancel")
	}
}

func TestEngineWithoutResolver(t *testing.T) {
	eng := stubEngine(t, false, "https://youtube.com/playlist?list=A")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { eng.Run(ctx); close(done) }()
	states := final(collect(t, eng))
	done_, dup, failed := 0, 0, 0
	for _, st := range states {
		switch st.State {
		case StateDone:
			done_++
			if st.Official {
				t.Error("nothing should be official without a resolver")
			}
		case StateDuplicate:
			dup++
		case StateFailed:
			failed++
		}
	}
	if done_ != 2 || dup != 0 || failed != 1 {
		t.Errorf("want 2 done, 0 dup, 1 failed; got %d %d %d", done_, dup, failed)
	}
	cancel()
	for range eng.Events() {
	}
	<-done
}

func retryEngine(t *testing.T, autoRetry bool, stateDir string) *Engine {
	t.Helper()
	t.Setenv("MVD_STUB_YTDLP", "1")
	if stateDir != "" {
		t.Setenv("MVD_STUB_STATE_DIR", stateDir)
	}
	eng, err := New(Options{
		YtDlp:             os.Args[0],
		URLs:              []string{"https://youtube.com/playlist?list=R"},
		OutputDir:         t.TempDir(),
		OutputTemplate:    "%(title)s.%(ext)s",
		Quality:           "best",
		MergeOutputFormat: "mp4",
		Workers:           2,
		AutoRetry:         autoRetry,
		RetryCooldown:     10 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	return eng
}

// videoStates maps video id -> its last state, from a run's events.
func videoStates(events []interface{}) map[string]EvEntryState {
	states := final(events)
	out := map[string]EvEntryState{}
	for _, ev := range events {
		if pl, ok := ev.(EvPlaylistListed); ok {
			for _, i := range pl.Entries {
				out[i.VideoID] = states[i.ID]
			}
		}
	}
	return out
}

func TestEngineAutoRetry(t *testing.T) {
	eng := retryEngine(t, true, t.TempDir())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { eng.Run(ctx); close(done) }()

	events := collect(t, eng)
	byVideo := videoStates(events)

	if st := byVideo["trans4290001"]; st.State != StateDone {
		t.Errorf("rate-limited entry should recover in the sweep: %+v", st)
	}
	if st := byVideo["glitch000001"]; st.State != StateDone {
		t.Errorf("glitch entry should recover on the immediate retry: %+v", st)
	}
	if st := byVideo["private00001"]; st.State != StateFailed {
		t.Errorf("private entry must not be retried: %+v", st)
	}

	sawCooldown, sawImmediate := false, false
	for _, ev := range events {
		if lg, ok := ev.(EvLog); ok {
			if strings.Contains(lg.Line, "auto-retry after cooldown") {
				sawCooldown = true
			}
			if strings.Contains(lg.Line, "retrying now") {
				sawImmediate = true
			}
		}
	}
	if !sawCooldown || !sawImmediate {
		t.Errorf("expected both retry kinds logged (cooldown=%v immediate=%v)", sawCooldown, sawImmediate)
	}

	cancel()
	for range eng.Events() {
	}
	<-done
}

func TestEngineAutoRetryOff(t *testing.T) {
	eng := retryEngine(t, false, t.TempDir())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { eng.Run(ctx); close(done) }()

	byVideo := videoStates(collect(t, eng))
	for _, v := range []string{"trans4290001", "glitch000001", "private00001"} {
		if st := byVideo[v]; st.State != StateFailed {
			t.Errorf("with auto-retry off %s should fail, got %+v", v, st)
		}
	}

	cancel()
	for range eng.Events() {
	}
	<-done
}

// stop cancels a running engine, drains its events and waits for Run to return.
func stop(eng *Engine, cancel context.CancelFunc, done <-chan struct{}) {
	cancel()
	for range eng.Events() {
	}
	<-done
}

// startEngine runs the engine in the background.
func startEngine(eng *Engine) (context.CancelFunc, <-chan struct{}) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { eng.Run(ctx); close(done) }()
	return cancel, done
}

func indexOfEvent(events []interface{}, match func(interface{}) bool) int {
	for i, ev := range events {
		if match(ev) {
			return i
		}
	}
	return -1
}

func TestEngineAddSourceWhileRunning(t *testing.T) {
	eng := stubEngine(t, false, "https://youtube.com/playlist?list=B")
	cancel, done := startEngine(eng)

	first := collect(t, eng) // the first playlist is finished and the engine is idle
	if len(final(first)) != 1 {
		t.Fatalf("want 1 entry before adding, got %d", len(final(first)))
	}

	src, ok := eng.AddSource("https://youtube.com/playlist?list=A")
	if !ok || src.Index != 1 {
		t.Fatalf("AddSource = %+v, %v; want index 1", src, ok)
	}

	second := collect(t, eng) // must not end before the new playlist is listed and done

	added := indexOfEvent(second, func(ev interface{}) bool {
		a, is := ev.(EvPlaylistAdded)
		return is && a.Source.Index == 1
	})
	listing := indexOfEvent(second, func(ev interface{}) bool {
		l, is := ev.(EvPlaylistListing)
		return is && l.Playlist == 1
	})
	listed := indexOfEvent(second, func(ev interface{}) bool {
		l, is := ev.(EvPlaylistListed)
		return is && l.Playlist == 1
	})
	if added < 0 || listing < 0 || listed < 0 || added >= listing || listing >= listed {
		t.Fatalf("want added < listing < listed for playlist 1, got %d %d %d", added, listing, listed)
	}

	states := final(append(first, second...))
	if len(states) != 4 {
		t.Fatalf("want 4 entries in all (1 + 3), got %d", len(states))
	}
	for id, st := range states {
		if st.State != StateDone && st.State != StateFailed {
			t.Errorf("entry %d is still %v when the engine reported idle", id, st.State)
		}
	}
	if got := len(eng.Sources()); got != 2 {
		t.Errorf("Sources() = %d, want 2", got)
	}

	stop(eng, cancel, done)
}

// The invariant that keeps a source added just as the last download finishes from
// producing a false "idle": while one waits to be listed, a drain must not be
// signalled. It cannot be hit reliably through a running engine, since it is a
// window between two lock acquisitions, so it is pinned directly.
func TestEngineAddSourceKeepsTheEngineBusyUntilListed(t *testing.T) {
	eng := stubEngine(t, false)

	eng.signalDrain()
	select {
	case <-eng.drained:
	default:
		t.Fatal("with nothing queued and nothing to list, a drain should be signalled")
	}

	if _, ok := eng.AddSource("https://youtube.com/playlist?list=B"); !ok {
		t.Fatal("AddSource refused a source on an engine that has not run")
	}
	eng.signalDrain()
	select {
	case <-eng.drained:
		t.Fatal("a drain was signalled while a source was waiting to be listed")
	default:
	}

	// Once the producer has taken it and found nothing more, drains are signalled again.
	if _, ok := eng.nextToList(); !ok {
		t.Fatal("nextToList returned nothing for the waiting source")
	}
	if _, ok := eng.nextToList(); ok {
		t.Fatal("nextToList returned a source twice")
	}
	eng.signalDrain()
	select {
	case <-eng.drained:
	default:
		t.Fatal("a drain should be signalled once everything is listed")
	}
}

func TestEngineAddSourceBeforeRun(t *testing.T) {
	eng := stubEngine(t, false)
	src, ok := eng.AddSource("https://youtube.com/playlist?list=B")
	if !ok || src.Index != 0 {
		t.Fatalf("AddSource = %+v, %v; want index 0", src, ok)
	}

	cancel, done := startEngine(eng)
	events := collect(t, eng)

	if indexOfEvent(events, func(ev interface{}) bool { _, is := ev.(EvPlaylistAdded); return is }) < 0 {
		t.Error("no EvPlaylistAdded for a source added before Run")
	}
	states := final(events)
	if len(states) != 1 || states[0].State != StateDone {
		t.Errorf("want the one entry done, got %+v", states)
	}
	if got := len(eng.Sources()); got != 1 {
		t.Errorf("Sources() = %d, want 1 (listed once, not twice)", got)
	}
	stop(eng, cancel, done)
}

func TestEngineAddSourceAfterStopIsRefused(t *testing.T) {
	eng := stubEngine(t, false, "https://youtube.com/playlist?list=B")
	cancel, done := startEngine(eng)
	collect(t, eng)
	stop(eng, cancel, done)

	if _, ok := eng.AddSource("https://youtube.com/playlist?list=A"); ok {
		t.Error("AddSource accepted a source after Run finished")
	}
	if got := len(eng.Sources()); got != 1 {
		t.Errorf("Sources() = %d, want 1", got)
	}
}

func TestEngineAddSourceConcurrently(t *testing.T) {
	eng := stubEngine(t, false) // nothing to list at first: the engine goes idle at once
	cancel, done := startEngine(eng)
	collect(t, eng)

	const n = 8
	indices := make(chan int, n)
	for i := 0; i < n; i++ {
		go func() {
			src, ok := eng.AddSource("https://youtube.com/playlist?list=NOPE")
			if !ok {
				src.Index = -1
			}
			indices <- src.Index
		}()
	}
	seen := map[int]bool{}
	for i := 0; i < n; i++ {
		seen[<-indices] = true
	}
	for i := 0; i < n; i++ {
		if !seen[i] {
			t.Errorf("index %d was never handed out: %v", i, seen)
		}
	}

	// Every playlist is listed (and fails, as NOPE is not one) before the idle event.
	events := collect(t, eng)
	failed := 0
	for _, ev := range events {
		if _, is := ev.(EvPlaylistFailed); is {
			failed++
		}
	}
	if failed != n {
		t.Errorf("want %d playlists listed before idle, got %d", n, failed)
	}
	stop(eng, cancel, done)
}

func TestTaskQueue(t *testing.T) {
	q := newTaskQueue()
	q.Push(1)
	q.Push(2)
	if id, ok := q.Pop(context.Background()); !ok || id != 1 {
		t.Fatalf("want 1, got %d %v", id, ok)
	}
	if id, ok := q.Pop(context.Background()); !ok || id != 2 {
		t.Fatalf("want 2, got %d %v", id, ok)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, ok := q.Pop(ctx); ok {
		t.Fatal("Pop on empty queue should return false once ctx ends")
	}

	q.Close()
	if _, ok := q.Pop(context.Background()); ok {
		t.Fatal("Pop on closed queue should return false")
	}
}

func TestRunLogger(t *testing.T) {
	path := t.TempDir() + "/mvd.log"
	l, err := newRunLogger(path)
	if err != nil {
		t.Fatal(err)
	}
	l.Write(0, -1, "listing")
	l.Write(1, 7, "line")
	l.Close()
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "[P1] listing") || !strings.Contains(string(data), "[P2/E7] line") {
		t.Errorf("unexpected log:\n%s", data)
	}
	if l, err := newRunLogger(""); err != nil || l.path != "" {
		t.Error("empty path should disable the log")
	}
}
