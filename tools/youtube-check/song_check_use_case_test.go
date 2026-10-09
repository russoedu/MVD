package main

import (
	"testing"

	"youtube-downloader/libs/mvd-core/ytdlp"
)

func TestEverySongIsLookedUpExceptOnesThatAreAlreadyTheOfficialVideo(t *testing.T) {
	entries := []ytdlp.PlaylistEntry{{ID: "a", Title: "Song A"}, {ID: "b", Title: "Song B (Official Video)"}, {ID: "c", Title: "Song C"}}
	baseline := Baseline{Songs: []Song{{ID: "a", Official: "va"}, {ID: "b"}}}
	lookup := songLookup{
		Wanted: func(title string) bool { return title != "Song B (Official Video)" },
		Find:   func(entry ytdlp.PlaylistEntry) (string, string) { return "v" + entry.ID, "found" },
	}

	outcomes := checkSongs(entries, baseline, lookup, 3)

	if len(outcomes) != 3 || outcomes[0].ID != "a" || outcomes[2].ID != "c" {
		t.Fatalf("the answers should follow the playlist: %+v", outcomes)
	}
	if !outcomes[0].Known || outcomes[0].Expected != "va" || outcomes[0].Found != "va" {
		t.Errorf("a recorded song carries its recorded answer: %+v", outcomes[0])
	}
	if outcomes[1].Found != "" || outcomes[1].Why == "" {
		t.Errorf("an official video is not looked up: %+v", outcomes[1])
	}
	if outcomes[2].Known {
		t.Errorf("a song not in the baseline is new: %+v", outcomes[2])
	}
}
