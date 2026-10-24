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
	if m.adding {
		return m.addingOutline()
	}
	t := m.state.Tally()
	title := fmt.Sprintf("MVD · Downloads: %d done, %d running, %d queued, %d failed", t.Done+t.Duplicate, t.Running, t.Queued, t.Failed)
	if m.state.Idle {
		title = fmt.Sprintf("MVD · Downloads finished: %d done, %d failed", t.Done+t.Duplicate, t.Failed)
	}
	if m.selPlaylist < len(m.state.Playlists) {
		title += " · " + cells.Display(m.state.Playlists[m.selPlaylist].Title)
	}

	out := outline.ScreenOutline{Title: title, ItemsLabel: "Downloads", Keys: outline.KeysOf(m.hints())}
	if m.confirmQuit {
		out.Prompt = "Downloads are still running. Quit? y/n"
	}

	ids := m.visibleEntries()
	if len(ids) == 0 {
		return out
	}
	sel := viewport.Clamp(m.selEntry, 0, len(ids)-1)
	from := max(0, sel-outlineEntryWindow)
	to := min(len(ids), sel+outlineEntryWindow+1)
	for i := from; i < to; i++ {
		en := m.state.Entries[ids[i]]
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
