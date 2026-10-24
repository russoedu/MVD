package workspace

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/menubar"
)

func TestWhatCanBeASourceOfSongs(t *testing.T) {
	for _, ok := range []string{
		"https://www.youtube.com/watch?v=jNQXAC9IVRw",
		"http://example.test/x",
		"https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M",
		"www.youtube.com/playlist?list=PL1",
		"youtu.be/jNQXAC9IVRw",
		"music.apple.com/us/playlist/x/pl.1",
		`C:\Users\me\songs.txt`,
		`"C:\Users\me\Music\my songs.csv"`,
		"/home/me/songs.tsv",
		"  https://a.test/b  ",
		"",
	} {
		if reason := notALink(ok); reason != "" {
			t.Errorf("%q was refused: %s", ok, reason)
		}
	}
}

func TestWhatIsNotASourceOfSongsSaysWhy(t *testing.T) {
	cases := map[string]string{
		"hello this is not a link": "words that are not links",
		"hello":                    "does not look like a link",
		"ftp://example.test/x":     "Only web addresses",
		"https://":                 "not a complete web address",
		"javascript:alert(1)":      "does not look like a link",
		"just.some.words and more": "words that are not links",
	}
	for text, want := range cases {
		reason := notALink(text)
		if reason == "" || !strings.Contains(reason, want) {
			t.Errorf("%q: got %q, want a reason containing %q", text, reason, want)
		}
	}
}

func TestALineThatIsNotALinkIsNotAddedAndTheRowStaysOpenWithTheReason(t *testing.T) {
	m := listOf()
	next, _ := m.Update(menubar.SelectedMsg{Name: entryAddLink})
	m = next.(Model)
	m, _ = pressKey(m, tea.KeyPressMsg{Text: "h", Code: 'h'})
	m, _ = pressKey(m, tea.KeyPressMsg{Text: " ", Code: ' '})
	m, _ = pressKey(m, tea.KeyPressMsg{Text: "x", Code: 'x'})
	m, _ = pressKey(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if !m.inline.on {
		t.Error("the row closed on a line that is not a link")
	}
	if len(m.session.Sources()) != 0 {
		t.Errorf("a line that is not a link was added: %v", m.session.Sources())
	}
	if screen := ansi.Strip(m.render()); !strings.Contains(screen, "not links") {
		t.Errorf("the reason is not shown:\n%s", screen)
	}
}

func TestSeveralLinksPastedTogetherAreSeveralSources(t *testing.T) {
	links, reason := parseLinks("https://youtu.be/a https://youtu.be/b\nopen.spotify.com/playlist/c")
	if reason != "" || len(links) != 3 {
		t.Errorf("links %v, reason %q", links, reason)
	}
	if _, reason := parseLinks("https://youtu.be/a and a word"); reason == "" {
		t.Error("a word among links was accepted")
	}
}

func TestSeveralLinksTypedInTheRowAreAddedAsSeveralRows(t *testing.T) {
	m := listOf()
	next, _ := m.Update(menubar.SelectedMsg{Name: entryAddLink})
	m = next.(Model)
	for _, r := range "https://youtu.be/a https://youtu.be/b" {
		m, _ = pressKey(m, tea.KeyPressMsg{Text: string(r), Code: r})
	}
	m, _ = pressKey(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if got := m.session.Sources(); len(got) != 2 {
		t.Errorf("sources %v, want two", got)
	}
}

func TestALinkIsAddedWhenEnterIsPressed(t *testing.T) {
	m := listOf()
	next, _ := m.Update(menubar.SelectedMsg{Name: entryAddLink})
	m = next.(Model)
	for _, r := range "https://youtu.be/abc" {
		m, _ = pressKey(m, tea.KeyPressMsg{Text: string(r), Code: r})
	}
	m, _ = pressKey(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.inline.on || len(m.session.Sources()) != 1 {
		t.Errorf("row open %v, sources %v", m.inline.on, m.session.Sources())
	}
}

func TestAnErrorGoesAtTheNextKey(t *testing.T) {
	m := listOf()
	m.notice = "Stop the downloads before starting a new list"
	m, _ = pressKey(m, tea.KeyPressMsg{Code: tea.KeyDown})
	if m.notice != "" {
		t.Errorf("the error is still there: %q", m.notice)
	}
	if screen := ansi.Strip(m.render()); strings.Contains(screen, "Stop the downloads") {
		t.Error("the error is still on screen")
	}
}
