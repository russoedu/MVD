package official

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// YouTubeMusicClient asks YouTube Music what an upload is. It is an optional
// signal: the endpoint is not documented, so callers treat every error as "no
// answer".
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

// TrackInfo is what YouTube Music says about an upload. Any field may be empty.
type TrackInfo struct {
	// Type is YouTube Music's tag: "ATV" for an auto-generated art track, "OMV"
	// for a video the artist's channel uploaded, "UGC" for anyone else's.
	Type string
	// Title and Artist are the song's own, free of what a channel name adds
	// ("Kate Bush" for the channel "KateBushMusic").
	Title, Artist, Album string
	DurationSec          int
}

// TrackDescriber returns YouTube Music's description of an upload. The app
// wires YouTubeMusicClient.Describe here; it is optional.
type TrackDescriber func(videoID string) (TrackInfo, error)

var (
	videoIDShape   = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)
	musicVideoType = regexp.MustCompile(`MUSIC_VIDEO_TYPE_([A-Z_]+)`)
)

// Describe asks YouTube Music about a video. It returns an empty TrackInfo and
// no error when the answer does not describe that video.
func (c *YouTubeMusicClient) Describe(videoID string) (TrackInfo, error) {
	if !videoIDShape.MatchString(videoID) {
		return TrackInfo{}, fmt.Errorf("%q is not a video id", videoID)
	}

	body, err := json.Marshal(map[string]interface{}{
		"context": map[string]interface{}{
			"client": map[string]interface{}{"clientName": "WEB_REMIX", "clientVersion": "1.20240925.01.00", "hl": "en"},
		},
		"videoId":     videoID,
		"isAudioOnly": true,
	})
	if err != nil {
		return TrackInfo{}, err
	}

	req, err := http.NewRequest(http.MethodPost, c.NextURL, bytes.NewReader(body))
	if err != nil {
		return TrackInfo{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://music.youtube.com")
	req.Header.Set("Referer", "https://music.youtube.com/")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return TrackInfo{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return TrackInfo{}, errors.New("YouTube Music answered " + resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return TrackInfo{}, err
	}
	return parseTrackInfo(data, videoID)
}

// parseTrackInfo reads the panel entry of a video out of a "next" answer.
func parseTrackInfo(data []byte, videoID string) (TrackInfo, error) {
	var answer interface{}
	if err := json.Unmarshal(data, &answer); err != nil {
		return TrackInfo{}, err
	}

	panel := findPanel(answer, videoID)
	if panel == nil {
		return TrackInfo{}, nil
	}

	info := TrackInfo{
		Title:       panelText(panel["title"]),
		Artist:      panelText(panel["shortBylineText"]),
		DurationSec: parseClock(panelText(panel["lengthText"])),
	}
	// "Kate Bush • Hounds Of Love • 1985": the artist, the album, the year.
	if parts := strings.Split(panelText(panel["longBylineText"]), " • "); len(parts) >= 2 {
		if info.Artist == "" {
			info.Artist = strings.TrimSpace(parts[0])
		}
		if album := strings.TrimSpace(parts[1]); album != "" && !yearToken.MatchString(album) && !looksLikeCount(album) {
			info.Album = album
		}
	}
	if raw, err := json.Marshal(panel); err == nil {
		if match := musicVideoType.FindSubmatch(raw); match != nil {
			info.Type = string(match[1])
		}
	}
	return info, nil
}

// findPanel finds the playlist panel entry whose video is videoID.
func findPanel(node interface{}, videoID string) map[string]interface{} {
	switch value := node.(type) {
	case []interface{}:
		for _, item := range value {
			if found := findPanel(item, videoID); found != nil {
				return found
			}
		}
	case map[string]interface{}:
		if panel, ok := value["playlistPanelVideoRenderer"].(map[string]interface{}); ok && panel["videoId"] == videoID {
			return panel
		}
		for _, item := range value {
			if found := findPanel(item, videoID); found != nil {
				return found
			}
		}
	}
	return nil
}

// panelText joins the runs of a YouTube text object.
func panelText(node interface{}) string {
	object, ok := node.(map[string]interface{})
	if !ok {
		return ""
	}
	runs, _ := object["runs"].([]interface{})
	var text strings.Builder
	for _, run := range runs {
		if piece, ok := run.(map[string]interface{}); ok {
			_, _ = fmt.Fprint(&text, piece["text"])
		}
	}
	return strings.TrimSpace(text.String())
}

// looksLikeCount reports whether a byline part is a view or play count
// ("187M views"), which a video's byline has where a song's has the album.
func looksLikeCount(part string) bool {
	lower := strings.ToLower(part)
	return strings.HasSuffix(lower, " views") || strings.HasSuffix(lower, " plays") || strings.HasSuffix(lower, " likes")
}

// parseClock reads "4:59" or "1:02:03" as seconds, 0 when it is neither.
func parseClock(clock string) int {
	seconds := 0
	for _, part := range strings.Split(clock, ":") {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || value < 0 {
			return 0
		}
		seconds = seconds*60 + value
	}
	return seconds
}
