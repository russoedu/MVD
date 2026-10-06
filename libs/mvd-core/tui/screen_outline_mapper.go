package tui

import "strconv"

// Outline describes the current screen. It reads the same state and the same
// key hints the views draw from, so the two cannot drift apart.
func (m setupModel) Outline() ScreenOutline {
	switch m.screen {
	case screenConfig:
		return m.config.outline()
	case screenAdvanced:
		return m.advanced.outline()
	default:
		return m.list.outline()
	}
}

func outlineKeys(hints []keyHint) []OutlineKey {
	keys := make([]OutlineKey, len(hints))
	for i, h := range hints {
		keys[i] = OutlineKey{Key: h.key, Description: h.desc}
	}
	return keys
}

func (m listModel) outline() ScreenOutline {
	n := len(m.urls())
	title := "MVD · Download list"
	if n > 0 {
		title += " (" + strconv.Itoa(n) + " item" + plural(n) + ")"
	}
	out := ScreenOutline{Title: title, Text: m.ta.Value(), HasText: true, Keys: outlineKeys(m.hints())}
	if m.confirmQuit {
		out.Prompt = "Discard the list and quit? y/n"
	}
	return out
}

func (m configModel) outline() ScreenOutline {
	if m.mode == editFolder {
		return m.folder.outline()
	}
	items := make([]OutlineItem, len(cfgLabels))
	for i, label := range cfgLabels {
		items[i] = OutlineItem{Label: label, Value: m.display(i), Selected: i == m.cursor}
	}
	if m.mode == editPicking {
		return ScreenOutline{Title: "MVD · Preferences", ItemsLabel: "Settings", Items: items, Prompt: "Choose a folder in the window that opened", Keys: outlineKeys(m.hints())}
	}
	if m.mode != editNone {
		item := &items[m.cursor]
		item.Editing = true
		if m.mode == editRadio {
			item.Choices, item.Chosen = m.radioOpts, m.radioIdx
			item.Value = m.radioOpts[m.radioIdx]
		} else {
			item.Value = m.input.Value()
		}
	}
	return ScreenOutline{Title: "MVD · Preferences", ItemsLabel: "Settings", Items: items, Keys: outlineKeys(m.hints())}
}

func (m advancedModel) outline() ScreenOutline {
	items := make([]OutlineItem, len(advLabels))
	for i, label := range advLabels {
		items[i] = OutlineItem{Label: label, Value: m.display(i), Selected: i == m.cursor}
	}
	if m.mode != editNone {
		items[m.cursor].Editing = true
		items[m.cursor].Value = m.input.Value()
	}
	return ScreenOutline{Title: "MVD · Advanced", ItemsLabel: "Settings", Items: items, Keys: outlineKeys(m.hints())}
}

func (m folderModel) outline() ScreenOutline {
	items := make([]OutlineItem, len(m.entries))
	for i, name := range m.entries {
		items[i] = OutlineItem{Label: name, Selected: i == m.cursor}
	}
	out := ScreenOutline{Title: "MVD · Choose folder", Text: m.dir, HasText: true, ItemsLabel: "Sub-folders", Items: items, Keys: outlineKeys(m.hints())}
	if m.err != "" {
		out.Prompt = m.err
	}
	return out
}
