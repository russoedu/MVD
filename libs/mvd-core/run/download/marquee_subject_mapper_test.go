package download

import (
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
)

func TestAMarqueeStartsAgainWhenTheSelectedNameChanges(t *testing.T) {
	m := sampleModel(t)
	next, _ := m.Update(evMsg{ev: engine.EvEntryState{Entry: 0, State: engine.StateDone}})
	m = next.(Model)
	for i := 0; i < 300; i++ {
		next, _ = m.Update(tickMsg{})
		m = next.(Model)
	}
	if m.marquee < 300 {
		t.Fatalf("the marquee did not run: %d", m.marquee)
	}

	next, _ = m.Update(evMsg{ev: engine.EvEntryState{Entry: 1, State: engine.StateDone}})
	if got := next.(Model).marquee; got < 300 {
		t.Errorf("an event that leaves the text alone restarted the marquee: %d", got)
	}

	next, _ = m.Update(evMsg{ev: engine.EvPlaylistListed{Playlist: 0, Title: "A much longer name that arrived after the first", Entries: []engine.EntryInfo{
		{ID: 0, Playlist: 0, Index: 1, VideoID: "rnlp_avexYQ", Title: "x", Channel: "y"},
	}}})
	if got := next.(Model).marquee; got != 0 {
		t.Errorf("a new name did not start from its beginning: %d", got)
	}
}
