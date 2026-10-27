package engine

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"youtube-downloader/libs/mvd-core/ytdlp"
)

func TestPlanOnlyProposesVersionsAndDownloadsNothing(t *testing.T) {
	eng := stubEngine(t, true, "https://youtube.com/playlist?list=A", "https://youtube.com/playlist?list=B")
	eng.opts.PlanOnly = true
	eng.live = liveOf(eng.opts)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { eng.Run(ctx); close(done) }()
	events := collect(t, eng)
	cancel()
	<-done

	for _, ev := range events {
		switch e := ev.(type) {
		case EvProgress:
			t.Fatalf("a plan-only run downloaded something: %+v", e)
		case EvEntryState:
			if e.State == StateDownloading || e.State == StateDone {
				t.Fatalf("a plan-only run reached %v", e.State)
			}
		}
	}

	byUpload := map[string]PlannedEntry{}
	for _, p := range eng.Plan() {
		byUpload[p.Upload.ID] = p
	}
	if len(byUpload) != 4 {
		t.Fatalf("plan has %d entries, want 4: %+v", len(byUpload), byUpload)
	}
	if p := byUpload["aaaaaaaaaa1"]; p.Kind != KindOfficial || p.TargetID != "OFFICIAL001" || p.Reason == "" {
		t.Errorf("art track: %+v", p)
	}
	if p := byUpload["bbbbbbbbbb1"]; p.Kind != KindBetter || p.TargetID != "BETTER00001" {
		t.Errorf("normal upload: %+v", p)
	}
	if p := byUpload["failfailfai"]; p.Kind != KindOwnOfficial || p.TargetID != "failfailfai" {
		t.Errorf("an upload the lookup left as it is: %+v", p)
	}
	if p := byUpload["aaaaaaaaaa1"]; p.Playlist != "Playlist A" || p.PlaylistURL != "https://youtube.com/playlist?list=A" || p.Upload.PlaylistIndex != 1 {
		t.Errorf("playlist of the entry: %+v", p)
	}
}

func TestPlanOnlyKeepsTheSongsWithNoOfficialVideoInThePlan(t *testing.T) {
	eng := stubEngine(t, true, "https://youtube.com/playlist?list=B")
	eng.opts.PlanOnly = true
	eng.opts.OfficialOnly = true
	eng.live = liveOf(eng.opts)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { eng.Run(ctx); close(done) }()
	collect(t, eng)
	cancel()
	<-done

	plan := eng.Plan()
	if len(plan) != 1 || plan[0].Kind != KindNone || plan[0].TargetID != "bbbbbbbbbb1" {
		t.Fatalf("plan %+v: a song with no official video is kept, with no proposal", plan)
	}
}

func TestARunFromAPlanDownloadsTheVideosItNamesAndLooksForNothing(t *testing.T) {
	t.Setenv("MVD_STUB_YTDLP", "1")
	upload := func(id, title string, index int) ytdlp.PlaylistEntry {
		return ytdlp.PlaylistEntry{ID: id, Title: title, Channel: "Artist - Topic", PlaylistTitle: "Mine", Playlist: "Mine", PlaylistIndex: index, PlaylistCount: 2}
	}
	eng, err := New(Options{
		YtDlp:             os.Args[0],
		OutputDir:         t.TempDir(),
		OutputTemplate:    "%(playlist_title)s/%(playlist_index)02d - %(title)s.%(ext)s",
		Quality:           "best",
		MergeOutputFormat: "mp4",
		Workers:           2,
		// The resolver would send an art track to OFFICIAL001: the plan says otherwise.
		Resolver: fakeResolver{},
		Plan: []PlannedEntry{
			{Playlist: "Mine", PlaylistURL: "https://youtube.com/playlist?list=M", Upload: upload("aaaaaaaaaa1", "One", 1), TargetID: "CHOSEN00001", Kind: KindChosen},
			{Playlist: "Mine", PlaylistURL: "https://youtube.com/playlist?list=M", Upload: upload("aaaaaaaaaa2", "Two", 2), TargetID: "aaaaaaaaaa2", Kind: KindOriginal},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := eng.Sources(); len(got) != 1 || got[0].URL != "https://youtube.com/playlist?list=M" {
		t.Fatalf("sources %+v: one playlist, as the plan has", got)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { eng.Run(ctx); close(done) }()
	events := collect(t, eng)
	cancel()
	<-done

	var listed *EvPlaylistListed
	for _, ev := range events {
		if l, ok := ev.(EvPlaylistListed); ok {
			listed = &l
		}
	}
	if listed == nil || listed.Title != "Mine" || len(listed.Entries) != 2 {
		t.Fatalf("playlist announced as %+v", listed)
	}

	states := final(events)
	if st := states[0]; st.State != StateDone || st.TargetID != "CHOSEN00001" || st.Official || st.Better {
		t.Errorf("the chosen video: %+v", st)
	}
	if st := states[1]; st.State != StateDone || st.TargetID != "aaaaaaaaaa2" {
		t.Errorf("the original: %+v", st)
	}
}

// A list kept the songs a search found, and one whose video the run that stopped had not found
// yet: carrying it on downloads the found one as planned and searches for the other alone, from
// its title and artist, instead of searching for the whole list again.
func TestASongOfAPlanWithNoVideoYetIsSearchedForAndTheOthersAreNot(t *testing.T) {
	t.Setenv("MVD_STUB_YTDLP", "1")
	searched := []string{}
	var mu sync.Mutex
	tracks := recordingTracks{fakeTrackSource{}, func(title string) {
		mu.Lock()
		searched = append(searched, title)
		mu.Unlock()
	}}
	address := "https://music.test/search"
	eng, err := New(Options{
		YtDlp:             os.Args[0],
		OutputDir:         t.TempDir(),
		OutputTemplate:    "%(playlist_title)s/%(playlist_index)02d - %(title)s.%(ext)s",
		Quality:           "best",
		MergeOutputFormat: "mp4",
		Workers:           2,
		Tracks:            tracks,
		Plan: []PlannedEntry{
			{Playlist: "Road trip", PlaylistURL: address, Upload: ytdlp.PlaylistEntry{Title: "Official Song", Channel: "Band", PlaylistIndex: 1}, TargetID: "officialvid1", Kind: KindOfficial},
			{Playlist: "Road trip", PlaylistURL: address, Upload: ytdlp.PlaylistEntry{Title: "Lyric Only Song", Channel: "Band", PlaylistIndex: 2}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { eng.Run(ctx); close(done) }()
	events := collect(t, eng)
	cancel()
	<-done

	states := final(events)
	if st := states[0]; st.State != StateDone || st.TargetID != "officialvid1" {
		t.Errorf("the song found before: %+v", st)
	}
	if st := states[1]; st.State != StateDone || st.TargetID != "lyricvideo1" {
		t.Errorf("the song not found before: %+v", st)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(searched) != 1 || searched[0] != "Lyric Only Song" {
		t.Errorf("searched for %q, want only the song with no video yet", searched)
	}
}

// recordingTracks is a track source that says which songs it was asked to find.
type recordingTracks struct {
	fakeTrackSource
	found func(title string)
}

func (r recordingTracks) Find(ctx context.Context, track Track, logf func(string, ...interface{})) (string, bool, error) {
	r.found(track.Title)
	return r.fakeTrackSource.Find(ctx, track, logf)
}

// onceAgain is a resolver that can search again: it knows a second better version.
type onceAgain struct{ fakeResolver }

func (onceAgain) PickWithout(_ Identification, rejected []string, _ func(string, ...interface{})) Resolution {
	for _, id := range rejected {
		if id == "BETTER00001" {
			return Resolution{VideoID: "BETTER00002", Reason: "the next best version"}
		}
	}
	return Resolution{VideoID: "BETTER00001"}
}

func runB(t *testing.T, resolver OfficialResolver, check func(ctx context.Context, path string) (string, error)) EvEntryState {
	t.Helper()
	t.Setenv("MVD_STUB_YTDLP", "1")
	eng, err := New(Options{
		YtDlp:             os.Args[0],
		URLs:              []string{"https://youtube.com/playlist?list=B"},
		OutputDir:         t.TempDir(),
		OutputTemplate:    "%(playlist_title)s/%(playlist_index)02d - %(title)s.%(ext)s",
		Quality:           "best",
		MergeOutputFormat: "mp4",
		Workers:           1,
		Resolver:          resolver,
		CheckDownload:     check,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { eng.Run(ctx); close(done) }()
	events := collect(t, eng)
	cancel()
	<-done
	return final(events)[0]
}

func oneSidedFor(id string) func(ctx context.Context, path string) (string, error) {
	return func(_ context.Context, path string) (string, error) {
		if filepath.Base(path) == id+".mp4" {
			return "the sound plays on one side only", nil
		}
		return "", nil
	}
}

func TestAReplacementWithOneSidedSoundIsSearchedAgain(t *testing.T) {
	st := runB(t, onceAgain{}, oneSidedFor("BETTER00001"))
	if st.State != StateDone || st.TargetID != "BETTER00002" {
		t.Errorf("got %+v, want the next best version", st)
	}
}

func TestWhenNothingElseIsFoundThePlaylistsOwnUploadIsKept(t *testing.T) {
	st := runB(t, fakeResolver{}, oneSidedFor("BETTER00001"))
	if st.State != StateDone || st.TargetID != "bbbbbbbbbb1" || st.Official || st.Better {
		t.Errorf("got %+v, want the playlist's own upload", st)
	}
}

func TestTheOwnUploadIsNeverCheckedOrRejected(t *testing.T) {
	var checked []string
	var mu sync.Mutex
	st := runB(t, nil, func(_ context.Context, path string) (string, error) {
		mu.Lock()
		defer mu.Unlock()
		checked = append(checked, path)
		return "the sound plays on one side only", nil
	})
	mu.Lock()
	defer mu.Unlock()
	if st.State != StateDone || st.TargetID != "bbbbbbbbbb1" || len(checked) != 0 {
		t.Errorf("got %+v, checked %v", st, checked)
	}
}
