package applemusic

import (
	"encoding/json"
	"errors"
	"html"
	"regexp"
	"sort"
	"strings"
)

var (
	serverData   = regexp.MustCompile(`(?s)<script[^>]*id="serialized-server-data"[^>]*>(.*?)</script>`)
	playlistData = regexp.MustCompile(`(?s)<script[^>]*schema:music-playlist[^>]*>(.*?)</script>`)
)

// ParsePlaylistPage reads the playlist out of the HTML of its web page: the
// tracks from the page's server data, the name from its playlist metadata.
func ParsePlaylistPage(page string) (Playlist, error) {
	match := serverData.FindStringSubmatch(page)
	if match == nil {
		return Playlist{}, errors.New("the Apple Music page has no playlist data")
	}

	var data interface{}
	if err := json.Unmarshal([]byte(match[1]), &data); err != nil {
		return Playlist{}, errors.New("cannot read the Apple Music playlist data: " + err.Error())
	}

	playlist := Playlist{Title: playlistName(page)}
	seen := map[string]bool{}
	collectTracks(data, seen, &playlist.Tracks)
	if len(playlist.Tracks) == 0 {
		return Playlist{}, errors.New("the Apple Music page lists no songs (is the playlist public?)")
	}
	return playlist, nil
}

// playlistName reads the name from the page's schema.org playlist data.
func playlistName(page string) string {
	if match := playlistData.FindStringSubmatch(page); match != nil {
		var meta struct {
			Name string `json:"name"`
		}
		if json.Unmarshal([]byte(match[1]), &meta) == nil && meta.Name != "" {
			return html.UnescapeString(meta.Name)
		}
	}
	return "Apple Music playlist"
}

// collectTracks walks the page data in order and gathers every song row: an
// object whose id starts with "track-lockup" and that has a title and an
// artist. A song shown twice on the page is taken once.
func collectTracks(node interface{}, seen map[string]bool, out *[]Track) {
	switch value := node.(type) {
	case []interface{}:
		for _, item := range value {
			collectTracks(item, seen, out)
		}
	case map[string]interface{}:
		if track, id, ok := songRow(value); ok {
			if !seen[id] {
				seen[id] = true
				*out = append(*out, track)
			}
			return
		}
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			collectTracks(value[key], seen, out)
		}
	}
}

func songRow(object map[string]interface{}) (Track, string, bool) {
	id, _ := object["id"].(string)
	title, _ := object["title"].(string)
	artist, _ := object["artistName"].(string)
	if !strings.HasPrefix(id, "track-lockup") || title == "" {
		return Track{}, "", false
	}
	duration, _ := object["duration"].(float64)
	return Track{Title: title, Artist: artist, DurationMs: int(duration)}, id, true
}
