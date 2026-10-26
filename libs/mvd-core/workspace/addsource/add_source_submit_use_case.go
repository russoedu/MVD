package addsource

import (
	"strings"

	"youtube-downloader/libs/mvd-core/musicsearch"
)

// Result is what trying to add comes to: when Done, the addresses to put in the list and the window
// closes; otherwise State holds what is wrong.
type Result struct {
	State     State
	Addresses []string
	Done      bool
}

// Submit tries to add what is written. parseLinks reads the text of the link field into addresses, or
// says why it cannot (the renderer's rules for what a link is).
func (s State) Submit(parseLinks func(text string) (links []string, reason string)) Result {
	if s.Mode == ModeSearch {
		q := musicsearch.Query{Artist: s.Artist, Album: s.Album, Song: s.Song}
		if problem := q.Problem(); problem != "" {
			s.Problem = problem
			s.Focus = FieldArtist
			if strings.TrimSpace(s.Artist) != "" {
				s.Focus = FieldAlbum
			}

			return Result{State: s}
		}

		return Result{State: s, Addresses: []string{q.Address()}, Done: true}
	}
	if strings.TrimSpace(s.URLs) == "" {
		s.Problem, s.Focus = "Paste a link", FieldURL

		return Result{State: s}
	}
	links, reason := parseLinks(s.URLs)
	if reason != "" || len(links) == 0 {
		if reason == "" {
			reason = "Paste a link"
		}
		s.Problem, s.Focus = reason, FieldURL

		return Result{State: s}
	}

	return Result{State: s, Addresses: links, Done: true}
}
