package prefsform

import (
	"regexp"
	"strings"
)

// The name pattern of the files is shown as blocks, each a field of the song that cannot
// be changed, with the text the person typed between them. A Unit is one of those: a block
// or a character.

var blockToken = regexp.MustCompile(`%\(([^)]*)\)[0-9]*[sd]`)

// Unit is a block of the pattern, or one character of text between blocks.
type Unit struct {
	Token string // the yt-dlp field for a block; "" for a character
	Label string // what a block says
	R     rune   // the character, for text
}

func (u Unit) IsBlock() bool { return u.Token != "" }

// labelOf is what a block says: the name the palette gives it, else the first field of the
// token in capitals.
func labelOf(token string) string {
	for _, b := range TemplateBlocks {
		if b.Token == token {
			return b.Name
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

// ParseTemplate splits a yt-dlp output template into blocks and characters, without its
// extension.
func ParseTemplate(template string) []Unit {
	var units []Unit
	rest := strings.TrimSuffix(template, extensionSuffix)
	for rest != "" {
		loc := blockToken.FindStringIndex(rest)
		text := rest
		if loc != nil {
			text = rest[:loc[0]]
		}
		for _, r := range text {
			units = append(units, Unit{R: r})
		}
		if loc == nil {
			break
		}
		token := rest[loc[0]:loc[1]]
		units = append(units, Unit{Token: token, Label: labelOf(token)})
		rest = rest[loc[1]:]
	}
	return units
}

// FormatTemplate is the yt-dlp template the units make, ending with the extension.
func FormatTemplate(units []Unit) string {
	var b strings.Builder
	for _, u := range units {
		if u.IsBlock() {
			b.WriteString(u.Token)
		} else {
			b.WriteRune(u.R)
		}
	}
	b.WriteString(extensionSuffix)
	return b.String()
}

// InsertUnits puts add into units at index at.
func InsertUnits(units []Unit, at int, add ...Unit) []Unit {
	at = max(min(at, len(units)), 0)
	out := make([]Unit, 0, len(units)+len(add))
	out = append(out, units[:at]...)
	out = append(out, add...)
	return append(out, units[at:]...)
}

// RemoveUnit takes out Unit i.
func RemoveUnit(units []Unit, i int) []Unit {
	if i < 0 || i >= len(units) {
		return units
	}
	return append(append([]Unit(nil), units[:i]...), units[i+1:]...)
}

// MoveUnit takes Unit from out and puts it where the boundary to (counted before it was
// taken out) is.
func MoveUnit(units []Unit, from, to int) []Unit {
	if from < 0 || from >= len(units) || to < 0 || to > len(units) || to == from || to == from+1 {
		return units
	}
	moved := units[from]
	rest := RemoveUnit(units, from)
	if to > from {
		to--
	}
	return InsertUnits(rest, to, moved)
}
