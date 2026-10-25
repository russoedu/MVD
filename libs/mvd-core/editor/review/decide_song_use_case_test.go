package review

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlisteditor"
	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/ytdlp"
)

func song(id, target string, kind engine.PlanKind) playlistfile.Entry {
	return playlistfile.Entry{
		PlaylistURL: "https://list",
		Upload:      ytdlp.PlaylistEntry{ID: id, Title: "title " + id},
		TargetID:    target,
		Kind:        string(kind),
	}
}

func threeSongs() []playlistfile.Entry {
	return []playlistfile.Entry{
		song("u1", "o1", engine.KindOfficial),
		song("u2", "", engine.KindNone),
		song("u3", "b3", engine.KindBetter),
	}
}

// answerNote closes the question a build made for testing asks about a decision.
func answerNote(s *Session) {
	if playlisteditor.DebugBuild && s.Mode == Input {
		s.CancelInput()
	}
}

func TestConfirmingASongAcceptsItAndMovesToTheNext(t *testing.T) {
	s := NewFromRows(Options{}, threeSongs())
	s.Confirm()
	if s.Rows[0].Decision != playlistfile.DecisionConfirmed || s.Selected != 1 {
		t.Fatalf("decision %q, cursor %d", s.Rows[0].Decision, s.Selected)
	}
}

func TestASongWithNothingFoundCannotBeConfirmed(t *testing.T) {
	s := NewFromRows(Options{}, threeSongs())
	s.Select(1)
	s.Confirm()
	if s.Rows[1].Decision != playlistfile.DecisionNone {
		t.Fatalf("confirmed a song with nothing found: %q", s.Rows[1].Decision)
	}
	if text, ok := s.NoticeShown(time.Now()); !ok || !strings.Contains(text, "nothing was found") {
		t.Errorf("notice %q", text)
	}
}

func TestUndoingADecisionStaysOnTheSong(t *testing.T) {
	s := NewFromRows(Options{}, threeSongs())
	s.Confirm()
	s.Select(0)
	s.Decide(playlistfile.DecisionNone, "", "")
	if s.Rows[0].Decision != playlistfile.DecisionNone || s.Selected != 0 {
		t.Fatalf("decision %q, cursor %d", s.Rows[0].Decision, s.Selected)
	}
}

func TestADownloadedSongCannotBeDecidedOn(t *testing.T) {
	rows := threeSongs()
	rows[0].Downloaded = true
	s := NewFromRows(Options{}, rows)
	s.Decide(playlistfile.DecisionSkipped, "", "")
	if s.Rows[0].Decision != playlistfile.DecisionNone {
		t.Fatal("a downloaded song was skipped")
	}
}

func TestTakingTheOriginalNeedsAnUpload(t *testing.T) {
	rows := threeSongs()
	rows[0].Upload.ID = ""
	s := NewFromRows(Options{}, rows)
	s.Decide(playlistfile.DecisionOriginal, "", "")
	if s.Rows[0].Decision != playlistfile.DecisionNone {
		t.Fatal("took an original that is not there")
	}
}

func TestConfirmAllConfirmsOnlyWhatHasAVideoAndNoDecision(t *testing.T) {
	s := NewFromRows(Options{}, threeSongs())
	s.Select(2)
	s.Decide(playlistfile.DecisionSkipped, "", "")
	answerNote(&s)
	s.ConfirmAll()
	got := []string{s.Rows[0].Decision, s.Rows[1].Decision, s.Rows[2].Decision}
	want := []string{playlistfile.DecisionConfirmed, playlistfile.DecisionNone, playlistfile.DecisionSkipped}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("song %d: %q, want %q", i, got[i], want[i])
		}
	}
}

func TestEveryDecisionIsKeptInTheSessionFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "session.mvd")
	s := NewFromRows(Options{SessionFile: file}, threeSongs())
	s.Confirm()
	saved, err := playlistfile.Load(file)
	if err != nil {
		t.Fatal(err)
	}
	if got := saved.Songs(); len(got) != 3 || got[0].Decision != playlistfile.DecisionConfirmed {
		t.Fatalf("the session file has %+v", got)
	}
}

func TestReplacingAsksForAnAddressAndNamesTheVideoPicked(t *testing.T) {
	s := NewFromRows(Options{CanDescribe: true}, threeSongs())
	s.AskReplace()
	if s.Mode != Input || !strings.Contains(s.Prompt, "Address of the video") {
		t.Fatalf("mode %v prompt %q", s.Mode, s.Prompt)
	}

	s.SetInput("not an address")
	if effects := s.SubmitInput(); len(effects) != 0 || s.Rows[0].Decision != playlistfile.DecisionNone {
		t.Fatalf("a bad address decided something: %v", effects)
	}

	s.AskReplace()
	s.SetInput("https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	effects := s.SubmitInput()
	if s.Rows[0].Decision != playlistfile.DecisionReplaced || s.Rows[0].ChosenID != "dQw4w9WgXcQ" {
		t.Fatalf("decision %q chosen %q", s.Rows[0].Decision, s.Rows[0].ChosenID)
	}
	if len(effects) != 1 {
		t.Fatalf("effects %v", effects)
	}
	describe, ok := effects[0].(DescribeVideo)
	if !ok || describe.VideoID != "dQw4w9WgXcQ" {
		t.Fatalf("effect %#v", effects[0])
	}
	s.Described(describe.Key, "A title")
	if s.Rows[0].ChosenTitle != "A title" {
		t.Errorf("the title is %q", s.Rows[0].ChosenTitle)
	}
}

func TestOpeningAVideoShowsItsAddressWhenTheHostCannotOpenIt(t *testing.T) {
	s := NewFromRows(Options{}, threeSongs())
	if effects := s.Open(false); len(effects) != 0 {
		t.Fatalf("effects %v", effects)
	}
	if text, _ := s.NoticeShown(time.Now()); text != "https://www.youtube.com/watch?v=o1" {
		t.Errorf("notice %q", text)
	}
}

func TestOpeningAVideoAsksTheHostAndSaysHowItWent(t *testing.T) {
	s := NewFromRows(Options{CanOpen: true}, threeSongs())
	effects := s.Open(true)
	if len(effects) != 1 || effects[0] != (OpenURL{URL: "https://www.youtube.com/watch?v=u1"}) {
		t.Fatalf("effects %#v", effects)
	}
	s.Opened("https://x", errors.New("no browser"))
	if text, _ := s.NoticeShown(time.Now()); text != "cannot open the browser: no browser" {
		t.Errorf("notice %q", text)
	}
	s.Opened("https://x", nil)
	if text, _ := s.NoticeShown(time.Now()); text != "opened https://x" {
		t.Errorf("notice %q", text)
	}
}

func TestTheCursorStaysOnTheList(t *testing.T) {
	s := NewFromRows(Options{}, threeSongs())
	s.Move(-5)
	if s.Selected != 0 {
		t.Errorf("moved above the list: %d", s.Selected)
	}
	s.Move(50)
	if s.Selected != 2 {
		t.Errorf("moved below the list: %d", s.Selected)
	}
	s.Select(7)
	if s.Selected != 2 {
		t.Errorf("selected a song that is not there: %d", s.Selected)
	}
}
