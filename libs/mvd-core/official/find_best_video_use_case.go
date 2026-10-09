package official

import (
	"fmt"
	"strings"
)

// confidentScore is a score that needs no further search, for a video that says
// it is official or that a database lists: from the artist's own channel, with a
// matching title.
const confidentScore = 8.5

// maxChecked is how many of the best candidates are checked for being art tracks.
const maxChecked = 5

// searchQueries are what is searched for on YouTube, the most precise first.
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

// FindBestVideo looks for the video of a song and returns the best candidate,
// and every search result seen. It asks, cheapest and most exact first: the
// cache of earlier runs, a database of known videos (Wikidata), YouTube Music's
// video search, then three wordings of a YouTube search, and stops as soon as
// one gives a confident music video. The error is that of the YouTube search,
// when it fails first and nothing else answered; the other sources only fail
// silently.
func FindBestVideo(search Searcher, sources Sources, song Song, logf func(format string, a ...interface{})) (Pick, bool, []SearchResult, error) {
	if sources.Cache != nil {
		if cached, ok := sources.Cache.Get(song); ok {
			logf("remembered from an earlier run: %s %s by %q", cached.Kind, cached.ID, cached.Channel)
			return Pick{ID: cached.ID, Kind: cached.Kind, Score: confidentScore, Channel: cached.Channel, Sure: true, Why: "remembered (" + cached.Why + ")"}, true, nil, nil
		}
	}

	var results []SearchResult
	checked := map[string]string{}

	// best is the best candidate that is not an art track: the top few are checked
	// with YouTube Music, since an art track of the artist's channel looks like a
	// video of it in a search result.
	best := func() (Pick, bool) {
		for i, pick := range RankCandidates(results, song) {
			if i == maxChecked {
				break
			}
			kind := pick.MusicType
			if kind == "" && sources.Type != nil {
				var known bool
				if kind, known = checked[pick.ID]; !known {
					if tag, err := sources.Type(pick.ID); err == nil {
						kind = tag
					}
					checked[pick.ID] = kind
				}
			}
			if artTrack, _ := ArtTrackFromVideoType(kind); artTrack {
				logf("not %s by %q: an auto-generated track too", pick.ID, pick.Channel)
				continue
			}
			return pick, true
		}
		return Pick{}, false
	}
	confident := func() (Pick, bool) {
		pick, ok := best()
		return pick, ok && pick.Kind == KindVideo && pick.Sure && pick.Score >= confidentScore
	}
	finish := func(pick Pick) (Pick, bool, []SearchResult, error) {
		if sources.Cache != nil {
			sources.Cache.Put(song, pick)
		}
		return pick, true, results, nil
	}

	if sources.Known != nil {
		results = append(results, sources.Known(song, logf)...)
		if pick, ok := confident(); ok {
			logf("known video: %s", describePick(pick))
			return finish(pick)
		}
	}

	if sources.Music != nil {
		query := strings.Join(strings.Fields(SearchTitle(song.Title)+" "+firstArtist(song)), " ")
		if found, err := sources.Music(query); err != nil {
			logf("YouTube Music could not be searched: %v", err)
		} else {
			results = append(results, found...)
			if pick, ok := confident(); ok {
				logf("YouTube Music search for %q: %s", query, describePick(pick))
				return finish(pick)
			}
		}
	}

	if search == nil {
		return Pick{}, false, results, fmt.Errorf("no way to search YouTube")
	}
	for i, query := range searchQueries(song) {
		found, err := search(query)
		if err != nil {
			if i == 0 && len(results) == 0 {
				return Pick{}, false, nil, fmt.Errorf("search for %q failed: %w", query, err)
			}
			logf("search for %q failed: %v", query, err)
			break
		}
		results = append(results, found...)

		if pick, ok := confident(); ok {
			logf("search for %q: %s", query, describePick(pick))
			return finish(pick)
		}
	}

	pick, ok := best()
	if !ok {
		logf("no candidate among %d results", len(results))
		return Pick{}, false, results, nil
	}
	ranked := RankCandidates(results, song)
	for i, candidate := range ranked[:min(len(ranked), 3)] {
		logf("candidate %d: %s", i+1, describePick(candidate))
	}
	return finish(pick)
}

func firstArtist(song Song) string {
	if len(song.Artists) == 0 {
		return ""
	}
	return song.Artists[0]
}

func describePick(pick Pick) string {
	return fmt.Sprintf("%s %s by %q (score %.1f: %s)", pick.Kind, pick.ID, pick.Channel, pick.Score, pick.Why)
}
