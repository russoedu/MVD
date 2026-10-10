package playlisteditor

import (
	"net/url"
	"regexp"
	"strings"
)

var videoIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

// VideoID finds the YouTube video a person pasted: a watch, short or embed address, a
// youtu.be link, a YouTube Music address, or the bare 11-character id. ok is false when
// the text is none of those.
func VideoID(text string) (id string, ok bool) {
	text = strings.TrimSpace(text)
	if videoIDPattern.MatchString(text) {
		return text, true
	}
	if !strings.Contains(text, "://") {
		text = "https://" + text
	}
	u, err := url.Parse(text)
	if err != nil {
		return "", false
	}
	host := strings.ToLower(strings.TrimPrefix(u.Hostname(), "www."))
	switch host {
	case "youtu.be":
		id = strings.Trim(u.Path, "/")
	case "youtube.com", "m.youtube.com", "music.youtube.com", "youtube-nocookie.com":
		if v := u.Query().Get("v"); v != "" {
			id = v
			break
		}
		for _, prefix := range []string{"/shorts/", "/embed/", "/live/", "/v/"} {
			if rest, found := strings.CutPrefix(u.Path, prefix); found {
				id = strings.Split(rest, "/")[0]
			}
		}
	}
	if videoIDPattern.MatchString(id) {
		return id, true
	}
	return "", false
}

// VideoAddress is the address that opens a video.
func VideoAddress(id string) string { return "https://www.youtube.com/watch?v=" + id }
