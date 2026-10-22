package look

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// sgr is the escape that sets the foreground (38) or the background (48) to the #rgb or
// #rrggbb colour, or "" when it is not a colour.
func sgr(code int, colour string) string {
	c := strings.TrimPrefix(strings.TrimSpace(colour), "#")
	if len(c) == 3 {
		c = string([]byte{c[0], c[0], c[1], c[1], c[2], c[2]})
	}
	if len(c) != 6 {
		return ""
	}
	n, err := strconv.ParseUint(c, 16, 32)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("\x1b[%d;2;%d;%d;%dm", code, n>>16&0xff, n>>8&0xff, n&0xff)
}

// Paint puts the theme's background, and its text colour for what has none of its own, under
// a drawn screen: every line is made as wide as the widest and begins with the colours, and
// they are set again after every reset in it, so the cells nothing coloured show them. A
// theme with no background leaves the screen as it was.
func Paint(content string) string {
	bg := sgr(48, Colours.Background)
	if bg == "" {
		return content
	}
	fg := sgr(38, Colours.Text)
	start := bg + fg

	lines := strings.Split(content, "\n")
	width := 0
	for _, l := range lines {
		width = max(width, ansi.StringWidth(l))
	}
	reset := strings.NewReplacer(
		"\x1b[0m", "\x1b[0m"+start,
		"\x1b[m", "\x1b[m"+start,
		"\x1b[49m", "\x1b[49m"+bg,
		"\x1b[39m", "\x1b[39m"+fg,
	)
	for i, l := range lines {
		pad := strings.Repeat(" ", max(width-ansi.StringWidth(l), 0))
		lines[i] = start + reset.Replace(l) + start + pad + "\x1b[0m"
	}
	return strings.Join(lines, "\n")
}
