package review

import (
	"strings"
	"testing"
	"time"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/ytdlp"
)

// fakeLive is a run that holds its songs, with the states the test gives them.
type fakeLive struct {
	snap     []engine.EntrySnapshot
	released map[int]string
	refuse   bool
}

func (f fakeLive) Snapshot() []engine.EntrySnapshot { return f.snap }

func (f fakeLive) Release(entry int, videoID string) bool {
	if f.refuse {
		return false
	}
	f.released[entry] = videoID

	return true
}

func (f fakeLive) ReleaseAll() int { return len(f.snap) }

type fakePlanRun struct{ plan []engine.PlannedEntry }

func (f *fakePlanRun) Sources() []engine.PlaylistSource { return nil }
func (f *fakePlanRun) Plan() []engine.PlannedEntry      { return f.plan }

func held(id, target string, state engine.EntryState) engine.EntrySnapshot {
	return engine.EntrySnapshot{
		PlannedEntry: engine.PlannedEntry{PlaylistURL: "https://list", Upload: ytdlp.PlaylistEntry{ID: id, Title: id}, TargetID: target, Kind: engine.KindOfficial},
		State:        state,
	}
}

func liveWith(states ...engine.EntryState) (Session, fakeLive) {
	f := fakeLive{released: map[int]string{}}
	for i, st := range states {
		f.snap = append(f.snap, held(string(rune('a'+i)), "t"+string(rune('a'+i)), st))
	}

	return NewLive(Options{}, f, nil), f
}

func TestDecidingOnASongOfALiveReviewSendsItToDownload(t *testing.T) {
	s, f := liveWith(engine.StatePlanned, engine.StatePlanned)
	s.Confirm()
	if v, ok := f.released[0]; !ok || v != "" {
		t.Fatalf("released %v", f.released)
	}
	if s.Version != 1 || s.States[0] != engine.StateQueued {
		t.Errorf("version %d state %v", s.Version, s.States[0])
	}
	if s.Selected != 1 {
		t.Errorf("cursor %d", s.Selected)
	}
}

func TestTheOriginalIsSentAsTheUploadOfALiveReview(t *testing.T) {
	s, f := liveWith(engine.StatePlanned)
	s.Decide(playlistfile.DecisionOriginal, "", "")
	answerNote(&s)
	if f.released[0] != "a" {
		t.Fatalf("released %v", f.released)
	}
}

func TestASongStillBeingLookedUpCannotBeDecidedOn(t *testing.T) {
	s, f := liveWith(engine.StateResolving)
	s.Confirm()
	if len(f.released) != 0 || s.Rows[0].Decision != playlistfile.DecisionNone {
		t.Fatalf("released %v decision %q", f.released, s.Rows[0].Decision)
	}
	if text, _ := s.NoticeShown(time.Now()); !strings.Contains(text, "still being looked up") {
		t.Errorf("notice %q", text)
	}
}

func TestASongAlreadySentCannotBeDecidedOnAgain(t *testing.T) {
	s, _ := liveWith(engine.StateDownloading)
	s.Confirm()
	if text, _ := s.NoticeShown(time.Now()); !strings.Contains(text, "sent to download already") {
		t.Errorf("notice %q", text)
	}
	if s.Version != 0 {
		t.Errorf("version %d", s.Version)
	}
}

func TestARunThatRefusesASongKeepsTheDecisionOff(t *testing.T) {
	f := fakeLive{released: map[int]string{}, refuse: true, snap: []engine.EntrySnapshot{held("a", "ta", engine.StatePlanned)}}
	s := NewLive(Options{}, f, nil)
	s.Confirm()
	if s.Rows[0].Decision != playlistfile.DecisionNone {
		t.Fatalf("decision %q", s.Rows[0].Decision)
	}
}

func TestDecisionsSurviveARefreshOfALiveReview(t *testing.T) {
	s, _ := liveWith(engine.StatePlanned, engine.StatePlanned)
	s.Confirm()
	s.Refresh()
	if s.Rows[0].Decision != playlistfile.DecisionConfirmed {
		t.Fatalf("the decision was lost: %q", s.Rows[0].Decision)
	}
}

func TestDecisionsBroughtFromAnEarlierSessionAreOnTheirSongs(t *testing.T) {
	f := fakeLive{released: map[int]string{}, snap: []engine.EntrySnapshot{held("a", "ta", engine.StatePlanned)}}
	brought := []playlistfile.Entry{{PlaylistURL: "https://list", Upload: ytdlp.PlaylistEntry{ID: "a"}, Decision: playlistfile.DecisionSkipped}}
	s := NewLive(Options{}, f, brought)
	if s.Rows[0].Decision != playlistfile.DecisionSkipped {
		t.Fatalf("decision %q", s.Rows[0].Decision)
	}
}

func TestConfirmAllOfALiveReviewSendsEverythingThatIsHeld(t *testing.T) {
	s, _ := liveWith(engine.StatePlanned, engine.StateDownloading)
	s.ConfirmAll()
	if s.Rows[0].Decision != playlistfile.DecisionConfirmed || s.Rows[1].Decision != playlistfile.DecisionNone {
		t.Fatalf("decisions %q %q", s.Rows[0].Decision, s.Rows[1].Decision)
	}
	if text, _ := s.NoticeShown(time.Now()); text != "2 songs sent to download as proposed" {
		t.Errorf("notice %q", text)
	}
}

func TestTheWordsOfALiveReviewSayWhereEachSongHasGot(t *testing.T) {
	s, _ := liveWith(engine.StateQueued, engine.StateDownloading, engine.StateFailed, engine.StateDone, engine.StatePlanned)
	s.Rows[0].TargetID = ""
	want := []string{"looking up…", "downloading", "failed", "downloaded", "official"}
	for i, w := range want {
		if got := s.RowLabel(i); got != w {
			t.Errorf("song %d says %q, want %q", i, got, w)
		}
	}
	if !s.Waiting(0) || !s.Sent(1) || s.Sent(0) || !s.Dimmed(3) || s.Dimmed(4) {
		t.Error("waiting, sent or dimmed is wrong")
	}
	if title := s.Title(); !strings.HasPrefix(title, "MVD · Choose what to download  ·  5 songs") || !strings.Contains(title, "1 being looked up") {
		t.Errorf("title %q", title)
	}
}
