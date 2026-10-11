package look

// Palette has the fields of the Palette every screen of the designs has, so a screen's own
// Palette type is a conversion away.
type Palette struct {
	Accent, Focus, Highlight, Success, Error, Dim, Text, Selected string
}

// CurrentPalette returns the current interface colours as the palette of the screens.
func CurrentPalette() Palette {
	c := Colours
	return Palette{
		Accent: c.Accent, Focus: c.Focus, Highlight: c.Highlight, Success: c.Success,
		Error: c.Error, Dim: c.Dim, Text: c.Text, Selected: c.Selected,
	}
}
