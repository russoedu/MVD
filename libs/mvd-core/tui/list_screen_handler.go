package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// textareaBlink drives the textarea cursor.
var textareaBlink = textarea.Blink

type listOutcome int

const (
	listNone listOutcome = iota
	listStart
	listPrefs
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

func (m listModel) update(msg tea.Msg) (listModel, tea.Cmd, listOutcome) {
	if k, ok := msg.(tea.KeyMsg); ok {
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
		case "ctrl+p":
			return m, nil, listPrefs
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
	title := "MVD · Download list"
	if n > 0 {
		title += "  (" + strconv.Itoa(n) + " item" + plural(n) + ")"
	}

	var bar string
	if m.confirmQuit {
		bar = fit(" "+styRed.Render("Discard the list and quit? ")+styKey.Render("y")+styDim.Render("/")+styKey.Render("n"), width)
	} else {
		bar = keyBar(width, []keyHint{
			{"ctrl+s", "start"}, {"ctrl+p", "preferences"},
			{"ctrl+r", "reset"}, {"ctrl+q", "quit"},
		})
	}
	return screenFrame(width, height, title, m.ta.View(), bar)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
