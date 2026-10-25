package download

import (
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/songlist"
)

// toneColour is the colour of the current theme that a tone of a status is drawn in.
func toneColour(t songlist.Tone) string {
	c := look.Colours
	switch t {
	case songlist.ToneDim:
		return c.Dim
	case songlist.ToneFocus:
		return c.Focus
	case songlist.ToneHighlight:
		return c.Highlight
	case songlist.ToneSuccess:
		return c.Success
	case songlist.ToneError:
		return c.Error
	}

	return c.Text
}

// toneStyle is the style of the current theme that a tone of a status is drawn in.
func toneStyle(t songlist.Tone) func(...string) string {
	switch t {
	case songlist.ToneDim:
		return look.Dim.Render
	case songlist.ToneFocus:
		return look.Cyan.Render
	case songlist.ToneHighlight:
		return look.Yellow.Render
	case songlist.ToneSuccess:
		return look.Green.Render
	case songlist.ToneError:
		return look.Red.Render
	}

	return look.Text.Render
}
