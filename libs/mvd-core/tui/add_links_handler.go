package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"youtube-downloader/libs/mvd-core/engine"
)

// SourceAdder is what a run offers when it can take more links while it works.
// The engine does; the download screen shows its "add" key only for a run that does.
type SourceAdder interface {
	// AddSource queues another playlist or video; false once the run has ended.
	AddSource(url string) (engine.PlaylistSource, bool)
}

// canAdd reports whether the run behind the screen takes more links.
func (m model) canAdd() bool {
	_, ok := m.ctrl.(SourceAdder)
	return ok
}

// startAdding opens the box for pasting links.
func (m model) startAdding() (tea.Model, tea.Cmd) {
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
	m.adding, m.links = true, ta
	return m, textareaBlink
}

// updateAdding handles a key while the box is open: ctrl+s adds the links,
// esc closes the box, anything else is typed into it.
func (m model) updateAdding(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.adding = false
		return m, nil
	case "ctrl+s":
		return m.addLinks()
	}
	var cmd tea.Cmd
	m.links, cmd = m.links.Update(msg)
	return m, cmd
}

// addLinks queues every line of the box on the run and closes the box.
func (m model) addLinks() (tea.Model, tea.Cmd) {
	adder, ok := m.ctrl.(SourceAdder)
	if !ok {
		m.adding = false
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
	m.adding = false
	switch {
	case lines == 0:
		m.flash("no links to add")
	case added < lines:
		m.flash(fmt.Sprintf("added %d of %d links; the run has ended", added, lines))
	default:
		m.flash(fmt.Sprintf("added %d link%s", added, plural(added)))
		m.follow = true
	}
	return m, nil
}

func (m model) addBoxWidth() int {
	return max(20, min(100, m.width-8))
}

// renderAddBox draws the box for adding links, centred over the screen.
func (m model) renderAddBox() string {
	body := m.links.View() + "\n\n" + keyBar(m.addBoxWidth(), addLinksHints())
	box := renderBox("Add links", "", m.addBoxWidth()+4, 8+4, strings.Split(body, "\n"), true)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box, lipgloss.WithWhitespaceChars(" "))
}

func addLinksHints() []keyHint {
	return []keyHint{{"ctrl+s", "add"}, {"esc", "cancel"}}
}

// addingOutline describes the box while it is open.
func (m model) addingOutline() ScreenOutline {
	return ScreenOutline{
		Title:     "MVD · Add links",
		Text:      m.links.Value(),
		HasText:   true,
		TextLabel: "Links to add, one per line",
		Keys:      outlineKeys(addLinksHints()),
	}
}
