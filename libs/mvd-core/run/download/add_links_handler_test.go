package download

import (
	"strings"
	"testing"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/testing/keys"
	"youtube-downloader/libs/mvd-core/testing/mouse"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/engine"
)

// addableController is a run that takes more links.
type addableController struct {
	fakeController
	added []string
	ended bool
}

func (a *addableController) AddSource(url string) (engine.PlaylistSource, bool) {
	if a.ended {
		return engine.PlaylistSource{}, false
	}
	a.added = append(a.added, url)
	return engine.PlaylistSource{Index: len(a.added), URL: url}, true
}

func addingScreen(t *testing.T) (Model, *addableController) {
	t.Helper()
	c := &addableController{fakeController: fakeController{events: make(chan interface{}), sources: []engine.PlaylistSource{{Index: 0, URL: "https://a"}}}}
	m := New(c)
	m.width, m.height = 120, 36
	return m, c
}

func press(t *testing.T, m Model, msgs ...tea.Msg) Model {
	t.Helper()
	var next tea.Model = m
	for _, msg := range msgs {
		next, _ = next.Update(msg)
	}
	return next.(Model)
}

func TestAddingLinksQueuesEveryLineOnTheRun(t *testing.T) {
	m, c := addingScreen(t)

	m = press(t, m, keys.Press("a"))
	if !m.adding {
		t.Fatal("the key should open the box")
	}
	if !strings.Contains(m.View().Content, "Add links") {
		t.Errorf("the box should be on screen:\n%s", m.View().Content)
	}

	// A paste of two links and a blank line, as a terminal sends it.
	m = press(t, m, tea.PasteMsg{Content: "https://b\n\n https://c "})
	m = press(t, m, keys.Press("ctrl+s"))

	if m.adding {
		t.Error("adding should close the box")
	}
	if len(c.added) != 2 || c.added[0] != "https://b" || c.added[1] != "https://c" {
		t.Errorf("queued %v", c.added)
	}
	if m.status != "added 2 links" {
		t.Errorf("status %q", m.status)
	}
}

func TestEscapeClosesTheBoxWithoutAdding(t *testing.T) {
	m, c := addingScreen(t)
	m = press(t, m, keys.Press("a"), keys.Text("https://b"), keys.Press("esc"))

	if m.adding || len(c.added) != 0 {
		t.Errorf("adding %v, queued %v", m.adding, c.added)
	}
}

func TestAnEmptyBoxAddsNothingAndSaysSo(t *testing.T) {
	m, c := addingScreen(t)
	m = press(t, m, keys.Press("a"), keys.Press("ctrl+s"))

	if len(c.added) != 0 || m.status != "no links to add" {
		t.Errorf("queued %v, status %q", c.added, m.status)
	}
}

func TestLinksAddedAfterTheRunEndedAreReported(t *testing.T) {
	m, c := addingScreen(t)
	c.ended = true
	m = press(t, m, keys.Press("a"), keys.Text("https://b"), keys.Press("ctrl+s"))

	if m.status != "added 0 of 1 links; the run has ended" {
		t.Errorf("status %q", m.status)
	}
}

func TestARunThatCannotTakeLinksHasNoAddKey(t *testing.T) {
	m := sampleModel(t) // its controller is not a SourceAdder
	m.width, m.height = 120, 36

	for _, h := range m.hints() {
		if h.Key == "a" {
			t.Error("the key bar should not offer add")
		}
	}
	m = press(t, m, keys.Press("a"))
	if m.adding || m.status != "this run cannot take more links" {
		t.Errorf("adding %v, status %q", m.adding, m.status)
	}
}

func TestTheAddKeyIsInTheKeyBarAndClickable(t *testing.T) {
	m, _ := addingScreen(t)

	found := false
	for x := 0; x < m.width && !found; x++ {
		if h, ok := keybar.At(m.hints(), x); ok && h.Key == "a" {
			found = true
			m = press(t, m, mouse.Click(x, m.height-1))
		}
	}
	if !found {
		t.Fatal("the key bar should offer add")
	}
	if !m.adding {
		t.Error("clicking the entry should open the box")
	}
}

func TestTheBoxIsDescribedForAScreenReader(t *testing.T) {
	m, _ := addingScreen(t)
	m = press(t, m, keys.Press("a"), keys.Text("https://b"))

	out := m.Outline()
	if out.Title != "MVD · Add links" || out.Text != "https://b" || !out.HasText || out.TextLabel == "" {
		t.Errorf("outline %+v", out)
	}
}

func TestTheMouseDoesNothingWhileTheBoxIsOpen(t *testing.T) {
	m, _ := addingScreen(t)
	m = press(t, m, keys.Press("a"))
	before := m.selPlaylist

	m = press(t, m, mouse.Click(5, 5))
	if !m.adding || m.selPlaylist != before {
		t.Error("a click should not reach the screen behind the box")
	}
}
