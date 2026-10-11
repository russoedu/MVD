package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	folderpage "youtube-downloader/libs/mvd-core/setup/folder"
	"youtube-downloader/libs/mvd-core/viewport"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type folderOutcome int

const (
	folderNone folderOutcome = iota
	folderChosen
	folderCancel
)

// folderModel is a directory navigator: browse into/out of folders, make a
// new one, and pick the current directory.
type folderModel struct {
	dir           string
	entries       []string // sub-directory names
	cursor        int
	creating      bool
	input         textinput.Model
	err           string
	width, height int
}

func newFolderModel(start string) folderModel {
	ti := textinput.New()
	ti.Prompt = "new folder: "
	m := folderModel{dir: resolveDir(start), input: ti}
	m.load()
	return m
}

func resolveDir(start string) string {
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

func (m *folderModel) load() {
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

func (m folderModel) setSize(w, h int) folderModel {
	m.width, m.height = w, h
	m.input.SetWidth(max(10, w-16))
	return m
}

func (m folderModel) update(msg tea.Msg) (folderModel, tea.Cmd, folderOutcome) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		if m.creating {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd, folderNone
		}
		return m, nil, folderNone
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
			return m, cmd, folderNone
		}
		return m, nil, folderNone
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
		return m, nil, folderChosen
	case "esc":
		return m, nil, folderCancel
	}
	return m, nil, folderNone
}

func (m folderModel) hints() []keybar.Hint {
	if m.creating {
		return []keybar.Hint{{Key: "enter", Desc: "create"}, {Key: "esc", Desc: "cancel"}}
	}
	return []keybar.Hint{
		{Key: "↑↓", Desc: "move"}, {Key: "→", Desc: "open"}, {Key: "←", Desc: "up"},
		{Key: "n", Desc: "new folder"}, {Key: "enter", Desc: "choose this folder"}, {Key: "esc", Desc: "cancel"},
	}
}

func (m folderModel) view(width, height int) string {
	var rows []folderpage.FoldersRow
	for i, name := range m.entries {
		row := folderpage.FoldersRow{Entry: "  " + name + "/"}
		if i == m.cursor {
			row.Entry, row.Style = "▸ "+name+"/", folderpage.RowSelected
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		rows = []folderpage.FoldersRow{{Entry: "(no sub-folders)", EntryStyle: folderpage.CellStyle{Fg: look.Colours.Dim}}}
	}
	visible := max(1, height-5)
	if len(rows) > visible {
		start := viewport.Start(len(rows), m.cursor, visible)
		rows = rows[start : start+visible]
	}

	props := folderpage.Default()
	props.Palette = folderpage.Palette(look.CurrentPalette())
	props.Hints = keybar.Of[folderpage.HintsHint](m.hints())
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
	return folderpage.View(props, width, height).View
}
