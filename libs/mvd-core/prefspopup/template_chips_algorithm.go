package prefspopup

import (
	"regexp"
	"strings"
)

// The name pattern of the files is shown as blocks, each a field of the song that cannot
// be changed, with the text the person typed between them. A unit is one of those: a block
// or a character.

var blockToken = regexp.MustCompile(`%\(([^)]*)\)[0-9]*[sd]`)

// unit is a block of the pattern, or one character of text between blocks.
type unit struct {
	token string // the yt-dlp field for a block; "" for a character
	label string // what a block says
	r     rune   // the character, for text
}

func (u unit) isBlock() bool { return u.token != "" }

// labelOf is what a block says: the name the palette gives it, else the first field of the
// token in capitals.
func labelOf(token string) string {
	for _, b := range templateBlocks {
		if b.token == token {
			return b.name
		}
	}
	parts := blockToken.FindStringSubmatch(token)
	if parts == nil {
		return strings.ToUpper(token)
	}
	name := strings.TrimSpace(strings.Split(parts[1], ",")[0])
	return strings.ToUpper(strings.ReplaceAll(name, "_", " "))
}

// extensionSuffix ends every output template: the file's own extension is not the person's to
// move or take away, so it is kept out of the blocks and added again when the template is made.
const extensionSuffix = ".%(ext)s"

// parseTemplate splits a yt-dlp output template into blocks and characters, without its
// extension.
func parseTemplate(template string) []unit {
	var units []unit
	rest := strings.TrimSuffix(template, extensionSuffix)
	for rest != "" {
		loc := blockToken.FindStringIndex(rest)
		text := rest
		if loc != nil {
			text = rest[:loc[0]]
		}
		for _, r := range text {
			units = append(units, unit{r: r})
		}
		if loc == nil {
			break
		}
		token := rest[loc[0]:loc[1]]
		units = append(units, unit{token: token, label: labelOf(token)})
		rest = rest[loc[1]:]
	}
	return units
}

// formatTemplate is the yt-dlp template the units make, ending with the extension.
func formatTemplate(units []unit) string {
	var b strings.Builder
	for _, u := range units {
		if u.isBlock() {
			b.WriteString(u.token)
		} else {
			b.WriteRune(u.r)
		}
	}
	b.WriteString(extensionSuffix)
	return b.String()
}

// insertUnits puts add into units at index at.
func insertUnits(units []unit, at int, add ...unit) []unit {
	at = max(min(at, len(units)), 0)
	out := make([]unit, 0, len(units)+len(add))
	out = append(out, units[:at]...)
	out = append(out, add...)
	return append(out, units[at:]...)
}

// removeUnit takes out unit i.
func removeUnit(units []unit, i int) []unit {
	if i < 0 || i >= len(units) {
		return units
	}
	return append(append([]unit(nil), units[:i]...), units[i+1:]...)
}

// moveUnit takes unit from out and puts it where the boundary to (counted before it was
// taken out) is.
func moveUnit(units []unit, from, to int) []unit {
	if from < 0 || from >= len(units) || to < 0 || to > len(units) || to == from || to == from+1 {
		return units
	}
	moved := units[from]
	rest := removeUnit(units, from)
	if to > from {
		to--
	}
	return insertUnits(rest, to, moved)
}
