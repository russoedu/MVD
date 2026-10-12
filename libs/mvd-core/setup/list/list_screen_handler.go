package list

import (
	"youtube-downloader/libs/mvd-core/connect"
	"strconv"
	"strings"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/outline"
	"youtube-downloader/libs/mvd-core/wording"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
)

type Outcome int

const (
	None Outcome = iota
	Start
	Prefs
	Songs
	Review
	Quit
)

// Model is the download-list editor: a multi-line text area of URLs.
type Model struct {
	ta            textarea.Model
	width, height int
	confirmQuit   bool
}

func New(URLs []string) Model {
	ta := textarea.New()
	ta.Placeholder = "Paste or type playlist / video URLs, one per line"
	ta.ShowLineNumbers = true
	ta.CharLimit = 0
	ta.SetValue(strings.Join(URLs, "\n"))
	ta.Focus()
	return Model{ta: ta}
}

func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	m.ta.SetWidth(max(10, w-2))
	m.ta.SetHeight(max(3, h-4))
	return m
}

// URLs returns the trimmed, non-empty lines.
func (m Model) URLs() []string {
	var out []string
	for _, l := range strings.Split(m.ta.Value(), "\n") {
		if t := strings.TrimSpace(l); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// WithURL adds a line to the list unless it is already there.
func (m Model) WithURL(url string) Model {
	for _, existing := range m.URLs() {
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

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, Outcome) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		if m.confirmQuit {
			switch k.String() {
			case "y", "Y", "enter":
				return m, nil, Quit
			default:
				m.confirmQuit = false
				return m, nil, None
			}
		}
		switch k.String() {
		case "ctrl+s":
			if len(m.URLs()) == 0 {
				return m, nil, None
			}
			return m, nil, Start
		case "ctrl+e":
			if len(m.URLs()) == 0 {
				return m, nil, None
			}
			return m, nil, Review
		case "ctrl+p":
			return m, nil, Prefs
		case "ctrl+o":
			return m, nil, Songs
		case "ctrl+r":
			m.ta.SetValue("")
			return m, nil, None
		case "ctrl+q", "esc":
			if len(m.URLs()) > 0 {
				m.confirmQuit = true
				return m, nil, None
			}
			return m, nil, Quit
		}
	}
	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
	return m, cmd, None
}

func (m Model) View(width, height int) string {
	return View(m.props(width, height), width, height).View
}

// Live is the page this screen draws and its inputs, by the input names of the design,
// for the live connection to the designer.
func (m Model) Live(width, height int) (page string, inputs map[string]any) {
	return "List", connect.Plain(Inputs(m.props(width, height)))
}

// props fills the page for a window of width by height cells.
func (m Model) props(width, height int) Props {
	n := len(m.URLs())
	title := " MVD · Download list"
	if n > 0 {
		title += "  (" + strconv.Itoa(n) + " item" + wording.Plural(n) + ")"
	}

	props := Default()
	props.Palette = Palette(look.CurrentPalette())
	props.Title = title
	props.LinksView = m.ta.View()
	props.ShowConfirm, props.ShowKeys = m.confirmQuit, !m.confirmQuit
	props.Hints = keybar.Of[HintsHint](m.Hints())
	props.Confirm = " " + look.Red.Render("Discard the list and quit? ") + look.Key.Render("y") + look.Dim.Render("/") + look.Key.Render("n")
	return props
}

func (m Model) Hints() []keybar.Hint {
	return []keybar.Hint{
		{Key: "ctrl+s", Desc: "start"}, {Key: "ctrl+o", Desc: "songs"}, {Key: "ctrl+e", Desc: "review"}, {Key: "ctrl+p", Desc: "preferences"},
		{Key: "ctrl+r", Desc: "reset"}, {Key: "ctrl+q", Desc: "quit"},
	}
}

// Outline describes the screen in plain terms, from the same state and Hints the view draws.
func (m Model) Outline() outline.ScreenOutline {
	n := len(m.URLs())
	title := "MVD · Download list"
	if n > 0 {
		title += " (" + strconv.Itoa(n) + " item" + wording.Plural(n) + ")"
	}
	out := outline.ScreenOutline{Title: title, Text: m.ta.Value(), HasText: true, Keys: outline.KeysOf(m.Hints())}
	if m.confirmQuit {
		out.Prompt = "Discard the list and quit? y/n"
	}
	return out
}
