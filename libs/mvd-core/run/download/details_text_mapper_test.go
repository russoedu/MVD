package download

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/engine"
)

// detailsOf is the plain text of the details pane for the song at row i of the first link.
func detailsOf(t *testing.T, m Model, i int) string {
	t.Helper()
	m.focus, m.selPlaylist, m.selEntry = paneEntries, 0, i
	var out []string
	for _, row := range m.detailRows(60) {
		out = append(out, strings.TrimSpace(string(row.Line)))
	}

	return strings.Join(out, "\n")
}

func TestTheDetailsSayWhatHappenedToASongInEverydayWords(t *testing.T) {
	m := sampleModel(t)
	cases := map[int][]string{
		0: {"✓ Saved in your videos folder", "official music video"},
		1: {"you already have this song"},
		2: {"✗", "Press r to try again"},
		3: {"Looking for the best version"},
		4: {"Downloading"},
		5: {"picture and the sound together"},
	}
	for row, wants := range cases {
		got := detailsOf(t, m, row)
		for _, want := range wants {
			if !strings.Contains(got, want) {
				t.Errorf("song %d: the details lack %q:\n%s", row, want, got)
			}
		}
		for _, tool := range []string{"yt-dlp", "[youtube]", "[download]", "format(s)"} {
			if strings.Contains(got, tool) {
				t.Errorf("song %d: the tool's own words reached the details: %q", row, got)
			}
		}
	}
	if got := detailsOf(t, m, 0); strings.Contains(got, "-bsONE-kZwI") || strings.Contains(got, "rnlp_avexYQ") {
		t.Errorf("a video id reached the details: %s", got)
	}
}

func TestTheDetailsOfALinkSayHowItsSongsAreDoing(t *testing.T) {
	m := sampleModel(t)
	m.focus, m.selPlaylist = paneLists, 0
	var text []string
	for _, row := range m.detailRows(80) {
		text = append(text, strings.TrimSpace(string(row.Line)))
	}
	got := strings.Join(text, "\n")
	for _, want := range []string{"90s UK Dance Hits", "6 songs", "1 saved", "1 failed"} {
		if !strings.Contains(got, want) {
			t.Errorf("the details of the link lack %q:\n%s", want, got)
		}
	}
	m.selPlaylist = 2 // the link that could not be read
	text = nil
	for _, row := range m.detailRows(80) {
		text = append(text, strings.TrimSpace(string(row.Line)))
	}
	if got := strings.Join(text, "\n"); !strings.Contains(got, "✗") || strings.Contains(got, "yt-dlp") {
		t.Errorf("the details of a link that cannot be read say %q", got)
	}
}

func TestTheKeyTSwitchesBetweenTheDetailsAndTheTechnicalLog(t *testing.T) {
	m := sampleModel(t)
	m.width, m.height = 120, 36
	m.focus, m.selEntry = paneEntries, 4
	if m.technical || m.outputTitle() != "Details" {
		t.Fatalf("the plain details are not the default (title %q)", m.outputTitle())
	}
	m = press(t, m, tea.KeyPressMsg{Text: "t", Code: 't'})
	if !m.technical {
		t.Error("t did not show the technical log")
	}
	if m.outputTitle() == "Details" {
		t.Errorf("the technical log is titled %q", m.outputTitle())
	}
	if got := m.outputLines(); len(got) == 0 {
		t.Error("the technical log is empty")
	}
	m = press(t, m, tea.KeyPressMsg{Text: "t", Code: 't'})
	if m.technical {
		t.Error("t did not go back to the details")
	}
}

func TestTheWordsAtTheEndOfARowAreNotCodes(t *testing.T) {
	m := sampleModel(t)
	want := map[engine.EntryState]string{
		engine.StateQueued: "waiting", engine.StateResolving: "finding", engine.StateMerging: "joining",
		engine.StateDuplicate: "have it", engine.StateNotFound: "no official", engine.StateFailed: "failed",
	}
	for state, word := range want {
		en := m.state.Entries[0]
		en.State = state
		if got, _ := m.entryTag(en); got != word {
			t.Errorf("%v says %q, want %q", state, got, word)
		}
	}
	en := m.state.Entries[4]
	en.Total = 0
	if got, _ := m.entryTag(en); got != "starting" {
		t.Errorf("a download of unknown size says %q", got)
	}
}

func TestALinkIsNamedByWhatItIsAndNotByItsAddress(t *testing.T) {
	m := New(&fakeController{events: make(chan interface{}), sources: []engine.PlaylistSource{{Index: 0, URL: "https://www.youtube.com/watch?v=abc"}}})
	pl := m.state.Playlists[0]
	if got := m.playlistName(pl); got != "youtube.com/watch?v=abc" {
		t.Errorf("an address is shown as %q", got)
	}
	m.state.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "https://www.youtube.com/watch?v=abc", Entries: []engine.EntryInfo{{ID: 0, Playlist: 0, Index: 1, VideoID: "abc", Title: "Me at the zoo"}}})
	if got := m.playlistName(pl); got != "Me at the zoo" {
		t.Errorf("a single video is named %q", got)
	}
}

func TestAnUnreadableLinkCountsAsFailedInTheHeader(t *testing.T) {
	m := sampleModel(t) // one song failed and one link could not be read
	counters, _ := m.headerTexts()
	if !strings.Contains(ansi.Strip(counters), "Failed 2") {
		t.Errorf("the header says %q", ansi.Strip(counters))
	}
}
