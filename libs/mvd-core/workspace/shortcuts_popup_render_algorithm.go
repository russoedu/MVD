package workspace

import (
	"strings"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/menubar"
)

// screenKey is a key of the download screen that no menu lists, and what it does.
type screenKey struct{ keys, does string }

// downloadKeys are the keys of the screen that shows a run, as its own help lists them.
var downloadKeys = []screenKey{
	{"↑ ↓  j k", "Move the selection, scroll the output"},
	{"Tab", "Move between the lists, the songs and the output"},
	{"Enter", "Follow the download that is running"},
	{"Space", "Tick or untick the song"},
	{"D", "Accept the songs that are ticked (or all, with a capital)"},
	{"X", "Reject the songs that are ticked"},
	{"O", "Open the video of the song"},
	{"S", "Swap the video of the song for another"},
	{"C", "Choose the columns, and their order"},
	{"Shift+F", "Choose which statuses to show"},
	{"F", "Show only the songs that failed"},
	{"R", "Retry the failed song, or all of a list"},
	{"L", "Show the output on the whole screen"},
	{"?", "Help for this screen"},
}

// keyColumn is how wide the column of keys is in the window.
const keyColumn = 14

// shortcutRow is one line of the window: the key, then what it does.
func shortcutRow(keys, does string) string {
	pad := strings.Repeat(" ", max(keyColumn-len([]rune(keys)), 1))
	return "  " + look.Key.Render(keys) + pad + look.Text.Render(does)
}

// shortcutSection is a heading of the Keyboard shortcuts window and its keys.
type shortcutSection struct {
	heading string
	keys    []screenKey
}

// shortcutSections are what the Keyboard shortcuts window lists: every shortcut of every menu,
// read from the menus themselves so the window cannot say anything the menus do not, then the
// keys of the download screen. Two entries that share a key (Download and Pause downloads) are said once.
func shortcutSections() []shortcutSection {
	var sections []shortcutSection
	for _, menu := range menus() {
		var keys []screenKey
		at := map[string]int{}
		for _, e := range menu.Entries {
			if e.Separator || e.Shortcut == "" {
				continue
			}
			label := menubar.ShortcutLabel(e.Shortcut)
			if i, ok := at[label]; ok {
				keys[i].does += " / " + e.Name
				continue
			}
			at[label] = len(keys)
			keys = append(keys, screenKey{label, e.Name})
		}
		if len(keys) > 0 {
			sections = append(sections, shortcutSection{menu.Name, keys})
		}
	}

	return append(sections, shortcutSection{"While downloading", downloadKeys})
}

// shortcutLines are the lines of the Keyboard shortcuts window, drawn for the terminal.
func shortcutLines() []string {
	var lines []string
	for _, section := range shortcutSections() {
		lines = append(lines, "", " "+look.Title.Render(section.heading))
		for _, k := range section.keys {
			lines = append(lines, shortcutRow(k.keys, k.does))
		}
	}
	return lines
}

// shortcutsLayout is the size of the Keyboard shortcuts window for a screen, and the lines of
// it that show with the scroll it has.
type shortcutsLayout struct {
	lines         []string
	width, height int
	maxScroll     int
}

func (m Model) shortcutsLayout() shortcutsLayout {
	all := shortcutLines()
	w := min(max(m.width*55/100, 52), 76)
	w = min(w, max(m.width-2, 4))
	// The window is as tall as the lines, up to the screen; a footer line says how to leave.
	h := min(len(all)+3, max(m.height-2, 6))
	body := h - 3
	maxScroll := max(len(all)-body, 0)
	start := min(max(m.shortcutsScroll, 0), maxScroll)
	shown := make([]string, 0, h-2)
	for _, l := range all[start:min(start+body, len(all))] {
		shown = append(shown, cells.Fit(l, w-2))
	}
	for len(shown) < body {
		shown = append(shown, strings.Repeat(" ", w-2))
	}
	hint := "esc close"
	if maxScroll > 0 {
		hint = "↑↓ scroll  esc close"
	}
	shown = append(shown, cells.Fit(" "+look.Dim.Render(hint), w-2))
	return shortcutsLayout{lines: shown, width: w, height: h, maxScroll: maxScroll}
}

// shortcutsView is the Keyboard shortcuts window.
func (m Model) shortcutsView() string {
	l := m.shortcutsLayout()
	return dialog.ClosableFrame("Keyboard shortcuts", l.lines, l.width, l.height)
}
