package folder

import (
	"youtube-downloader/libs/mvd-core/connect"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/outline"
	"youtube-downloader/libs/mvd-core/viewport"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type Outcome int

const (
	None Outcome = iota
	Chosen
	Cancel
)

// Model is a directory navigator: browse into/out of folders, make a
// new one, and pick the current directory.
type Model struct {
	dir           string
	entries       []string // sub-directory names
	cursor        int
	creating      bool
	input         textinput.Model
	err           string
	width, height int
}

func New(start string) Model {
	ti := textinput.New()
	ti.Prompt = "new folder: "
	m := Model{dir: ResolveDir(start), input: ti}
	m.load()
	return m
}

func ResolveDir(start string) string {
	if start == "" {
		if wd, err := os.Getwd(); err == nil {
			return wd
		}
		return "."
	}
	if abs, err := filepath.Abs(start); err == nil {
		return abs
	}
	return start
}

func (m *Model) load() {
	m.entries = nil
	m.cursor = 0
	m.err = ""
	ents, err := os.ReadDir(m.dir)
	if err != nil {
		m.err = err.Error()
		return
	}
	for _, e := range ents {
		if e.IsDir() {
			m.entries = append(m.entries, e.Name())
		}
	}
	sort.Strings(m.entries)
}

func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	m.input.SetWidth(max(10, w-16))
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, Outcome) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		if m.creating {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd, None
		}
		return m, nil, None
	}

	if m.creating {
		switch k.String() {
		case "esc":
			m.creating = false
		case "enter":
			name := strings.TrimSpace(m.input.Value())
			if name != "" {
				if err := os.Mkdir(filepath.Join(m.dir, name), 0o755); err != nil {
					m.err = err.Error()
				} else {
					m.dir = filepath.Join(m.dir, name)
					m.load()
				}
			}
			m.creating = false
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd, None
		}
		return m, nil, None
	}

	switch k.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.entries)-1 {
			m.cursor++
		}
	case "right", "l":
		if m.cursor < len(m.entries) {
			m.dir = filepath.Join(m.dir, m.entries[m.cursor])
			m.load()
		}
	case "left", "h":
		if parent := filepath.Dir(m.dir); parent != m.dir {
			m.dir = parent
			m.load()
		}
	case "n":
		m.creating = true
		m.input.SetValue("")
		m.input.Focus()
	case "enter":
		return m, nil, Chosen
	case "esc":
		return m, nil, Cancel
	}
	return m, nil, None
}

func (m Model) Hints() []keybar.Hint {
	if m.creating {
		return []keybar.Hint{{Key: "enter", Desc: "create"}, {Key: "esc", Desc: "cancel"}}
	}
	return []keybar.Hint{
		{Key: "↑↓", Desc: "move"}, {Key: "→", Desc: "open"}, {Key: "←", Desc: "up"},
		{Key: "n", Desc: "new folder"}, {Key: "enter", Desc: "choose this folder"}, {Key: "esc", Desc: "cancel"},
	}
}

func (m Model) View(width, height int) string {
	return View(m.props(width, height), width, height).View
}

// Live is the page this screen draws and its inputs, by the input names of the design,
// for the live connection to the designer.
func (m Model) Live(width, height int) (page string, inputs map[string]any) {
	return "Folder", connect.Plain(Inputs(m.props(width, height)))
}

// props fills the page for a window of width by height cells.
func (m Model) props(width, height int) Props {
	var rows []FoldersRow
	for i, name := range m.entries {
		row := FoldersRow{Entry: "  " + name + "/"}
		if i == m.cursor {
			row.Entry, row.Style = "▸ "+name+"/", RowSelected
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		rows = []FoldersRow{{Entry: "(no sub-folders)", EntryStyle: CellStyle{Fg: look.Colours.Dim}}}
	}
	visible := max(1, height-5)
	if len(rows) > visible {
		start := viewport.Start(len(rows), m.cursor, visible)
		rows = rows[start : start+visible]
	}

	props := Default()
	props.Palette = Palette(look.CurrentPalette())
	props.Hints = keybar.Of[HintsHint](m.Hints())
	props.Path = cells.Truncate(m.dir, width-2)
	props.HasProblem, props.NoProblem = m.err != "", m.err == ""
	props.Problem = cells.Truncate(m.err, width-2)
	props.Folders = rows
	props.FolderRows, props.ProblemRows = len(rows), 0
	if m.err != "" {
		props.ProblemRows = 1
	}
	props.Creating = m.creating
	props.NewFolderView = m.input.View()
	return props
}

// Outline describes the screen in plain terms, from the same state and Hints the view draws.
func (m Model) Outline() outline.ScreenOutline {
	items := make([]outline.Item, len(m.entries))
	for i, name := range m.entries {
		items[i] = outline.Item{Label: name, Selected: i == m.cursor}
	}
	out := outline.ScreenOutline{Title: "MVD · Choose folder", Text: m.dir, HasText: true, ItemsLabel: "Sub-folders", Items: items, Keys: outline.KeysOf(m.Hints())}
	if m.err != "" {
		out.Prompt = m.err
	}
	return out
}

// Dir is the folder being browsed.
func (m Model) Dir() string { return m.dir }

// Listing is a browser showing a listing that was not read from disk: for previews and
// tests, where the folders must not depend on the machine.
func (m Model) Listing(dir string, entries []string, cursor int, err string, creating bool) Model {
	m.dir, m.entries, m.cursor, m.err, m.creating = dir, entries, cursor, err, creating
	return m
}
