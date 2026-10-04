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
//
// An entry with no playlist context at all is a video that was given on its own.
// Playlist placeholders with no default then render as nothing rather than "NA",
// and the folder and "07 - " style prefix they leave behind are dropped, so the file
// is named after the video and sits directly in the output folder. The template must
// be relative: it is joined to the output folder afterwards.
func ApplyPlaylistFields(template string, e PlaylistEntry) string {
	standalone := e.PlaylistTitle == "" && e.Playlist == "" && e.PlaylistID == "" &&
		e.PlaylistIndex == 0 && e.PlaylistCount == 0
	dropped := false

	rendered := templateFieldPattern.ReplaceAllStringFunc(template, func(match string) string {
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
			if standalone {
				dropped = true

				return ""
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

	if dropped {
		return tidyDroppedFields(rendered)
	}

	return rendered
}

// tidyDroppedFields removes what dropped playlist placeholders leave behind: empty
// folder names, and separators ("07 - ") at the start of the file name. Folders are
// joined with "/", which yt-dlp and the operating system both accept.
func tidyDroppedFields(path string) string {
	segments := strings.FieldsFunc(path, func(r rune) bool { return r == '/' || r == '\\' })
	if len(segments) == 0 {
		return "%(title)s.%(ext)s"
	}

	last := len(segments) - 1
	segments[last] = strings.TrimLeft(segments[last], " -_")
	if segments[last] == "" {
		segments[last] = "%(title)s.%(ext)s"
	}

	return strings.Join(segments, "/")
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
