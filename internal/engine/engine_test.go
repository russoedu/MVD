package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"youtube-downloader/internal/ytdlp"
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
	if strings.Contains(id, "fail") {
		fmt.Fprintln(os.Stderr, "ERROR: [youtube] "+id+": Video unavailable")
		return 1
	}
	fmt.Println("MVD|512|1024|NA|2048.0|1")
	fmt.Println("MVD|1024|1024|NA|2048.0|0")
	fmt.Printf("[Merger] Merging formats into \"%s.mp4\"\n", id)
	return 0
}

// fakeResolver maps both art tracks to the same official video.
type fakeResolver struct{}

func (fakeResolver) Wanted(channel, uploader string) bool {
	return strings.HasSuffix(channel, " - Topic")
}

func (fakeResolver) ResolveLog(videoID string, logf func(string, ...interface{})) (string, string) {
	logf("[official] %s: looked up", videoID)
	if strings.HasPrefix(videoID, "aaaa") {
		return "OFFICIAL001", `official video by "Label Records"`
	}
	return "", "no official video link found in description"
}

func stubEngine(t *testing.T, official bool, urls ...string) *Engine {
	t.Helper()
	os.Setenv("MVD_STUB_YTDLP", "1")
	t.Cleanup(func() { os.Unsetenv("MVD_STUB_YTDLP") })

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
	eng := stubEngine(t, true, "https://youtube.com/playlist?list=A", "https://youtube.com/playlist?list=B", "https://youtube.com/playlist?list=NOPE")
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
	if listed != 2 || failedLists != 1 {
		t.Fatalf("want 2 listed and 1 failed playlist, got %d and %d", listed, failedLists)
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
	if b1.State != StateDone || b1.Official || b1.TargetID != "bbbbbbbbbb1" {
		t.Errorf("b1 should be downloaded as is: %+v", b1)
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
