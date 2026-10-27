package workspace

import (
	"fmt"

	"youtube-downloader/libs/mvd-core/outline"
)

// Outline describes the screen in plain terms, for a screen reader: the review or the downloads, with the question being asked or what the last action said.
func (m Model) Outline() outline.ScreenOutline {
	var out outline.ScreenOutline
	switch {
	case m.life.Stopping:
		return outline.ScreenOutline{Title: "MVD · Stopping the downloads"}
	default:
		out = m.screen.Outline()
		if !m.running() {
			n := len(m.session.Sources())
			out.Title = fmt.Sprintf("MVD · Download list (%d items)", n)
			if n == 1 {
				out.Title = "MVD · Download list (1 item)"
			}
		}
	}
	switch {
	case m.asking != askNone:
		out.Prompt = m.input.Prompt
		out.Text, out.HasText, out.TextLabel = m.input.Value(), true, m.input.Prompt
	case out.Prompt == "":
		out.Prompt = m.notice
	}
	return out
}
