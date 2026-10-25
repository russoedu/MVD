package download

import (
	"youtube-downloader/libs/mvd-core/run/downloadsession"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/outline"
	addlinkspage "youtube-downloader/libs/mvd-core/run/addlinks"
	"youtube-downloader/libs/mvd-core/screenview"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"

)

// SourceAdder is what a run offers when it can take more links while it works.
type SourceAdder = downloadsession.SourceAdder

// startAdding opens the box for pasting links.
func (m Model) startAdding() (tea.Model, tea.Cmd) {
	m.session.BeginAdding()
	if !m.session.Adding {
		return m, nil
	}
	ta := textarea.New()
	ta.Placeholder = "Paste or type playlist / video URLs, one per line"
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetWidth(m.addBoxWidth())
	ta.SetHeight(8)
	ta.Focus()
	m.links = ta
	return m, textarea.Blink
}

// updateAdding handles a key while the box is open: ctrl+s adds the links,
// esc closes the box, anything else is typed into it.
func (m Model) updateAdding(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.session.CancelAdding()
		return m, nil
	case "ctrl+s":
		return m.addLinks()
	}
	var cmd tea.Cmd
	m.links, cmd = m.links.Update(msg)
	m.session.SetLinksText(m.links.Value())
	return m, cmd
}

// addLinks queues every line of the box on the run and closes the box.
func (m Model) addLinks() (tea.Model, tea.Cmd) {
	m.session.SetLinksText(m.links.Value())
	m.session.CommitLinks()
	return m, nil
}

func (m Model) addBoxWidth() int {
	return max(20, min(100, m.width-8))
}

// renderAddBox draws the box for adding links, centred over the screen by its design.
func (m Model) renderAddBox() string {
	props := addlinkspage.Default()
	props.Palette = addlinkspage.Palette(look.CurrentPalette())
	props.Hints = keybar.Of[addlinkspage.HintsHint](addLinksHints())
	props.LinksView = m.links.View()
	return screenview.Padded(addlinkspage.View(props, m.width, m.height).View, m.width)
}

func addLinksHints() []keybar.Hint {
	return []keybar.Hint{{Key: "ctrl+s", Desc: "add"}, {Key: "esc", Desc: "cancel"}}
}

// addingOutline describes the box while it is open.
func (m Model) addingOutline() outline.ScreenOutline {
	return outline.ScreenOutline{
		Title:     "MVD · Add links",
		Text:      m.session.LinksText,
		HasText:   true,
		TextLabel: "Links to add, one per line",
		Keys:      outline.KeysOf(addLinksHints()),
	}
}
