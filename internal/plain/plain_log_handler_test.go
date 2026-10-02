package plain

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"youtube-downloader/internal/engine"
)

func TestRun(t *testing.T) {
	events := make(chan interface{}, 64)
	sources := []engine.PlaylistSource{{Index: 0, URL: "https://a"}, {Index: 1, URL: "https://b"}}
	ctx, cancel := context.WithCancel(context.Background())

	events <- engine.EvPlaylistListing{Playlist: 0, URL: "https://a"}
	events <- engine.EvLog{Playlist: 0, Entry: -1, Line: "listing https://a"}
	events <- engine.EvPlaylistListed{Playlist: 0, Title: "Playlist A", Entries: []engine.EntryInfo{
		{ID: 0, Playlist: 0, Index: 1, VideoID: "v0", Title: "Track One (Radio Edit)"},
		{ID: 1, Playlist: 0, Index: 2, VideoID: "v1", Title: "Broken"},
	}}
	events <- engine.EvPlaylistFailed{Playlist: 1, Err: "unable to recognize playlist"}
	events <- engine.EvEntryState{Entry: 0, State: engine.StateResolving}
	events <- engine.EvEntryState{Entry: 0, State: engine.StateDownloading, TargetID: "off", Official: true}
	events <- engine.EvProgress{Entry: 0, Percent: 55, Downloaded: 55, Total: 100, Speed: 1024, ETA: 3}
	events <- engine.EvProgress{Entry: 0, Percent: 57, Downloaded: 57, Total: 100, Speed: 1024, ETA: 3}
	events <- engine.EvEntryState{Entry: 0, State: engine.StateMerging, TargetID: "off", Official: true}
	events <- engine.EvEntryState{Entry: 0, State: engine.StateDone, TargetID: "off", Official: true}
	events <- engine.EvEntryState{Entry: 1, State: engine.StateDownloading, TargetID: "v1"}
	events <- engine.EvLog{Playlist: 0, Entry: 1, Line: "ERROR: [youtube] v1: Video unavailable"}
	events <- engine.EvEntryState{Entry: 1, State: engine.StateFailed, TargetID: "v1", Err: "[youtube] v1: Video unavailable"}
	events <- engine.EvIdle{}
	go func() {
		<-ctx.Done()
		close(events)
	}()

	var out bytes.Buffer
	tl := Run(ctx, cancel, events, sources, &out)

	if tl.Done != 1 || tl.Failed != 1 || tl.Official != 1 {
		t.Fatalf("unexpected tally %+v\n%s", tl, out.String())
	}
	for _, want := range []string{
		"[P1] listing https://a",
		`[P1] "Playlist A": 2 entries`,
		"[P2] FAILED to list: unable to recognize playlist",
		"[P1/01] resolving official video for Track One (Radio Edit)",
		"[P1/01] downloading Track One (Radio Edit) (official video off)",
		"[P1/01]  55.0% of 100B",
		"[P1/01] merging",
		"[P1/01] DONE Track One (Radio Edit)",
		"[P1/02] ERROR: [youtube] v1: Video unavailable",
		"[P1/02] FAILED Broken: [youtube] v1: Video unavailable",
		"Download Summary",
		"failed:               1",
		"[P2] https://b: unable to recognize playlist",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("plain output missing %q:\n%s", want, out.String())
		}
	}
	if strings.Count(out.String(), "55.0%") != 1 || strings.Contains(out.String(), "57.0%") {
		t.Errorf("progress should be printed once per 10%% step:\n%s", out.String())
	}
	if strings.Count(out.String(), "listing https://a") != 1 {
		t.Errorf("listing should be printed once:\n%s", out.String())
	}
}
