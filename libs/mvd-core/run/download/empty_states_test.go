package download

import (
	"strings"
	"testing"
)

func TestTheEmptyPlaylistsPaneSaysWhatToDoFirstHowAndWhatWorks(t *testing.T) {
	var text []string
	for _, r := range emptyPlaylistRows(60) {
		text = append(text, r.Title)
	}
	all := strings.Join(strings.Fields(strings.Join(text, " ")), " ")
	for _, want := range []string{"Add the songs you want as music videos.", "Click + Add link below, press Ctrl+L, or paste links with Ctrl+V.", "YouTube videos and playlists", "Spotify and Apple Music playlists"} {
		if !strings.Contains(all, want) {
			t.Errorf("the pane lacks %q: %q", want, all)
		}
	}
}

func TestTheEntriesPaneSaysWhatIsMissingAndHowToGetIt(t *testing.T) {
	m := sampleModel(t)
	m.state.Playlists = nil
	if got := m.emptyEntriesText(); !strings.Contains(got, "once you add one") {
		t.Errorf("an empty list says %q", got)
	}
	m = sampleModel(t)
	m.editable = true
	if got := m.emptyEntriesText(); !strings.Contains(got, "Start (Ctrl+R)") || !strings.Contains(got, "Look up first (Ctrl+Shift+R)") {
		t.Errorf("a list not looked up says %q", got)
	}
	m.editable = false
	if got := m.emptyEntriesText(); got != "No songs to show here." {
		t.Errorf("a run with no songs says %q", got)
	}
}

func TestTheOutputPaneSaysWhyItIsEmpty(t *testing.T) {
	m := sampleModel(t)
	m.selPlaylist = 2 // the playlist that could not be read
	if got := m.emptyOutputText(); !strings.Contains(got, "could not be read") {
		t.Errorf("a failed playlist says %q", got)
	}
	m.selPlaylist = 0
	m.editable = true
	if got := m.emptyOutputText(); !strings.Contains(got, "shows here while it works") {
		t.Errorf("before a run it says %q", got)
	}
	m.editable = false
	if got := m.emptyOutputText(); got != "Nothing was written for this one." {
		t.Errorf("during a run it says %q", got)
	}
}
