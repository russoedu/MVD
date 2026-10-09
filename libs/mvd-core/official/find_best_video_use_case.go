package official

import (
	"fmt"
	"strings"
)

// confidentScore is a score that needs no further search: a music video from
// the artist's own channel with a matching title.
const confidentScore = 8.5

// searchQueries are what is searched for, the most precise first.
func searchQueries(song Song) []string {
	title := SearchTitle(song.Title)
	artist := ""
	if len(song.Artists) > 0 {
		artist = song.Artists[0]
	}

	var queries []string
	seen := map[string]bool{}
	for _, query := range []string{
		title + " " + artist + " official video",
		artist + " " + title + " official music video",
		artist + " - " + title,
	} {
		query = strings.Join(strings.Fields(query), " ")
		if query != "" && !seen[query] {
			seen[query] = true
			queries = append(queries, query)
		}
	}
	return queries
}

// FindBestVideo searches YouTube for the video of a song and returns the best
// candidate, and every result seen. It stops at the first search that gives a
// confident music video, and goes on with other wordings otherwise, since the
// official video does not always rank first for one of them. The error is
// that of the first search, when it fails; a later failure only ends the search.
func FindBestVideo(search Searcher, song Song, logf func(format string, a ...interface{})) (Pick, bool, []SearchResult, error) {
	if search == nil {
		return Pick{}, false, nil, fmt.Errorf("no way to search YouTube")
	}

	var results []SearchResult
	for i, query := range searchQueries(song) {
		found, err := search(query)
		if err != nil {
			if i == 0 {
				return Pick{}, false, nil, fmt.Errorf("search for %q failed: %w", query, err)
			}
			logf("search for %q failed: %v", query, err)
			break
		}
		results = append(results, found...)

		if pick, ok := PickBestVideo(results, song); ok && pick.Kind == KindVideo && pick.Score >= confidentScore {
			logf("search for %q: %s", query, describePick(pick))
			return pick, true, results, nil
		}
	}

	ranked := RankCandidates(results, song)
	if len(ranked) == 0 {
		logf("no candidate among %d results", len(results))
		return Pick{}, false, results, nil
	}
	for i, pick := range ranked[:min(len(ranked), 3)] {
		logf("candidate %d: %s", i+1, describePick(pick))
	}
	return ranked[0], true, results, nil
}

func describePick(pick Pick) string {
	return fmt.Sprintf("%s %s by %q (score %.1f: %s)", pick.Kind, pick.ID, pick.Channel, pick.Score, pick.Why)
}
