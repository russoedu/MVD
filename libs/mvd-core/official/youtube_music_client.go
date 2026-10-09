package official

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

// YouTubeMusicClient asks YouTube Music what kind of upload a video is. It is
// an optional signal: the endpoint is not documented, so callers treat every
// error as "no answer".
type YouTubeMusicClient struct {
	HTTP      *http.Client
	NextURL   string // the "next" endpoint of YouTube Music
	UserAgent string
}

// NewYouTubeMusicClient returns a client pointed at the real YouTube Music.
func NewYouTubeMusicClient() *YouTubeMusicClient {
	return &YouTubeMusicClient{
		HTTP:      &http.Client{Timeout: 10 * time.Second},
		NextURL:   "https://music.youtube.com/youtubei/v1/next?prettyPrint=false",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	}
}

var (
	videoIDShape   = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)
	musicVideoType = regexp.MustCompile(`MUSIC_VIDEO_TYPE_([A-Z_]+)`)
)

// VideoType returns YouTube Music's own tag for an upload: "ATV" for an
// auto-generated art track, "OMV" for a video the artist's channel uploaded,
// "UGC" for anyone else's upload, and so on. It returns "" and no error when
// the answer does not name the video or has no tag.
func (c *YouTubeMusicClient) VideoType(videoID string) (string, error) {
	if !videoIDShape.MatchString(videoID) {
		return "", fmt.Errorf("%q is not a video id", videoID)
	}

	body, err := json.Marshal(map[string]interface{}{
		"context": map[string]interface{}{
			"client": map[string]interface{}{"clientName": "WEB_REMIX", "clientVersion": "1.20240925.01.00", "hl": "en"},
		},
		"videoId":     videoID,
		"isAudioOnly": true,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, c.NextURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://music.youtube.com")
	req.Header.Set("Referer", "https://music.youtube.com/")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("YouTube Music answered " + resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", err
	}
	if !bytes.Contains(data, []byte(`"videoId":"`+videoID+`"`)) {
		return "", nil
	}
	if match := musicVideoType.FindSubmatch(data); match != nil {
		return string(match[1]), nil
	}
	return "", nil
}
