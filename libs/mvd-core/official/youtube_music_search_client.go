package official

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// videosFilter is the search filter of YouTube Music that lists videos only.
const videosFilter = "EgWKAQIQAWoMEA4QChADEAQQCRAF"

// SearchVideos searches the videos of YouTube Music, whose results carry its
// own tag: OMV for a video the artist's channel uploaded, UGC for anyone
// else's. The channel of a result is the artist the entry names.
func (c *YouTubeMusicClient) SearchVideos(query string) ([]SearchResult, error) {
	body, err := json.Marshal(map[string]interface{}{
		"context": map[string]interface{}{
			"client": map[string]interface{}{"clientName": "WEB_REMIX", "clientVersion": "1.20240925.01.00", "hl": "en"},
		},
		"query":  query,
		"params": videosFilter,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.SearchURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://music.youtube.com")
	req.Header.Set("Referer", "https://music.youtube.com/")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("YouTube Music answered " + resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	return parseVideoResults(data)
}

// parseVideoResults reads the videos out of a YouTube Music search answer.
func parseVideoResults(data []byte) ([]SearchResult, error) {
	var answer interface{}
	if err := json.Unmarshal(data, &answer); err != nil {
		return nil, err
	}

	var results []SearchResult
	collectVideos(answer, &results)
	return results, nil
}

func collectVideos(node interface{}, out *[]SearchResult) {
	switch value := node.(type) {
	case []interface{}:
		for _, item := range value {
			collectVideos(item, out)
		}
	case map[string]interface{}:
		if row, ok := value["musicResponsiveListItemRenderer"].(map[string]interface{}); ok {
			if result, ok := videoRow(row); ok {
				*out = append(*out, result)
			}
			return
		}
		for _, item := range value {
			collectVideos(item, out)
		}
	}
}

// videoRow reads one row of the results: its title, and "Artist • 327K views • 4:08".
func videoRow(row map[string]interface{}) (SearchResult, bool) {
	raw, err := json.Marshal(row)
	if err != nil {
		return SearchResult{}, false
	}
	idMatch := videoIDField.FindSubmatch(raw)
	if idMatch == nil {
		return SearchResult{}, false
	}

	columns, _ := row["flexColumns"].([]interface{})
	column := func(i int) string {
		if i >= len(columns) {
			return ""
		}
		cell, _ := columns[i].(map[string]interface{})
		renderer, _ := cell["musicResponsiveListItemFlexColumnRenderer"].(map[string]interface{})
		return panelText(renderer["text"])
	}

	result := SearchResult{ID: string(idMatch[1]), Title: column(0)}
	if match := musicVideoType.FindSubmatch(raw); match != nil {
		result.MusicType = string(match[1])
	}
	for i, part := range strings.Split(column(1), " • ") {
		part = strings.TrimSpace(part)
		switch {
		case i == 0:
			result.Channel = part
		case looksLikeCount(part):
			result.Views = parseCount(part)
		case parseClock(part) > 0 && strings.Contains(part, ":"):
			result.Duration = parseClock(part)
		}
	}
	if result.Title == "" {
		return SearchResult{}, false
	}
	return result, true
}

// parseCount reads "327K views", "1.2M views" or "540 views" as a number.
func parseCount(text string) int64 {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return 0
	}
	number := strings.ToUpper(fields[0])
	multiplier := 1.0
	switch {
	case strings.HasSuffix(number, "K"):
		multiplier, number = 1e3, strings.TrimSuffix(number, "K")
	case strings.HasSuffix(number, "M"):
		multiplier, number = 1e6, strings.TrimSuffix(number, "M")
	case strings.HasSuffix(number, "B"):
		multiplier, number = 1e9, strings.TrimSuffix(number, "B")
	}
	value, err := strconv.ParseFloat(strings.ReplaceAll(number, ",", ""), 64)
	if err != nil {
		return 0
	}
	return int64(value * multiplier)
}

var videoIDField = regexp.MustCompile(`"videoId":"([A-Za-z0-9_-]{11})"`)
