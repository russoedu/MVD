package list

import (
	"testing"

	"youtube-downloader/libs/mvd-core/testing/keys"
)

func TestListModel(t *testing.T) {
	m := New([]string{"https://a", "https://b"})
	if got := m.URLs(); len(got) != 2 {
		t.Fatalf("want 2 urls, got %v", got)
	}

	// ctrl+s with entries starts; empty does nothing.
	if _, _, out := m.Update(keys.Press("ctrl+s")); out != Start {
		t.Errorf("ctrl+s should start, got %d", out)
	}
	if _, _, out := m.Update(keys.Press("ctrl+p")); out != Prefs {
		t.Errorf("ctrl+p should open prefs, got %d", out)
	}

	cleared, _, _ := m.Update(keys.Press("ctrl+r"))
	if len(cleared.URLs()) != 0 {
		t.Errorf("ctrl+r should clear, got %v", cleared.URLs())
	}
	if _, _, out := cleared.Update(keys.Press("ctrl+s")); out != None {
		t.Errorf("ctrl+s on empty list should do nothing, got %d", out)
	}
	if _, _, out := cleared.Update(keys.Press("ctrl+q")); out != Quit {
		t.Errorf("ctrl+q on empty list should quit, got %d", out)
	}

	// Quit with entries asks to confirm, then y quits.
	confirming, _, out := m.Update(keys.Press("ctrl+q"))
	if out != None || !confirming.confirmQuit {
		t.Fatalf("ctrl+q with entries should ask to confirm")
	}
	if _, _, out := confirming.Update(keys.Press("y")); out != Quit {
		t.Errorf("y should confirm quit, got %d", out)
	}
}

func TestListWithURLDoesNotRepeatALine(t *testing.T) {
	m := New([]string{"https://a"}).WithURL("x.txt").WithURL("x.txt")
	if got := m.URLs(); len(got) != 2 || got[1] != "x.txt" {
		t.Errorf("got %v", got)
	}
}
