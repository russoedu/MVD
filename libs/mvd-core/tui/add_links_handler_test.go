package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

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

func addingScreen(t *testing.T) (model, *addableController) {
	t.Helper()
	c := &addableController{fakeController: fakeController{events: make(chan interface{}), sources: []engine.PlaylistSource{{Index: 0, URL: "https://a"}}}}
	m := newModel(c)
	m.width, m.height = 120, 36
	return m, c
}

func press(t *testing.T, m model, msgs ...tea.Msg) model {
	t.Helper()
	var next tea.Model = m
	for _, msg := range msgs {
		next, _ = next.Update(msg)
	}
	return next.(model)
}

func TestAddingLinksQueuesEveryLineOnTheRun(t *testing.T) {
	m, c := addingScreen(t)

	m = press(t, m, key("a"))
	if !m.adding {
		t.Fatal("the key should open the box")
	}
	if !strings.Contains(m.View(), "Add links") {
		t.Errorf("the box should be on screen:\n%s", m.View())
	}

	// A paste of two links and a blank line, as a terminal sends it.
	m = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("https://b\n\n https://c "), Paste: true})
	m = press(t, m, key("ctrl+s"))

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
	m = press(t, m, key("a"), tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("https://b")}, key("esc"))

	if m.adding || len(c.added) != 0 {
		t.Errorf("adding %v, queued %v", m.adding, c.added)
	}
}

func TestAnEmptyBoxAddsNothingAndSaysSo(t *testing.T) {
	m, c := addingScreen(t)
	m = press(t, m, key("a"), key("ctrl+s"))

	if len(c.added) != 0 || m.status != "no links to add" {
		t.Errorf("queued %v, status %q", c.added, m.status)
	}
}

func TestLinksAddedAfterTheRunEndedAreReported(t *testing.T) {
	m, c := addingScreen(t)
	c.ended = true
	m = press(t, m, key("a"), tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("https://b")}, key("ctrl+s"))

	if m.status != "added 0 of 1 links; the run has ended" {
		t.Errorf("status %q", m.status)
	}
}

func TestARunThatCannotTakeLinksHasNoAddKey(t *testing.T) {
	m := sampleModel(t) // its controller is not a SourceAdder
	m.width, m.height = 120, 36

	for _, h := range m.hints() {
		if h.key == "a" {
			t.Error("the key bar should not offer add")
		}
	}
	m = press(t, m, key("a"))
	if m.adding || m.status != "this run cannot take more links" {
		t.Errorf("adding %v, status %q", m.adding, m.status)
	}
}

func TestTheAddKeyIsInTheKeyBarAndClickable(t *testing.T) {
	m, _ := addingScreen(t)

	found := false
	for x := 0; x < m.width && !found; x++ {
		if h, ok := hintAt(m.hints(), x); ok && h.key == "a" {
			found = true
			m = press(t, m, click(x, m.height-1))
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
	m = press(t, m, key("a"), tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("https://b")})

	out := m.Outline()
	if out.Title != "MVD · Add links" || out.Text != "https://b" || !out.HasText || out.TextLabel == "" {
		t.Errorf("outline %+v", out)
	}
}

func TestTheMouseDoesNothingWhileTheBoxIsOpen(t *testing.T) {
	m, _ := addingScreen(t)
	m = press(t, m, key("a"))
	before := m.selPlaylist

	m = press(t, m, click(5, 5))
	if !m.adding || m.selPlaylist != before {
		t.Error("a click should not reach the screen behind the box")
	}
}
