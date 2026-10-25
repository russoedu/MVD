package download

import (
	"fmt"
	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/outline"
	"youtube-downloader/libs/mvd-core/viewport"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
)

// outlineEntryWindow bounds how many entries an outline lists, around the
// selected one, so a list of thousands stays cheap to describe.
const outlineEntryWindow = 25

// Outline describes the download screen: the totals and playlist in the title,
// the selected playlist's entries (a window around the selection) as items.
func (m Model) Outline() outline.ScreenOutline {
	if m.session.Adding {
		return m.addingOutline()
	}
	t := m.session.State.Tally()
	title := fmt.Sprintf("MVD · Downloads: %d done, %d running, %d queued, %d failed", t.Done+t.Duplicate, t.Running, t.Queued, t.Failed)
	if m.session.State.Idle {
		title = fmt.Sprintf("MVD · Downloads finished: %d done, %d failed", t.Done+t.Duplicate, t.Failed)
	}
	if m.session.SelPlaylist < len(m.session.State.Playlists) {
		title += " · " + cells.Display(m.session.State.Playlists[m.session.SelPlaylist].Title)
	}

	out := outline.ScreenOutline{Title: title, ItemsLabel: "Downloads", Keys: outline.KeysOf(m.hints())}
	if m.session.ConfirmQuit {
		out.Prompt = "Downloads are still running. Quit? y/n"
	}

	ids := m.visibleEntries()
	if len(ids) == 0 {
		return out
	}
	sel := viewport.Clamp(m.session.SelEntry, 0, len(ids)-1)
	from := max(0, sel-outlineEntryWindow)
	to := min(len(ids), sel+outlineEntryWindow+1)
	for i := from; i < to; i++ {
		en := m.session.State.Entries[ids[i]]
		out.Items = append(out.Items, outline.Item{
			Label:    fmt.Sprintf("%02d %s", en.Index, entryTitle(en)),
			Value:    entryOutlineValue(en),
			Selected: i == sel,
		})
	}
	return out
}

func entryOutlineValue(en *runstate.Entry) string {
	switch en.State {
	case engine.StateDownloading:
		return fmt.Sprintf("downloading %.0f%%", en.Percent)
	case engine.StateFailed:
		if en.Err != "" {
			return "failed: " + runstate.PlainError(en.Err)
		}
	}
	return en.State.String()
}
