package prefsform

import (
	"regexp"
	"strconv"
	"strings"
)

var templateToken = regexp.MustCompile(`%\(([^)]*)\)(0?)(\d*)([sd])`)

// exampleSong is the song the example name is made from.
var exampleSong = map[string]string{
	"title": "Part Time Lover", "artist": "Double You", "uploader": "Double You",
	"playlist_title": "Dance 90's", "playlist": "Dance 90's", "playlist_index": "10",
	"id": "dQw4w9WgXcQ",
}

// ExampleName is what a file would be called by a yt-dlp output template, for a sample
// song, so the person sees the effect of the pattern. A field with alternatives
// (a,b) takes the first one the sample has; one it has none of reads NA.
func ExampleName(template, ext string) string {
	return templateToken.ReplaceAllStringFunc(template, func(token string) string {
		parts := templateToken.FindStringSubmatch(token)
		value := "NA"
		for _, name := range strings.Split(parts[1], ",") {
			name = strings.TrimSpace(name)
			if name == "ext" {
				value = ext
				break
			}
			if v, ok := exampleSong[name]; ok {
				value = v
				break
			}
		}
		if parts[4] == "d" {
			width, _ := strconv.Atoi(parts[3])
			if parts[2] == "0" && len(value) < width {
				value = strings.Repeat("0", width-len(value)) + value
			}
		}
		return value
	})
}

// TemplateBlock is one of the blocks that add a field to the template.
type TemplateBlock struct{ Name, Token string }

// TemplateBlocks lists the blocks the person can add to the file name pattern, with their yt-dlp fields.
var TemplateBlocks = []TemplateBlock{
	{"ARTIST", "%(artist,uploader)s"},
	{"TITLE", "%(title)s"},
	{"PLAYLIST NAME", "%(playlist_title,playlist)s"},
	{"PLAYLIST MUSIC INDEX", "%(playlist_index)02d"},
}
