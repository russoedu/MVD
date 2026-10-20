package prefspopup

import (
	"strings"

	"youtube-downloader/libs/mvd-core/outline"
)

// Outline describes the popup in plain terms: the tab shown, and its settings with their
// values, for a screen reader.
func (m Model) Outline() outline.ScreenOutline {
	page := m.pages[m.tab]
	items := make([]outline.Item, len(page.fields))
	for i, f := range page.fields {
		item := outline.Item{Label: f.label, Selected: i == m.cursor}
		if f.get != nil {
			item.Value = f.numberText(f.get(m.cfg))
		}
		switch {
		case m.editing && i == m.cursor:
			item.Editing, item.Value = true, m.input.Value()
		case f.kind == kindChoice:
			item.Choices = make([]string, len(f.options))
			for j, o := range f.options {
				item.Choices[j] = f.optionText(o)
				if o == f.get(m.cfg) {
					item.Chosen = j
				}
			}
		}
		items[i] = item
	}
	names := make([]string, len(m.pages))
	for i, p := range m.pages {
		names[i] = p.name
	}
	return outline.ScreenOutline{
		Title:      "MVD · Preferences",
		ItemsLabel: "Settings, " + page.name + " tab (tabs: " + strings.Join(names, ", ") + ")",
		Items:      items,
		Prompt:     m.status,
		Keys: []outline.Key{
			{Key: "↑↓", Description: "move"}, {Key: "←→", Description: "change"}, {Key: "enter", Description: "edit"},
			{Key: "tab", Description: "next page"}, {Key: "s", Description: "save"}, {Key: "esc", Description: "close"},
		},
	}
}
