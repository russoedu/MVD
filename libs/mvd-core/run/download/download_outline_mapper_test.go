package download

import (
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
)

func TestDownloadOutlineWindowsALongList(t *testing.T) {
	m := sampleModel(t)
	m.focus = paneEntries
	m.selEntry = 4
	out := m.Outline()
	if len(out.Items) != 6 || !out.Items[4].Selected {
		t.Fatalf("items: %+v", out.Items)
	}
	if out.Items[4].Value != "downloading 34%" || out.Items[2].Value != "failed: Video unavailable" {
		t.Errorf("values: %q, %q", out.Items[4].Value, out.Items[2].Value)
	}
	if len(out.Keys) == 0 {
		t.Error("expected the key bar")
	}

	var entries []engine.EntryInfo
	for i := 0; i < 500; i++ {
		entries = append(entries, engine.EntryInfo{ID: 100 + i, Playlist: 0, Index: i + 1, Title: "t"})
	}
	m.state.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "Big", Entries: entries})
	m.selEntry = 250
	if n := len(m.Outline().Items); n != 2*outlineEntryWindow+1 {
		t.Errorf("want a window of %d, got %d", 2*outlineEntryWindow+1, n)
	}
}
