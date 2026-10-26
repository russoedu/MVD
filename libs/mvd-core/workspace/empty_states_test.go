package workspace

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/menubar"
)

func screenOf(m Model) string { return ansi.Strip(m.render()) }

func TestAnEmptyListShowsItsHelpInsteadOfBlankBoxes(t *testing.T) {
	screen := screenOf(sized(New(Host{}, config.Config{}, downloadlist.New(nil))))
	for _, want := range []string{"Add the songs", "music videos", "Ctrl+L", "Ctrl+V", "Nothing yet"} {
		if !strings.Contains(screen, want) {
			t.Errorf("the empty screen lacks %q: %s", want, screen)
		}
	}
}

func TestALinkThatWasNotLookedUpYetSaysHowToGoOn(t *testing.T) {
	screen := screenOf(listOf("https://www.youtube.com/watch?v=abc"))
	if !strings.Contains(screen, "Nothing loaded yet") || !strings.Contains(screen, "Ctrl+R") {
		t.Errorf("nothing says what to do after adding a link:\n%s", screen)
	}
	if strings.Contains(screen, "show here once you add one") {
		t.Error("the pane still says the list is empty")
	}
}

func TestTheWindowForAddingSaysWhatLinksAreTakenAndOffersASearch(t *testing.T) {
	m := listOf()
	next, _ := m.Update(menubar.SelectedMsg{Name: entryAddLink})
	m = next.(Model)
	if !m.addOpen {
		t.Fatal("adding did not open the window")
	}
	screen := screenOf(m)
	for _, want := range []string{"Add music", "Platform URL", "Music search", "YouTube playlist", "Spotify", "Apple Music"} {
		if !strings.Contains(screen, want) {
			t.Errorf("the window does not say %q: %s", want, screen)
		}
	}
	if !strings.Contains(screen, "(•) Platform URL") {
		t.Errorf("the platform address is not the one chosen at first: %s", screen)
	}
}

func TestPointingAtAButtonThatCannotBeUsedSaysWhy(t *testing.T) {
	m := listOf() // nothing to start
	x, y, w, ok := m.buttonRows()
	if !ok {
		t.Fatal("no buttons")
	}
	next, _ := m.Update(tea.MouseMotionMsg{X: x + w/2, Y: y}) // Start
	if screen := screenOf(next.(Model)); !strings.Contains(screen, "Add a source first.") {
		t.Errorf("the disabled Start does not say why:\n%s", screen)
	}
	next, _ = m.Update(tea.MouseMotionMsg{X: x + w/2, Y: y + 2}) // Add link, which can be used
	if screen := screenOf(next.(Model)); strings.Contains(screen, "Add a source first.") {
		t.Error("a button that works should not say it cannot")
	}
}

func TestWhileNothingRunsTheBottomLineListsTheKeysThatStartTheWork(t *testing.T) {
	screen := screenOf(listOf("https://a"))
	for _, want := range []string{"F1 shortcuts", "Ctrl+L add source", "Ctrl+R load playlists", "Ctrl+O open list"} {
		if !strings.Contains(screen, want) {
			t.Errorf("the key hints lack %q: %s", want, screen)
		}
	}
}
