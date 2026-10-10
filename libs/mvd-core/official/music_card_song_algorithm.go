package official

import (
	"encoding/json"
	"strings"
)

// SongFromWatchPage reads the "Music" card of a watch page, the one that says "1 song"
// under the description with the song's title, its artist and its album, and returns
// the song it names. ok is false when the page has no such card or names several songs
// (a compilation or a mix, where the card cannot say which one this upload is).
func SongFromWatchPage(html []byte) (artist, title string, ok bool) {
	raw := extractJSONAfter(html, "var ytInitialData = ")
	if raw == nil {
		raw = extractJSONAfter(html, `window["ytInitialData"] = `)
	}
	if raw == nil {
		return "", "", false
	}
	var root interface{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return "", "", false
	}

	type song struct{ artist, title string }
	var cards []song
	var walk func(node interface{})
	walk = func(node interface{}) {
		switch n := node.(type) {
		case map[string]interface{}:
			if card, isCard := n["videoAttributeViewModel"].(map[string]interface{}); isCard {
				name, _ := card["title"].(string)
				by, _ := card["subtitle"].(string)
				if name = strings.TrimSpace(name); name != "" {
					cards = append(cards, song{artist: strings.TrimSpace(by), title: name})
				}
			}
			for _, v := range n {
				walk(v)
			}
		case []interface{}:
			for _, v := range n {
				walk(v)
			}
		}
	}
	walk(root)

	// The same card is written more than once in the page: only different songs count.
	var distinct []song
	for _, c := range cards {
		known := false
		for _, d := range distinct {
			known = known || d == c
		}
		if !known {
			distinct = append(distinct, c)
		}
	}
	if len(distinct) != 1 || distinct[0].artist == "" {
		return "", "", false
	}
	return distinct[0].artist, distinct[0].title, true
}
