package tui

import "youtube-downloader/libs/mvd-core/config"

// currentColours are the interface colours ApplyTheme last set; the screens made from
// Cuppa designs take them as their palette.
var currentColours = config.DefaultThemeColors()

// themePalette has the fields of the Palette every screen of the designs has, so a
// screen's own Palette type is a conversion away.
type themePalette struct {
	Accent, Focus, Highlight, Success, Error, Dim, Text, Selected string
}

// palette returns the current interface colours as the palette of the screens.
func palette() themePalette {
	c := currentColours
	return themePalette{
		Accent: c.Accent, Focus: c.Focus, Highlight: c.Highlight, Success: c.Success,
		Error: c.Error, Dim: c.Dim, Text: c.Text, Selected: c.Selected,
	}
}

// hintsOf turns the key hints of a screen into the typed hints of a key bar.
func hintsOf[T ~struct{ Key, Label string }](hints []keyHint) []T {
	out := make([]T, 0, len(hints))
	for _, h := range hints {
		out = append(out, T(struct{ Key, Label string }{h.key, h.desc}))
	}
	return out
}
