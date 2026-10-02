package ytdlp

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var templateFieldPattern = regexp.MustCompile(`%\(([A-Za-z0-9_,|]+)\)([-+ 0#]*\d*)([sd])`)

// ApplyPlaylistFields pre-renders the playlist related fields of a yt-dlp
// output template. When a single video is downloaded on its own, yt-dlp
// has no playlist context, so %(playlist_title)s and friends would become
// "NA". We substitute them with what the flat playlist listing told us, so
// files land in the same place they would with a normal playlist download.
func ApplyPlaylistFields(template string, e PlaylistEntry) string {
	return templateFieldPattern.ReplaceAllStringFunc(template, func(match string) string {
		sub := templateFieldPattern.FindStringSubmatch(match)
		expr, flags, verb := sub[1], sub[2], sub[3]

		defaultValue := ""
		hasDefault := false
		if i := strings.Index(expr, "|"); i >= 0 {
			defaultValue = expr[i+1:]
			hasDefault = true
			expr = expr[:i]
		}

		var strVal string
		var intVal int
		isInt := false
		found := false

		for _, field := range strings.Split(expr, ",") {
			field = strings.TrimSpace(field)
			switch field {
			case "playlist_title":
				strVal = e.PlaylistTitle
			case "playlist":
				strVal = e.Playlist
				if strVal == "" {
					strVal = e.PlaylistTitle
				}
			case "playlist_id":
				strVal = e.PlaylistID
			case "playlist_index", "playlist_autonumber":
				intVal, isInt = e.PlaylistIndex, true
			case "playlist_count":
				intVal, isInt = e.PlaylistCount, true
			default:
				// Not a playlist field: leave the whole placeholder to yt-dlp.
				return match
			}
			if (isInt && intVal != 0) || (!isInt && strVal != "") {
				found = true
				break
			}
			isInt = false
		}

		if !found {
			if hasDefault {
				return defaultValue
			}
			return "NA"
		}

		if isInt {
			if verb == "d" {
				return fmt.Sprintf("%"+flags+"d", intVal)
			}
			return fmt.Sprintf("%"+flags+"s", strconv.Itoa(intVal))
		}
		if verb == "d" {
			return match
		}
		return fmt.Sprintf("%"+flags+"s", sanitizeFilename(strVal))
	})
}

// sanitizeFilename mirrors yt-dlp's default (non restricted) filename
// sanitisation so that folder names match between both download modes.
func sanitizeFilename(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '/':
			b.WriteRune('\u29F8') // ⧸
		case '\\':
			b.WriteRune('\u29F9') // ⧹
		case '"', '*', ':', '<', '>', '?', '|':
			b.WriteRune(r + 0xFEE0) // full-width variant
		case '\n', '\r', '\t':
			b.WriteRune(' ')
		default:
			if r < 32 || r == 127 {
				continue
			}
			b.WriteRune(r)
		}
	}
	out := strings.TrimRight(b.String(), ". ")
	out = strings.TrimSpace(out)
	if out == "" {
		return "_"
	}
	return out
}
