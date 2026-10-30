package list

import (
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/outline"
	"youtube-downloader/libs/mvd-core/setup/urllist"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
)

// Outcome is where the screen leads; the host acts on it (see urllist).
type Outcome = urllist.Outcome

const (
	None   = urllist.None
	Start  = urllist.Start
	Prefs  = urllist.Prefs
	Songs  = urllist.Songs
	Review = urllist.Review
	Quit   = urllist.Quit
)

// Model is the download-list editor: a multi-line text area of URLs. What the list does is
// urllist's; the model turns keys into its intents and draws it.
type Model struct {
	ta            textarea.Model
	list          urllist.List
	width, height int
}

// New returns the editor holding the lines of URLs.
func New(URLs []string) Model {
	ta := textarea.New()
	ta.Placeholder = "Paste or type playlist / video URLs, one per line"
	ta.ShowLineNumbers = true
	ta.CharLimit = 0
	list := urllist.New(URLs)
	ta.SetValue(list.Text)
	ta.Focus()
	return Model{ta: ta, list: list}
}

// SetSize tells the screen that it has w columns and h rows. It also fits the text area to them.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	m.ta.SetWidth(max(10, w-2))
	m.ta.SetHeight(max(3, h-4))
	return m
}

// URLs returns the trimmed, non-empty lines.
func (m Model) URLs() []string { return m.list.URLs() }

// WithURL adds a line to the list unless it is already there.
func (m Model) WithURL(url string) Model {
	m.list.AddURL(url)
	m.ta.SetValue(m.list.Text)
	return m
}

// Update handles msg, which is usually a key press, and returns the new model and a command.
// It also returns the outcome, which says where the screen leads.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, Outcome) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		if m.list.ConfirmQuit {
			switch k.String() {
			case "y", "Y", "enter":
				return m, nil, m.list.Answer(true)
			default:
				return m, nil, m.list.Answer(false)
			}
		}
		switch k.String() {
		case "ctrl+s":
			return m, nil, m.list.Begin()
		case "ctrl+e":
			return m, nil, m.list.ReviewLinks()
		case "ctrl+p":
			return m, nil, m.list.OpenPreferences()
		case "ctrl+o":
			return m, nil, m.list.OpenSongs()
		case "ctrl+r":
			m.list.Reset()
			m.ta.SetValue(m.list.Text)
			return m, nil, None
		case "ctrl+q", "esc":
			return m, nil, m.list.Leave()
		}
	}
	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
	m.list.SetText(m.ta.Value())
	return m, cmd, None
}

// View draws the editor in width columns and height rows and returns it as text.
func (m Model) View(width, height int) string {
	title := " " + urllist.Heading
	if count := m.list.Count(); count != "" {
		title += "  " + count
	}

	props := Default()
	props.Palette = Palette(look.CurrentPalette())
	props.Title = title
	props.LinksView = m.ta.View()
	props.ShowConfirm, props.ShowKeys = m.list.ConfirmQuit, !m.list.ConfirmQuit
	props.Hints = keybar.Of[HintsHint](m.Hints())
	props.Confirm = " " + look.Red.Render(urllist.QuitQuestion+" ") + look.Key.Render("y") + look.Dim.Render("/") + look.Key.Render("n")
	return View(props, width, height).View
}

// Hints returns the key hints for the editor.
func (m Model) Hints() []keybar.Hint {
	return []keybar.Hint{
		{Key: "ctrl+s", Desc: "start"}, {Key: "ctrl+o", Desc: "songs"}, {Key: "ctrl+e", Desc: "review"}, {Key: "ctrl+p", Desc: "preferences"},
		{Key: "ctrl+r", Desc: "reset"}, {Key: "ctrl+q", Desc: "quit"},
	}
}

// Outline describes the screen in plain terms, from the same state and Hints the view draws.
func (m Model) Outline() outline.ScreenOutline {
	title := urllist.Heading
	if count := m.list.Count(); count != "" {
		title += " " + count
	}
	out := outline.ScreenOutline{Title: title, Text: m.list.Text, HasText: true, Keys: outline.KeysOf(m.Hints())}
	if m.list.ConfirmQuit {
		out.Prompt = urllist.QuitQuestion + " y/n"
	}
	return out
}
