package folder

import (
	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/outline"
	"youtube-downloader/libs/mvd-core/setup/folderbrowser"
	"youtube-downloader/libs/mvd-core/viewport"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// Outcome is what the host does next after a key press on the picker.
type Outcome int

const (
	None Outcome = iota
	Chosen
	Cancel
)

// Model is a directory navigator: browse into/out of folders, make a
// new one, and pick the current directory. What it does is the browser's;
// the model turns keys into the browser's intents and draws it.
type Model struct {
	browser       folderbrowser.Browser
	input         textinput.Model
	width, height int
}

// New returns a picker that starts in the folder start.
func New(start string) Model {
	ti := textinput.New()
	ti.Prompt = "new folder: "

	return Model{browser: folderbrowser.New(start, nil), input: ti}
}

// ResolveDir is the absolute folder the picker begins in.
func ResolveDir(start string) string { return folderbrowser.ResolveDir(start) }

// SetSize tells the screen that it has w columns and h rows. It also fits the new folder box to the width.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	m.input.SetWidth(max(10, w-16))
	return m
}

// Update handles msg, which is usually a key press, and returns the new model and a command.
// It also returns the outcome, which says when a folder was chosen or the picker was cancelled.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, Outcome) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		if m.browser.Creating {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd, None
		}
		return m, nil, None
	}

	if m.browser.Creating {
		switch k.String() {
		case "esc":
			m.browser.CancelCreating()
		case "enter":
			m.browser.SetNewName(m.input.Value())
			m.browser.CreateFolder()
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd, None
		}
		return m, nil, None
	}

	switch k.String() {
	case "up", "k":
		m.browser.Up()
	case "down", "j":
		m.browser.Down()
	case "right", "l":
		m.browser.Open()
	case "left", "h":
		m.browser.Parent()
	case "n":
		m.browser.StartCreating()
		m.input.SetValue("")
		m.input.Focus()
	case "enter":
		return m, nil, Chosen
	case "esc":
		return m, nil, Cancel
	}
	return m, nil, None
}

// Hints returns the key hints for the current state, browsing or naming a new folder.
func (m Model) Hints() []keybar.Hint {
	if m.browser.Creating {
		return []keybar.Hint{{Key: "enter", Desc: "create"}, {Key: "esc", Desc: "cancel"}}
	}
	return []keybar.Hint{
		{Key: "↑↓", Desc: "move"}, {Key: "→", Desc: "open"}, {Key: "←", Desc: "up"},
		{Key: "n", Desc: "new folder"}, {Key: "enter", Desc: "choose this folder"}, {Key: "esc", Desc: "cancel"},
	}
}

// View draws the picker in width columns and height rows and returns it as text.
func (m Model) View(width, height int) string {
	b := m.browser
	var rows []FoldersRow
	for i, name := range b.Entries {
		row := FoldersRow{Entry: "  " + name + "/"}
		if i == b.Cursor {
			row.Entry, row.Style = "▸ "+name+"/", RowSelected
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		rows = []FoldersRow{{Entry: "(no sub-folders)", EntryStyle: CellStyle{Fg: look.Colours.Dim}}}
	}
	visible := max(1, height-5)
	if len(rows) > visible {
		start := viewport.Start(len(rows), b.Cursor, visible)
		rows = rows[start : start+visible]
	}

	props := Default()
	props.Palette = Palette(look.CurrentPalette())
	props.Hints = keybar.Of[HintsHint](m.Hints())
	props.Path = cells.Truncate(b.Dir, width-2)
	props.HasProblem, props.NoProblem = b.Err != "", b.Err == ""
	props.Problem = cells.Truncate(b.Err, width-2)
	props.Folders = rows
	props.FolderRows, props.ProblemRows = len(rows), 0
	if b.Err != "" {
		props.ProblemRows = 1
	}
	props.Creating = b.Creating
	props.NewFolderView = m.input.View()
	return View(props, width, height).View
}

// Outline describes the screen in plain terms, from the same state and Hints the view draws.
func (m Model) Outline() outline.ScreenOutline {
	items := make([]outline.Item, len(m.browser.Entries))
	for i, name := range m.browser.Entries {
		items[i] = outline.Item{Label: name, Selected: i == m.browser.Cursor}
	}
	out := outline.ScreenOutline{Title: "MVD · Choose folder", Text: m.browser.Dir, HasText: true, ItemsLabel: "Sub-folders", Items: items, Keys: outline.KeysOf(m.Hints())}
	if m.browser.Err != "" {
		out.Prompt = m.browser.Err
	}
	return out
}

// Dir is the folder being browsed.
func (m Model) Dir() string { return m.browser.Dir }

// Listing is a browser showing a listing that was not read from disk: for previews and
// tests, where the folders must not depend on the machine.
func (m Model) Listing(dir string, entries []string, cursor int, err string, creating bool) Model {
	m.browser = m.browser.Listing(dir, entries, cursor, err, creating)
	return m
}
