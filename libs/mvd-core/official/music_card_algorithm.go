package official

import (
	"bytes"
	"encoding/json"
	"regexp"
)

var videoIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

// youtubeLinkPattern finds watch links inside free text (description).
var youtubeLinkPattern = regexp.MustCompile(`(?:youtube\.com/watch\?(?:[^\s"&]*&)?v=|youtu\.be/|youtube\.com/shorts/)([A-Za-z0-9_-]{11})`)

// extractJSONAfter finds `marker` in data and decodes the single JSON value
// that follows it. It returns nil when the marker is absent or the value is
// not valid JSON.
func extractJSONAfter(data []byte, marker string) []byte {
	idx := bytes.Index(data, []byte(marker))
	if idx < 0 {
		return nil
	}
	rest := data[idx+len(marker):]
	dec := json.NewDecoder(bytes.NewReader(rest))
	var v json.RawMessage
	if err := dec.Decode(&v); err != nil {
		return nil
	}
	return v
}

// candidatesFromWatchPage extracts official-video candidates from the raw
// HTML of a watch page, most trustworthy first.
func candidatesFromWatchPage(html []byte, selfID string) []string {
	var out []string
	if raw := extractJSONAfter(html, "var ytInitialData = "); raw != nil {
		out = append(out, candidatesFromInitialData(raw, selfID)...)
	} else if raw := extractJSONAfter(html, `window["ytInitialData"] = `); raw != nil {
		out = append(out, candidatesFromInitialData(raw, selfID)...)
	}
	if raw := extractJSONAfter(html, "var ytInitialPlayerResponse = "); raw != nil {
		var pr struct {
			VideoDetails struct {
				ShortDescription string `json:"shortDescription"`
			} `json:"videoDetails"`
		}
		if json.Unmarshal(raw, &pr) == nil {
			out = append(out, linksInText(pr.VideoDetails.ShortDescription, selfID)...)
		}
	}
	return uniqueStrings(out)
}

// candidatesFromInitialData walks a ytInitialData / innertube "next"
// response. Priority:
//  1. video ids inside the "Music" card of the description panel
//  2. any other video linked from the structured description panel
//  3. watch links written as plain text in the description
func candidatesFromInitialData(raw []byte, selfID string) []string {
	var root interface{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil
	}

	var music, panel, text []string

	var walk func(node interface{}, inMusic, inPanel bool)
	walk = func(node interface{}, inMusic, inPanel bool) {
		switch n := node.(type) {
		case map[string]interface{}:
			if id, ok := n["panelIdentifier"].(string); ok && id == "engagement-panel-structured-description" {
				inPanel = true
			}
			if inPanel {
				if s, ok := n["videoId"].(string); ok && videoIDPattern.MatchString(s) && s != selfID {
					if inMusic {
						music = append(music, s)
					} else {
						panel = append(panel, s)
					}
				}
				for _, key := range []string{"content", "simpleText", "text"} {
					if s, ok := n[key].(string); ok {
						text = append(text, linksInText(s, selfID)...)
					}
				}
			}
			for k, v := range n {
				walk(v, inMusic || isMusicSectionKey(k), inPanel)
			}
		case []interface{}:
			for _, v := range n {
				walk(v, inMusic, inPanel)
			}
		}
	}
	walk(root, false, false)

	return uniqueStrings(append(append(music, panel...), text...))
}

// isMusicSectionKey matches the renderers YouTube has used for the
// "Music" card at the bottom of a description.
func isMusicSectionKey(key string) bool {
	switch key {
	case "videoDescriptionMusicSectionRenderer",
		"videoAttributeViewModel",
		"horizontalCardListRenderer",
		"carouselLockupRenderer":
		return true
	}
	return false
}

func linksInText(text, selfID string) []string {
	var out []string
	for _, m := range youtubeLinkPattern.FindAllStringSubmatch(text, -1) {
		if m[1] != selfID {
			out = append(out, m[1])
		}
	}
	return out
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]bool, len(in))
	var out []string
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}
