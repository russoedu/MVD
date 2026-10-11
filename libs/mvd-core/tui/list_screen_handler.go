package tui

import (
	"strconv"
	"strings"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	listpage "youtube-downloader/libs/mvd-core/setup/list"
	"youtube-downloader/libs/mvd-core/wording"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
)

// textareaBlink drives the textarea cursor.
var textareaBlink = textarea.Blink

type listOutcome int

const (
	listNone listOutcome = iota
	listStart
	listPrefs
	listSongs
	listReview
	listQuit
)

// listModel is the download-list editor: a multi-line text area of URLs.
type listModel struct {
	ta            textarea.Model
	width, height int
	confirmQuit   bool
}

func newListModel(urls []string) listModel {
	ta := textarea.New()
	ta.Placeholder = "Paste or type playlist / video URLs, one per line"
	ta.ShowLineNumbers = true
	ta.CharLimit = 0
	ta.SetValue(strings.Join(urls, "\n"))
	ta.Focus()
	return listModel{ta: ta}
}

func (m listModel) setSize(w, h int) listModel {
	m.width, m.height = w, h
	m.ta.SetWidth(max(10, w-2))
	m.ta.SetHeight(max(3, h-4))
	return m
}

// urls returns the trimmed, non-empty lines.
func (m listModel) urls() []string {
	var out []string
	for _, l := range strings.Split(m.ta.Value(), "\n") {
		if t := strings.TrimSpace(l); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// withURL adds a line to the list unless it is already there.
func (m listModel) withURL(url string) listModel {
	for _, existing := range m.urls() {
		if existing == url {
			return m
		}
	}
	value := strings.TrimRight(m.ta.Value(), "\n")
	if value != "" {
		value += "\n"
	}
	m.ta.SetValue(value + url)
	return m
}

func (m listModel) update(msg tea.Msg) (listModel, tea.Cmd, listOutcome) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		if m.confirmQuit {
			switch k.String() {
			case "y", "Y", "enter":
				return m, nil, listQuit
			default:
				m.confirmQuit = false
				return m, nil, listNone
			}
		}
		switch k.String() {
		case "ctrl+s":
			if len(m.urls()) == 0 {
				return m, nil, listNone
			}
			return m, nil, listStart
		case "ctrl+e":
			if len(m.urls()) == 0 {
				return m, nil, listNone
			}
			return m, nil, listReview
		case "ctrl+p":
			return m, nil, listPrefs
		case "ctrl+o":
			return m, nil, listSongs
		case "ctrl+r":
			m.ta.SetValue("")
			return m, nil, listNone
		case "ctrl+q", "esc":
			if len(m.urls()) > 0 {
				m.confirmQuit = true
				return m, nil, listNone
			}
			return m, nil, listQuit
		}
	}
	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
	return m, cmd, listNone
}

func (m listModel) view(width, height int) string {
	n := len(m.urls())
	title := " MVD · Download list"
	if n > 0 {
		title += "  (" + strconv.Itoa(n) + " item" + wording.Plural(n) + ")"
	}

	props := listpage.Default()
	props.Palette = listpage.Palette(look.CurrentPalette())
	props.Title = title
	props.LinksView = m.ta.View()
	props.ShowConfirm, props.ShowKeys = m.confirmQuit, !m.confirmQuit
	props.Hints = keybar.Of[listpage.HintsHint](m.hints())
	props.Confirm = " " + look.Red.Render("Discard the list and quit? ") + look.Key.Render("y") + look.Dim.Render("/") + look.Key.Render("n")
	return listpage.View(props, width, height).View
}

func (m listModel) hints() []keybar.Hint {
	return []keybar.Hint{
		{Key: "ctrl+s", Desc: "start"}, {Key: "ctrl+o", Desc: "songs"}, {Key: "ctrl+e", Desc: "review"}, {Key: "ctrl+p", Desc: "preferences"},
		{Key: "ctrl+r", Desc: "reset"}, {Key: "ctrl+q", Desc: "quit"},
	}
}
