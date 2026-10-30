package songs

import (
	"time"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/outline"
	"youtube-downloader/libs/mvd-core/setup/songpaste"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
)

// Outcome is where the screen leads; the host acts on it (see songpaste).
type Outcome = songpaste.Outcome

const (
	None   = songpaste.None
	Added  = songpaste.Added
	Cancel = songpaste.Cancel
)

// Model is the screen for pasting songs: one "Artist - Title" per line, or
// a CSV exported from a music service. What it does is songpaste's; the
// model turns keys into its intents and draws it.
type Model struct {
	ta            textarea.Model
	paste         songpaste.Paste
	width, height int
}

// New returns the screen for pasting songs, which saves the song file in the folder dir.
func New(dir string) Model {
	ta := textarea.New()
	ta.Placeholder = "Paste or type your songs, one per line: Artist - Title\n(a first line like  # My road trip  names the list;\na CSV with title and artist columns works too)"
	ta.ShowLineNumbers = true
	ta.CharLimit = 0
	ta.Focus()
	return Model{ta: ta, paste: songpaste.New(dir)}
}

// SetSize tells the screen that it has w columns and h rows. It also fits the text area to them.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	m.ta.SetWidth(max(10, w-2))
	m.ta.SetHeight(max(3, h-5))
	return m
}

// Update handles msg, which is usually a key press, and returns the new model and a command.
// It also returns the outcome, which says when the songs were added or the screen was cancelled.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, Outcome) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "ctrl+s":
			return m, nil, m.paste.Save(time.Now())
		case "esc":
			return m, nil, m.paste.Leave()
		}
		m.paste.ClearError()
	}
	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
	m.paste.SetText(m.ta.Value())
	return m, cmd, None
}

// View draws the screen, with a summary of the songs or the problem found in width columns and height rows and returns it as text.
func (m Model) View(width, height int) string {
	props := Default()
	props.Palette = Palette(look.CurrentPalette())
	props.Hints = keybar.Of[HintsHint](m.Hints())
	props.SongsView = m.ta.View()
	if m.paste.Err != "" {
		props.ShowSummary, props.ShowProblem = false, true
		props.Problem = " " + m.paste.Err
	} else {
		props.ShowSummary, props.ShowProblem = true, false
		props.Summary = " " + m.paste.Summary()
	}
	return View(props, width, height).View
}

// Hints returns the key hints for the screen.
func (m Model) Hints() []keybar.Hint {
	return []keybar.Hint{{Key: "ctrl+s", Desc: "add to the list"}, {Key: "esc", Desc: "cancel"}}
}

// Outline describes the screen in plain terms, from the same state and hints the view draws.
func (m Model) Outline() outline.ScreenOutline {
	return outline.ScreenOutline{
		Title:     songpaste.Heading,
		Text:      m.paste.Text,
		HasText:   true,
		TextLabel: songpaste.TextLabel,
		Prompt:    m.paste.Summary(),
		Keys:      outline.KeysOf(m.Hints()),
	}
}

// Saved is the file the last save wrote.
func (m Model) Saved() string { return m.paste.Saved }

// WithText puts text in the editor, as if it had been typed or pasted.
func (m Model) WithText(text string) Model {
	m.ta.SetValue(text)
	m.paste.SetText(text)
	return m
}
