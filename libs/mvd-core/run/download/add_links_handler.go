package download

import (
	"youtube-downloader/libs/mvd-core/run/downloadsession"
	"fmt"
	"strings"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/outline"
	addlinkspage "youtube-downloader/libs/mvd-core/run/addlinks"
	"youtube-downloader/libs/mvd-core/screenview"
	"youtube-downloader/libs/mvd-core/wording"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"

)

// SourceAdder is what a run offers when it can take more links while it works.
type SourceAdder = downloadsession.SourceAdder

// canAdd reports whether the run behind the screen takes more links.
func (m Model) canAdd() bool {
	_, ok := m.session.Ctrl.(SourceAdder)
	return ok
}

// startAdding opens the box for pasting links.
func (m Model) startAdding() (tea.Model, tea.Cmd) {
	if !m.canAdd() {
		m.flash("this run cannot take more links")
		return m, nil
	}
	ta := textarea.New()
	ta.Placeholder = "Paste or type playlist / video URLs, one per line"
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetWidth(m.addBoxWidth())
	ta.SetHeight(8)
	ta.Focus()
	m.session.Adding, m.links = true, ta
	return m, textarea.Blink
}

// updateAdding handles a key while the box is open: ctrl+s adds the links,
// esc closes the box, anything else is typed into it.
func (m Model) updateAdding(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.session.Adding = false
		return m, nil
	case "ctrl+s":
		return m.addLinks()
	}
	var cmd tea.Cmd
	m.links, cmd = m.links.Update(msg)
	return m, cmd
}

// addLinks queues every line of the box on the run and closes the box.
func (m Model) addLinks() (tea.Model, tea.Cmd) {
	adder, ok := m.session.Ctrl.(SourceAdder)
	if !ok {
		m.session.Adding = false
		return m, nil
	}
	added, lines := 0, 0
	for _, line := range strings.Split(m.links.Value(), "\n") {
		url := strings.TrimSpace(line)
		if url == "" {
			continue
		}
		lines++
		if _, queued := adder.AddSource(url); queued {
			added++
		}
	}
	m.session.Adding = false
	switch {
	case lines == 0:
		m.flash("no links to add")
	case added < lines:
		m.flash(fmt.Sprintf("added %d of %d links; the run has ended", added, lines))
	default:
		m.flash(fmt.Sprintf("added %d link%s", added, wording.Plural(added)))
		m.session.Follow = true
	}
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
		Text:      m.links.Value(),
		HasText:   true,
		TextLabel: "Links to add, one per line",
		Keys:      outline.KeysOf(addLinksHints()),
	}
}
