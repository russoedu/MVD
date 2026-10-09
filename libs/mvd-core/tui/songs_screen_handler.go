package tui

import (
	"strconv"
	"time"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/songfile"
)

type songsOutcome int

const (
	songsNone songsOutcome = iota
	songsAdded
	songsCancel
)

// songsModel is the screen for pasting songs: one "Artist - Title" per line, or
// a CSV exported from a music service. Saving stores the list as a file that
// the download list then names, so any service can feed the app.
type songsModel struct {
	ta            textarea.Model
	dir           string // where the lists are stored
	width, height int
	saved         string // the file the last save wrote
	err           string
}

func newSongsModel(dir string) songsModel {
	ta := textarea.New()
	ta.Placeholder = "Paste or type your songs, one per line: Artist - Title\n(a first line like  # My road trip  names the list;\na CSV with title and artist columns works too)"
	ta.ShowLineNumbers = true
	ta.CharLimit = 0
	ta.Focus()
	return songsModel{ta: ta, dir: dir}
}

func (m songsModel) setSize(w, h int) songsModel {
	m.width, m.height = w, h
	m.ta.SetWidth(max(10, w-2))
	m.ta.SetHeight(max(3, h-5))
	return m
}

func (m songsModel) update(msg tea.Msg) (songsModel, tea.Cmd, songsOutcome) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "ctrl+s":
			path, err := songfile.Save(m.dir, m.ta.Value(), time.Now())
			if err != nil {
				m.err = err.Error()
				return m, nil, songsNone
			}
			m.saved, m.err = path, ""
			return m, nil, songsAdded
		case "esc":
			return m, nil, songsCancel
		}
		m.err = ""
	}
	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
	return m, cmd, songsNone
}

func (m songsModel) view(width, height int) string {
	body := m.ta.View() + "\n" + m.summary()
	return screenFrame(width, height, m.title(), body, keyBar(width, m.hints()))
}

func (m songsModel) title() string { return "MVD · Add songs" }

// summary is the line under the text area.
func (m songsModel) summary() string {
	if m.err != "" {
		return " " + styRed.Render(m.err)
	}
	return " " + styDim.Render(m.summaryText())
}

// summaryText tells how the text reads: how many songs, and how many lines no
// song could be read from. An error from the last save replaces it.
func (m songsModel) summaryText() string {
	if m.err != "" {
		return m.err
	}
	list := songfile.Parse(m.ta.Value())
	text := strconv.Itoa(len(list.Songs)) + " song" + plural(len(list.Songs))
	if list.Skipped > 0 {
		text += ", " + strconv.Itoa(list.Skipped) + " line" + plural(list.Skipped) + " not understood"
	}
	return text
}

func (m songsModel) hints() []keyHint {
	return []keyHint{{"ctrl+s", "add to the list"}, {"esc", "cancel"}}
}

func (m songsModel) outline() ScreenOutline {
	return ScreenOutline{
		Title:     m.title(),
		Text:      m.ta.Value(),
		HasText:   true,
		TextLabel: "Songs, one Artist - Title per line",
		Prompt:    m.summaryText(),
		Keys:      outlineKeys(m.hints()),
	}
}
