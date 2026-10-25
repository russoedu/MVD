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
