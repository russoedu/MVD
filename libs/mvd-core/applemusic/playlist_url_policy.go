package applemusic

import (
	"net/url"
	"regexp"
	"strings"
)

var playlistID = regexp.MustCompile(`^pl\.[A-Za-z0-9]{8,64}$`)

// PlaylistURL returns the address of the public playlist a link points at, for
// https://music.apple.com/<country>/playlist/<name>/pl.<id>; the query and the
// fragment are dropped. The name part of the path may be missing.
func PlaylistURL(link string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(link))
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() != "music.apple.com" {
		return "", false
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i, part := range parts {
		if part == "playlist" && i+1 < len(parts) && playlistID.MatchString(parts[len(parts)-1]) {
			return "https://music.apple.com/" + strings.Join(parts, "/"), true
		}
	}
	return "", false
}
