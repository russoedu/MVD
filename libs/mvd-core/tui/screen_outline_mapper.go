package tui

import (
	"youtube-downloader/libs/mvd-core/outline"
	"youtube-downloader/libs/mvd-core/settings"
)

// Outline describes the current screen. It reads the same state and the same
// key hints the views draw from, so the two cannot drift apart.
func (m setupModel) Outline() ScreenOutline {
	switch m.screen {
	case screenConfig:
		return m.config.outline()
	case screenAdvanced:
		return m.advanced.Outline()
	case screenColours:
		return m.colours.Outline()
	case screenSongs:
		return m.songs.Outline()
	default:
		return m.list.Outline()
	}
}

func (m configModel) outline() ScreenOutline {
	if m.mode == settings.EditFolder {
		return m.folder.Outline()
	}
	items := make([]OutlineItem, len(cfgLabels))
	for i, label := range cfgLabels {
		items[i] = OutlineItem{Label: label, Value: m.display(i), Selected: i == m.cursor}
	}
	if m.mode == settings.EditPicking {
		return ScreenOutline{Title: "MVD · Preferences", ItemsLabel: "Settings", Items: items, Prompt: "Choose a folder in the window that opened", Keys: outline.KeysOf(m.hints())}
	}
	if m.mode == settings.EditConfirming {
		return ScreenOutline{Title: "MVD · Preferences", ItemsLabel: "Settings", Items: items, Prompt: m.status, Keys: outline.KeysOf(m.hints())}
	}
	if m.mode != settings.EditNone {
		item := &items[m.cursor]
		item.Editing = true
		if m.mode == settings.EditRadio {
			item.Choices, item.Chosen = m.radioOpts, m.radioIdx
			item.Value = m.radioOpts[m.radioIdx]
		} else {
			item.Value = m.input.Value()
		}
	}
	return ScreenOutline{Title: "MVD · Preferences", ItemsLabel: "Settings", Items: items, Prompt: m.status, Keys: outline.KeysOf(m.hints())}
}
