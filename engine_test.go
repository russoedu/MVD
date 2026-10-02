package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestMain lets the test binary double as a yt-dlp stub, so engine tests
// run the same on every platform without shell scripts.
func TestMain(m *testing.M) {
	if os.Getenv("MVD_STUB_YTDLP") == "1" {
		os.Exit(stubYtDlp(os.Args[1:]))
	}
	os.Exit(m.Run())
}

// stubYtDlp mimics the two yt-dlp invocations the engine makes.
func stubYtDlp(args []string) int {
	isFlat := false
	for _, a := range args {
		if a == "--flat-playlist" {
			isFlat = true
		}
	}
	url := args[len(args)-1]

	if isFlat {
		var entries []PlaylistEntry
		switch {
		case strings.Contains(url, "list=A"):
			entries = []PlaylistEntry{
				{ID: "aaaaaaaaaa1", Title: "Track One (Radio Edit)", Channel: "Artist - Topic"},
				{ID: "aaaaaaaaaa2", Title: "Track One (Extended)", Channel: "Artist - Topic"},
				{ID: "failfailfai", Title: "Broken", Channel: "Someone"},
			}
			for i := range entries {
				entries[i].Playlist, entries[i].PlaylistTitle = "Playlist A", "Playlist A"
				entries[i].PlaylistID, entries[i].PlaylistIndex, entries[i].PlaylistCount = "A", i+1, 3
			}
		case strings.Contains(url, "list=B"):
			entries = []PlaylistEntry{{ID: "bbbbbbbbbb1", Title: "Normal upload", Channel: "Band",
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

func stubEngine(t *testing.T, official bool, urls ...string) (*Engine, *OfficialResolver) {
	t.Helper()
	os.Setenv("MVD_STUB_YTDLP", "1")
	t.Cleanup(func() { os.Unsetenv("MVD_STUB_YTDLP") })

	cfg := defaultConfig()
	cfg.OutputDir = t.TempDir()
	cfg.MaxConcurrentDownloads = 2
	cfg.ExtraArgs = nil
	cfg.DownloadOfficialMusicVideo = official

	eng, err := NewEngine(cfg, os.Args[0], urls, "")
	if err != nil {
		t.Fatal(err)
	}

	var res *OfficialResolver
	if official {
		f := &fakeYouTube{
			pages: map[string]string{
				"aaaaaaaaaa1": watchPage("aaaaaaaaaa1", "OFFICIAL001", ""),
				"aaaaaaaaaa2": watchPage("aaaaaaaaaa2", "OFFICIAL001", ""),
			},
			authors: map[string]string{"OFFICIAL001": "Label Records"},
		}
		_, res = f.server(t)
		eng.resolver = res
	}
	return eng, res
}

// collect drains events into a runState until EvIdle arrives (or timeout).
func collect(t *testing.T, eng *Engine, state *runState) []interface{} {
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
			state.apply(ev)
			if _, idle := ev.(EvIdle); idle {
				return seen
			}
		case <-timeout:
			t.Fatal("timed out waiting for EvIdle")
		}
	}
}

func TestEngineRun(t *testing.T) {
	eng, _ := stubEngine(t, true, "https://youtube.com/playlist?list=A", "https://youtube.com/playlist?list=B", "https://youtube.com/playlist?list=NOPE")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { eng.Run(ctx); close(done) }()

	state := newRunState(eng.Sources())
	events := collect(t, eng, state)

	tl := state.tally()
	if tl.Total != 4 || tl.Done != 2 || tl.Official != 1 || tl.Duplicate != 1 || tl.Failed != 1 || tl.Running != 0 || tl.Queued != 0 {
		t.Fatalf("unexpected tally: %+v", tl)
	}
	if state.Playlists[0].Title != "Playlist A" || state.Playlists[1].Title != "Playlist B" {
		t.Errorf("playlist titles not applied: %q %q", state.Playlists[0].Title, state.Playlists[1].Title)
	}
	if state.Playlists[2].Err == "" {
		t.Errorf("third playlist should have failed to list")
	}

	byVideo := map[string]*entryView{}
	for _, en := range state.Entries {
		byVideo[en.VideoID] = en
	}
	a1, a2, bad, b1 := byVideo["aaaaaaaaaa1"], byVideo["aaaaaaaaaa2"], byVideo["failfailfai"], byVideo["bbbbbbbbbb1"]
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
	if bad.State != StateFailed || !strings.Contains(bad.Err, "Video unavailable") {
		t.Errorf("bad entry should fail with yt-dlp's error: %+v", bad)
	}
	if b1.State != StateDone || b1.Official || b1.TargetID != "bbbbbbbbbb1" {
		t.Errorf("b1 should be downloaded as is: %+v", b1)
	}
	if a1.Percent != 100 || a1.Total != 1024 {
		t.Errorf("progress not applied: %+v", a1)
	}

	sawMerging, sawProgress := false, false
	for _, ev := range events {
		if st, ok := ev.(EvEntryState); ok && st.State == StateMerging {
			sawMerging = true
		}
		if _, ok := ev.(EvProgress); ok {
			sawProgress = true
		}
	}
	if !sawMerging || !sawProgress {
		t.Errorf("expected merging and progress events (merging=%v progress=%v)", sawMerging, sawProgress)
	}
	if len(a1.Log) == 0 || len(state.Playlists[0].Log) == 0 {
		t.Errorf("logs should be captured per entry and per playlist")
	}

	// Retry the failed entry: it fails again and the engine goes idle again.
	if !eng.Retry(bad.ID) {
		t.Fatal("retry of a failed entry should be accepted")
	}
	if eng.Retry(a1.ID) {
		t.Fatal("retry of a finished entry must be refused")
	}
	collect(t, eng, state)
	if state.Entries[bad.ID].State != StateFailed {
		t.Errorf("retried entry should fail again")
	}
	if n := eng.RetryPlaylist(0); n != 1 {
		t.Errorf("RetryPlaylist should re-queue 1 entry, got %d", n)
	}
	collect(t, eng, state)

	cancel()
	for range eng.Events() {
	}
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("engine did not stop after cancel")
	}
}

func TestRunPlain(t *testing.T) {
	eng, _ := stubEngine(t, false, "https://youtube.com/playlist?list=A")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { eng.Run(ctx); close(done) }()

	var out bytes.Buffer
	tl := runPlain(ctx, cancel, eng, &out)
	<-done

	if tl.Done != 2 || tl.Failed != 1 || tl.Official != 0 {
		t.Fatalf("unexpected tally %+v\n%s", tl, out.String())
	}
	for _, want := range []string{`"Playlist A": 3 entries`, "[P1/01] DONE Track One (Radio Edit)", "[P1/03] FAILED Broken: [youtube] failfailfai: Video unavailable", "Download Summary", "failed:               1"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("plain output missing %q:\n%s", want, out.String())
		}
	}
}

func TestParseProgressLine(t *testing.T) {
	p, ok := parseProgressLine("MVD|512|NA|1024|2048.5|7")
	if !ok || p.Downloaded != 512 || p.Total != 1024 || p.Percent != 50 || p.Speed != 2048.5 || p.ETA != 7 {
		t.Errorf("unexpected progress: %+v ok=%v", p, ok)
	}
	p, ok = parseProgressLine("MVD|512|NA|NA|NA|NA")
	if !ok || p.Total != 0 || p.Percent != 0 || p.Speed != 0 || p.ETA != -1 {
		t.Errorf("unknown totals should be zero: %+v", p)
	}
	if _, ok := parseProgressLine("[download]  34.2% of 112.4MiB"); ok {
		t.Error("normal yt-dlp lines are not progress lines")
	}
	if _, ok := parseProgressLine("MVD|1|2"); ok {
		t.Error("short lines are not progress lines")
	}
	if !isPostProcessLine(`[Merger] Merging formats into "x.mp4"`) || isPostProcessLine("[download] Destination: x") {
		t.Error("post process detection wrong")
	}
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

func TestHumanFormats(t *testing.T) {
	cases := map[int64]string{500: "500B", 2048: "2.0KiB", 117833728: "112.4MiB", 3 << 30: "3.0GiB"}
	for in, want := range cases {
		if got := humanBytes(in); got != want {
			t.Errorf("humanBytes(%d) = %q, want %q", in, got, want)
		}
	}
	if humanETA(9) != "0:09" || humanETA(3725) != "1:02:05" || humanETA(-1) != "--" {
		t.Error("humanETA wrong")
	}
	if humanSpeed(0) != "--" || humanSpeed(8.5*1024*1024) != "8.5MiB/s" {
		t.Error("humanSpeed wrong")
	}
}
