package dialog

import (
	"fmt"
	"strconv"
	"strings"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/look"
)

// rgb reads a #rgb or #rrggbb colour.
func rgb(colour string) (r, g, b float64, ok bool) {
	c := strings.TrimPrefix(strings.TrimSpace(colour), "#")
	if len(c) == 3 {
		c = string([]byte{c[0], c[0], c[1], c[1], c[2], c[2]})
	}
	if len(c) != 6 {
		return 0, 0, 0, false
	}
	n, err := strconv.ParseUint(c, 16, 32)
	if err != nil {
		return 0, 0, 0, false
	}
	return float64(n >> 16 & 0xff), float64(n >> 8 & 0xff), float64(n & 0xff), true
}

// mixColours is the colour that is the share t (0 to 1) of the way from a to b.
func mixColours(a, b string, t float64) string {
	ar, ag, ab, okA := rgb(a)
	br, bg, bb, okB := rgb(b)
	if !okA || !okB {
		return a
	}
	at := func(x, y float64) int { return int(x + (y-x)*t + 0.5) }
	return fmt.Sprintf("#%02x%02x%02x", at(ar, br), at(ag, bg), at(ab, bb))
}

// readableOn is the colour, black or white, that reads better on ground.
func readableOn(ground string) string {
	if config.Contrast("#000000", ground) >= config.Contrast("#ffffff", ground) {
		return "#000000"
	}
	return "#ffffff"
}

// pillColours are the ground and the text colour of a button of a kind, from the colours of the
// theme: the ground of the screen with some of the text colour mixed in, so it is a lighter
// grey on a dark theme and a darker one on a light theme. The main button is in the accent
// colour, in black or white text, whichever reads better. A button that cannot be used is dim
// but readable. Under the pointer, each is a step brighter.
func pillColours(kind PillKind, hover bool) (ground, text string) {
	base := look.Colours.Background
	if config.Contrast(base, "#000000") == 0 { // the theme has none: the terminal's own, which is dark
		base = "#000000"
	}
	switch kind {
	case PillPrimary:
		ground = look.Colours.Accent
		if hover {
			ground = mixColours(ground, "#ffffff", 0.3)
		}
		return ground, readableOn(ground)
	case PillDisabled:
		ground = groundFor(base, look.Colours.Text, 0.08)
		text = look.Colours.Dim
		if config.Contrast(text, ground) < config.MinTextContrast {
			text = look.Colours.Text
		}
		return ground, text
	}
	share := 0.2
	if hover {
		share = 0.34
	}
	return groundFor(base, look.Colours.Text, share), look.Colours.Text
}

// groundFor is base with the share of text mixed in, taking back as much of the share as the
// text needs to stay readable on it: a theme whose text has little contrast to spare gets a
// ground closer to the screen's own.
func groundFor(base, text string, share float64) string {
	ground := mixColours(base, text, share)
	for share > 0.02 && config.Contrast(text, ground) < config.MinTextContrast {
		share -= 0.02
		ground = mixColours(base, text, share)
	}
	return ground
}
