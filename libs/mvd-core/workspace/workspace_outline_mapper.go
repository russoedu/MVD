package workspace

import (
	"fmt"

	"youtube-downloader/libs/mvd-core/outline"
)

// Outline describes the screen in plain terms, for a screen reader: the review, the edit
// of the list or the downloads, with the question being asked or what the last action said.
func (m Model) Outline() outline.ScreenOutline {
	var out outline.ScreenOutline
	switch {
	case m.stopping:
		return outline.ScreenOutline{Title: "MVD · Stopping the downloads"}
	case m.editing:
		out = m.edit.Outline()
	case m.reviewing:
		out = m.editor.Outline()
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

// Outline describes the edit of the list: each address and its songs.
func (p editPopup) Outline() outline.ScreenOutline {
	out := outline.ScreenOutline{
		Title:      fmt.Sprintf("MVD · Edit the list: %d addresses", len(p.items)),
		ItemsLabel: "Addresses and songs",
		Keys: []outline.Key{
			{Key: "a", Description: "add"}, {Key: "enter", Description: "change address"}, {Key: "d", Description: "remove"}, {Key: "c", Description: "copy"}, {Key: "v", Description: "paste"},
			{Key: "K", Description: "move up"}, {Key: "J", Description: "move down"}, {Key: "s", Description: "save"}, {Key: "esc", Description: "cancel"},
		},
	}
	for i, r := range p.rows() {
		label := p.items[r.item].Source
		if r.song >= 0 {
			label = songName(p.items[r.item].Songs[r.song])
		}
		out.Items = append(out.Items, outline.Item{Label: label, Selected: i == p.sel})
	}
	return out
}
