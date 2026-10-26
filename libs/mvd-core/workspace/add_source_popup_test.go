package workspace

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/musicsearch"
)

// key sends one key press.
func key(m Model, msg tea.KeyPressMsg) Model {
	next, _ := m.Update(msg)

	return next.(Model)
}

// openSearch opens the window for adding and chooses the search by name, with the focus on the artist.
func openSearch(t *testing.T, m Model) Model {
	t.Helper()
	m, _ = choose(m, entryAddLink)
	m = key(m, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}) // up to the choice of mode
	m = key(m, tea.KeyPressMsg{Code: tea.KeyRight})                  // to the search
	m = key(m, tea.KeyPressMsg{Code: tea.KeyTab})                    // to the artist
	if !m.addOpen || m.add.Mode != 1 || m.add.Focus != 2 {
		t.Fatalf("the search was not chosen: %+v", m.add)
	}

	return m
}

func TestAnArtistAloneIsAddedAsASearchInTheList(t *testing.T) {
	m := openSearch(t, listOf())
	m = typeInto(m, "Daft Punk")
	m, _ = press(m, "enter")
	want := musicsearch.Query{Artist: "Daft Punk"}.Address()
	if got := addressesOf(m); m.addOpen || len(got) != 1 || got[0] != want {
		t.Errorf("window open %v, the list has %v, want [%s]", m.addOpen, got, want)
	}
}

func TestAnArtistAnAlbumAndASongAreOneSearch(t *testing.T) {
	m := openSearch(t, listOf())
	m = typeInto(m, "Daft Punk")
	m = key(m, tea.KeyPressMsg{Code: tea.KeyTab})
	m = typeInto(m, "Discovery")
	m = key(m, tea.KeyPressMsg{Code: tea.KeyTab})
	m = typeInto(m, "One More Time")
	m, _ = press(m, "enter")
	want := musicsearch.Query{Artist: "Daft Punk", Album: "Discovery", Song: "One More Time"}.Address()
	if got := addressesOf(m); len(got) != 1 || got[0] != want {
		t.Errorf("the list has %v, want [%s]", got, want)
	}
}

func TestASongOrAnAlbumAloneIsRefusedWithTheReason(t *testing.T) {
	m := openSearch(t, listOf())
	m = key(m, tea.KeyPressMsg{Code: tea.KeyTab}) // past the artist, left empty
	m = typeInto(m, "Discovery")
	m, _ = press(m, "enter")
	if !m.addOpen || len(addressesOf(m)) != 0 {
		t.Fatalf("an album with no artist was added: open %v, list %v", m.addOpen, addressesOf(m))
	}
	if screen := screenOf(m); !strings.Contains(screen, "An album needs its artist") {
		t.Errorf("the reason is not shown:\n%s", screen)
	}
	if m.add.Focus != 2 {
		t.Errorf("the focus should go to the artist, it is on %v", m.add.Focus)
	}
}

func TestTheSearchShowsItsThreeFieldsAndWhatEachFindsAndTheChoiceCanBeClicked(t *testing.T) {
	m := listOf()
	m, _ = choose(m, entryAddLink)
	l := m.addLayout()
	span := l.radio[1]
	next, _, _ := m.addMouse(tea.MouseClickMsg{X: span.X0 + 2, Y: l.radioY + 1, Button: tea.MouseLeft})
	m = next.(Model)
	if m.add.Mode != 1 {
		t.Fatal("a click on Music search did not choose it")
	}
	screen := screenOf(m)
	for _, want := range []string{"( ) Platform URL", "(\u2022) Music search", "Artist", "Album", "Song", "every song of the artist"} {
		if !strings.Contains(screen, want) {
			t.Errorf("the window lacks %q:\n%s", want, screen)
		}
	}
}

func TestASearchPastedOrWrittenInAListFileIsAcceptedAsASource(t *testing.T) {
	address := musicsearch.Query{Artist: "Daft Punk", Song: "One More Time"}.Address()
	if links, reason := parseLinks(address); reason != "" || len(links) != 1 {
		t.Errorf("a search address was refused: %v %q", links, reason)
	}
	if _, reason := parseLinks(musicsearch.Scheme + "album=Discovery"); reason == "" {
		t.Error("a search with no artist was accepted")
	}
}
