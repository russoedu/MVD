package spotify

import (
	"net/url"
	"regexp"
	"strings"
)

var playlistID = regexp.MustCompile(`^[A-Za-z0-9]{10,40}$`)

// PlaylistID returns the id of the playlist a link or URI points at, for
// https://open.spotify.com/playlist/<id> (with or without a language prefix
// or query) and spotify:playlist:<id>.
func PlaylistID(link string) (string, bool) {
	link = strings.TrimSpace(link)
	if rest, ok := strings.CutPrefix(link, "spotify:playlist:"); ok {
		return rest, playlistID.MatchString(rest)
	}

	u, err := url.Parse(link)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() != "open.spotify.com" {
		return "", false
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i, part := range parts {
		if part == "playlist" && i+1 < len(parts) {
			return parts[i+1], playlistID.MatchString(parts[i+1])
		}
	}
	return "", false
}
