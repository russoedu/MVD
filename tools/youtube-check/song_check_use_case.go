package main

import (
	"sync"

	"youtube-downloader/libs/mvd-core/ytdlp"
)

// songLookup is the official video lookup of the app, for one song.
type songLookup struct {
	// Wanted says whether the app would look the song up at all (one whose title says it
	// is the official video is not).
	Wanted func(title string) bool
	// Find returns the video id found ("" for none) and the lookup's own words about it.
	Find func(entry ytdlp.PlaylistEntry) (found, why string)
}

// checkSongs asks the lookup about every entry of the playlist, a few at a time so as not
// to hammer YouTube, and returns the answers in the order of the playlist.
func checkSongs(entries []ytdlp.PlaylistEntry, baseline Baseline, lookup songLookup, workers int) []Outcome {
	recorded := map[string]Song{}
	for _, song := range baseline.Songs {
		recorded[song.ID] = song
	}

	outcomes := make([]Outcome, len(entries))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < max(1, workers); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				entry := entries[i]
				outcome := Outcome{ID: entry.ID, Title: entry.Title}
				if song, ok := recorded[entry.ID]; ok {
					outcome.Known, outcome.Expected = true, song.Official
				}
				if lookup.Wanted(entry.Title) {
					outcome.Found, outcome.Why = lookup.Find(entry)
				} else {
					outcome.Why = "the title already says it is the official video"
				}
				outcomes[i] = outcome
			}
		}()
	}
	for i := range entries {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	return outcomes
}
