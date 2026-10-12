package songs

import (
	"strconv"
	"time"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/outline"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/songfile"
	"youtube-downloader/libs/mvd-core/wording"
)

type Outcome int

const (
	None Outcome = iota
	Added
	Cancel
)

// Model is the screen for pasting songs: one "Artist - Title" per line, or
// a CSV exported from a music service. Saving stores the list as a file that
// the download list then names, so any service can feed the app.
type Model struct {
	ta            textarea.Model
	dir           string // where the lists are stored
	width, height int
	saved         string // the file the last save wrote
	err           string
}

func New(dir string) Model {
	ta := textarea.New()
	ta.Placeholder = "Paste or type your songs, one per line: Artist - Title\n(a first line like  # My road trip  names the list;\na CSV with title and artist columns works too)"
	ta.ShowLineNumbers = true
	ta.CharLimit = 0
	ta.Focus()
	return Model{ta: ta, dir: dir}
}

func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	m.ta.SetWidth(max(10, w-2))
	m.ta.SetHeight(max(3, h-5))
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, Outcome) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "ctrl+s":
			path, err := songfile.Save(m.dir, m.ta.Value(), time.Now())
			if err != nil {
				m.err = err.Error()
				return m, nil, None
			}
			m.saved, m.err = path, ""
			return m, nil, Added
		case "esc":
			return m, nil, Cancel
		}
		m.err = ""
	}
	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
	return m, cmd, None
}

func (m Model) View(width, height int) string {
	props := Default()
	props.Palette = Palette(look.CurrentPalette())
	props.Hints = keybar.Of[HintsHint](m.Hints())
	props.SongsView = m.ta.View()
	if m.err != "" {
		props.ShowSummary, props.ShowProblem = false, true
		props.Problem = " " + m.err
	} else {
		props.ShowSummary, props.ShowProblem = true, false
		props.Summary = " " + m.summaryText()
	}
	return View(props, width, height).View
}

func (m Model) title() string { return "MVD · Add songs" }

// summaryText tells how the text reads: how many songs, and how many lines no
// song could be read from. An error from the last save replaces it.
func (m Model) summaryText() string {
	if m.err != "" {
		return m.err
	}
	list := songfile.Parse(m.ta.Value())
	text := strconv.Itoa(len(list.Songs)) + " song" + wording.Plural(len(list.Songs))
	if list.Skipped > 0 {
		text += ", " + strconv.Itoa(list.Skipped) + " line" + wording.Plural(list.Skipped) + " not understood"
	}
	return text
}

func (m Model) Hints() []keybar.Hint {
	return []keybar.Hint{{Key: "ctrl+s", Desc: "add to the list"}, {Key: "esc", Desc: "cancel"}}
}

// Outline describes the screen in plain terms, from the same state and hints the view draws.
func (m Model) Outline() outline.ScreenOutline {
	return outline.ScreenOutline{
		Title:     m.title(),
		Text:      m.ta.Value(),
		HasText:   true,
		TextLabel: "Songs, one Artist - Title per line",
		Prompt:    m.summaryText(),
		Keys:      outline.KeysOf(m.Hints()),
	}
}

// Saved is the file the last save wrote.
func (m Model) Saved() string { return m.saved }

// WithText puts text in the editor, as if it had been typed or pasted.
func (m Model) WithText(text string) Model {
	m.ta.SetValue(text)
	return m
}
