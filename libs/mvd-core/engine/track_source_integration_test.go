package engine

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

// fakeTrackSource lists a three song playlist and finds a video for two of them.
type fakeTrackSource struct{}

func (fakeTrackSource) Handles(link string) bool { return strings.HasPrefix(link, "https://music.test/") }

func (fakeTrackSource) Tracks(_ context.Context, link string) (string, []Track, error) {
	if strings.HasSuffix(link, "/broken") {
		return "", nil, errors.New("the playlist is private")
	}
	return "Road trip", []Track{
		{Title: "Official Song", Artist: "Band, Guest"},
		{Title: "Lyric Only Song", Artist: "Band"},
		{Title: "Lost Song", Artist: "Band"},
	}, nil
}

func (fakeTrackSource) Find(_ context.Context, track Track, logf func(string, ...interface{})) (string, bool, error) {
	logf("looked up %s", track.Title)
	switch track.Title {
	case "Official Song":
		return "officialvid1", true, nil
	case "Lyric Only Song":
		return "lyricvideo1", false, nil
	}
	return "", false, nil
}

func trackEngine(t *testing.T, urls ...string) *Engine {
	t.Helper()
	t.Setenv("MVD_STUB_YTDLP", "1")
	eng, err := New(Options{
		YtDlp:             os.Args[0],
		URLs:              urls,
		OutputDir:         t.TempDir(),
		OutputTemplate:    "%(playlist_title)s/%(playlist_index)02d - %(title)s.%(ext)s",
		Quality:           "best",
		MergeOutputFormat: "mp4",
		Workers:           2,
		Tracks:            fakeTrackSource{},
	})
	if err != nil {
		t.Fatal(err)
	}
	return eng
}

func TestEngineDownloadsTheTracksOfAnotherService(t *testing.T) {
	eng := trackEngine(t, "https://music.test/playlist/1", "https://music.test/broken")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { eng.Run(ctx); close(done) }()

	events := collect(t, eng)

	var entries []EntryInfo
	failedLists := 0
	for _, ev := range events {
		switch e := ev.(type) {
		case EvPlaylistListed:
			if e.Title != "Road trip" {
				t.Errorf("playlist title = %q", e.Title)
			}
			entries = e.Entries
		case EvPlaylistFailed:
			failedLists++
			if !strings.Contains(e.Err, "private") {
				t.Errorf("the playlist error should reach the person: %q", e.Err)
			}
		}
	}
	if len(entries) != 3 || failedLists != 1 {
		t.Fatalf("want 3 tracks and 1 failed playlist, got %d and %d", len(entries), failedLists)
	}
	if entries[0].Title != "Official Song" || entries[0].Channel != "Band, Guest" || entries[0].VideoID != "" {
		t.Errorf("a track shows its title and artists and has no video yet: %+v", entries[0])
	}

	states := final(events)
	official := states[entries[0].ID]
	if official.State != StateDone || !official.Official || official.TargetID != "officialvid1" {
		t.Errorf("the official video should be downloaded: %+v", official)
	}
	lyric := states[entries[1].ID]
	if lyric.State != StateDone || lyric.Official || lyric.TargetID != "lyricvideo1" {
		t.Errorf("the non-official video should be downloaded as the fallback: %+v", lyric)
	}
	lost := states[entries[2].ID]
	if lost.State != StateFailed || !strings.Contains(lost.Err, "no matching video") {
		t.Errorf("a track with no video should fail saying so: %+v", lost)
	}

	cancel()
	<-done
}
