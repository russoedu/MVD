package engine

import (
	"context"
	"testing"
)

func TestSnapshotNamesEverySongWithItsVersionAndState(t *testing.T) {
	eng := stubEngine(t, true, "https://youtube.com/playlist?list=A")
	eng.opts.HoldForReview = true
	eng.live = liveOf(eng.opts)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { eng.Run(ctx); close(done) }()
	collect(t, eng)

	snap := eng.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("snapshot has %d songs, want 3: %+v", len(snap), snap)
	}
	byUpload := map[string]EntrySnapshot{}
	for _, s := range snap {
		byUpload[s.Upload.ID] = s
		if s.State != StatePlanned || s.PlaylistURL != "https://youtube.com/playlist?list=A" {
			t.Errorf("song %+v: want planned in playlist A", s)
		}
	}
	if s := byUpload["aaaaaaaaaa1"]; s.Kind != KindOfficial || s.TargetID != "OFFICIAL001" {
		t.Errorf("art track: %+v", s)
	}
	cancel()
	for range eng.Events() {
	}
	<-done
}
