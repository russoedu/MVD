package urllist

import "testing"

func TestStartAndReviewNeedAtLeastOneLink(t *testing.T) {
	l := New(nil)
	if l.Begin() != None || l.ReviewLinks() != None {
		t.Error("an empty list should start nothing and review nothing")
	}
	l = New([]string{"https://a"})
	if l.Begin() != Start || l.ReviewLinks() != Review {
		t.Error("a list with a link should start and review")
	}
	if l.OpenPreferences() != Prefs || l.OpenSongs() != Songs {
		t.Error("preferences and songs open whatever the list holds")
	}
}

func TestLeavingAnEmptyListQuitsAtOnceAndAFullOneAsksFirst(t *testing.T) {
	empty := New(nil)
	if empty.Leave() != Quit || empty.ConfirmQuit {
		t.Error("an empty list should quit without asking")
	}
	l := New([]string{"https://a"})
	if l.Leave() != None || !l.ConfirmQuit {
		t.Fatal("a list with a link should ask before quitting")
	}
	if l.Answer(false) != None || l.ConfirmQuit {
		t.Error("anything but yes should keep the list and close the question")
	}
	l.Leave()
	if l.Answer(true) != Quit {
		t.Error("yes should quit")
	}
}

func TestResetEmptiesTheList(t *testing.T) {
	l := New([]string{"https://a", "https://b"})
	l.Reset()
	if len(l.URLs()) != 0 || l.Text != "" {
		t.Errorf("reset left %q", l.Text)
	}
}

func TestAddURLDoesNotRepeatALine(t *testing.T) {
	l := New([]string{"https://a"})
	l.AddURL("x.txt")
	l.AddURL("x.txt")
	if got := l.URLs(); len(got) != 2 || got[1] != "x.txt" {
		t.Errorf("got %v", got)
	}
}

func TestTheCountReadsAsAWordPerItem(t *testing.T) {
	if got := New(nil).Count(); got != "" {
		t.Errorf("an empty list counts %q", got)
	}
	if got := New([]string{"a"}).Count(); got != "(1 item)" {
		t.Errorf("one link counts %q", got)
	}
	if got := New([]string{"a", " ", "b"}).Count(); got != "(2 items)" {
		t.Errorf("two links and a blank line count %q", got)
	}
}
