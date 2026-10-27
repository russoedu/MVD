package prefspopup

import (
	"youtube-downloader/libs/mvd-core/prefsform"
	"strings"

	"youtube-downloader/libs/mvd-core/outline"
)

// Outline describes the popup in plain terms: the tab shown, and its settings with their
// values, for a screen reader.
func (m Model) Outline() outline.ScreenOutline {
	page := m.form.Pages[m.form.Tab]
	items := make([]outline.Item, len(page.Fields))
	for i, f := range page.Fields {
		item := outline.Item{Label: f.Label, Selected: i == m.form.Cursor}
		if f.Get != nil {
			item.Value = f.NumberText(f.Get(m.form.Cfg))
		}
		switch {
		case m.form.Editing && i == m.form.Cursor && f.Kind != prefsform.KindTemplate:
			item.Editing, item.Value = true, m.input.Value()
		case m.form.Editing && i == m.form.Cursor:
			item.Editing = true
		case f.Kind == prefsform.KindChoice:
			options := f.OptionsFor(m.form.Cfg)
			item.Choices = make([]string, len(options))
			for j, o := range options {
				item.Choices[j] = f.OptionTextFor(m.form.Cfg, o)
				if o == f.Get(m.form.Cfg) {
					item.Chosen = j
				}
			}
		}
		items[i] = item
	}
	names := make([]string, len(m.form.Pages))
	for i, p := range m.form.Pages {
		names[i] = p.Name
	}
	return outline.ScreenOutline{
		Title:      "MVD · Preferences",
		ItemsLabel: "Settings, " + page.Name + " tab (tabs: " + strings.Join(names, ", ") + ")",
		Items:      items,
		Prompt:     m.form.Status,
		Keys: []outline.Key{
			{Key: "↑↓", Description: "move"}, {Key: "←→", Description: "change"}, {Key: "enter", Description: "edit"},
			{Key: "tab", Description: "next page"}, {Key: "s", Description: "save"}, {Key: "esc", Description: "close"},
		},
	}
}
