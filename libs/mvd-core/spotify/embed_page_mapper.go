package spotify

import (
	"encoding/json"
	"errors"
	"regexp"
)

var nextData = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__"[^>]*>(.*?)</script>`)

// embedPage is the part of the embed page's data this package reads.
type embedPage struct {
	Props struct {
		PageProps struct {
			State struct {
				Data struct {
					Entity struct {
						Type      string `json:"type"`
						Title     string `json:"title"`
						TrackList []struct {
							Title    string `json:"title"`
							Subtitle string `json:"subtitle"`
							Duration int    `json:"duration"`
						} `json:"trackList"`
					} `json:"entity"`
				} `json:"data"`
			} `json:"state"`
		} `json:"pageProps"`
	} `json:"props"`
}

// ParseEmbedPage reads the playlist out of the HTML of its embed page.
func ParseEmbedPage(html string) (Playlist, error) {
	match := nextData.FindStringSubmatch(html)
	if match == nil {
		return Playlist{}, errors.New("the Spotify page has no playlist data")
	}

	var page embedPage
	if err := json.Unmarshal([]byte(match[1]), &page); err != nil {
		return Playlist{}, errors.New("cannot read the Spotify playlist data: " + err.Error())
	}

	entity := page.Props.PageProps.State.Data.Entity
	if entity.Type != "playlist" {
		return Playlist{}, errors.New("the link is not a Spotify playlist")
	}

	playlist := Playlist{Title: entity.Title}
	for _, t := range entity.TrackList {
		if t.Title == "" {
			continue
		}
		playlist.Tracks = append(playlist.Tracks, Track{Title: t.Title, Artist: t.Subtitle, DurationMs: t.Duration})
	}
	return playlist, nil
}
