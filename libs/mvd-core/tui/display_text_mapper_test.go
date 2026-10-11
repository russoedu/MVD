package tui

import (
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
)

func TestEntryTitleFallback(t *testing.T) {
	mk := func(id, title string) *runstate.Entry {
		return &runstate.Entry{EntryInfo: engine.EntryInfo{VideoID: id, Title: title}}
	}
	if got := entryTitle(mk("abc", "")); got != "(abc)" {
		t.Errorf("empty title should fall back to the id, got %q", got)
	}
	if got := entryTitle(mk("abc", "🔊")); got != "(abc)" {
		t.Errorf("title made only of emoji should fall back to the id, got %q", got)
	}
	if got := entryTitle(mk("abc", "Song")); got != "Song" {
		t.Errorf("got %q", got)
	}
}
