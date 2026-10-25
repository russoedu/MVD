package prefspopup

import (
	"youtube-downloader/libs/mvd-core/prefsform"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/look"
)

// chipGrey is the colour of every block of the template, in the pattern and in the palette.
const chipGrey = "#3a3a3a"

var (
	chipStyle  = lipgloss.NewStyle().Background(lipgloss.Color(chipGrey)).Foreground(lipgloss.Color("#e5e7eb"))
	chipGrip   = lipgloss.NewStyle().Background(lipgloss.Color(chipGrey)).Foreground(lipgloss.Color("#9ca3af"))
	chipRemove = lipgloss.NewStyle().Background(lipgloss.Color(chipGrey)).Foreground(lipgloss.Color("#ff4d4d"))
	chipHover  = lipgloss.NewStyle().Background(lipgloss.Color("#5a5a5a")).Foreground(lipgloss.Color("#ffffff")).Underline(true)
	caretStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffe600")).Bold(true)
)

// chipPad is the cells of a block around its name: a space, the grip, a space before the
// name; a space, the cross, a space after it.
const chipLeft, chipRight = 3, 3

// boundary is a place between two units, where the cursor or a block being carried can go.
type boundary struct{ y, x, index int }

// chips draws the pattern as blocks and text, wrapped to w columns, and says where each part
// takes clicks. The lines are counted from the first of them.
func (m Model) chips(field, w int) (lines []string, hits []hit, bounds []boundary) {
	units := prefsform.ParseTemplate(m.form.Cfg.OutputTemplate)
	here := field == m.form.Cursor
	var line strings.Builder
	x, y := gutterWidth, 0
	flush := func() {
		lines = append(lines, line.String())
		line.Reset()
		x, y = gutterWidth, y+1
	}
	mark := func(i int) {
		bounds = append(bounds, boundary{y: y, x: x, index: i})
		if (m.form.Editing && here && m.form.Caret == i && !m.dragging) || (m.dragging && m.dragTo == i) {
			line.WriteString(caretStyle.Render("▎"))
			x++
		}
	}
	if len(units) == 0 {
		mark(0)
		line.WriteString(look.Dim.Render("Click here and type, or add blocks below"))
		hits = append(hits, hit{y: 0, x0: gutterWidth, x1: w, kind: hitCaret, field: field, arg: 0})
		line.WriteString(look.Dim.Render(" .ext"))
		return append(lines, line.String()), hits, bounds
	}
	for i, u := range units {
		need := 1
		if u.IsBlock() {
			need = width(u.Label) + chipLeft + chipRight
		}
		if x+need+1 > w && x > gutterWidth {
			hits = append(hits, hit{y: y, x0: x, x1: w, kind: hitCaret, field: field, arg: i})
			flush()
		}
		mark(i)
		if u.IsBlock() {
			bw := width(u.Label) + chipLeft + chipRight
			line.WriteString(chipStyle.Render(" ") + chipGrip.Render("≡") + chipStyle.Render(" "+u.Label+" ") + chipRemove.Render("×") + chipStyle.Render(" "))
			hits = append(hits,
				hit{y: y, x0: x, x1: x + 3, kind: hitChipHandle, field: field, arg: i},
				hit{y: y, x0: x + 3, x1: x + bw - 3, kind: hitCaret, field: field, arg: i + 1},
				hit{y: y, x0: x + bw - 3, x1: x + bw, kind: hitChipRemove, field: field, arg: i})
			x += bw
		} else {
			line.WriteString(look.Text.Render(string(u.R)))
			hits = append(hits, hit{y: y, x0: x, x1: x + 1, kind: hitCaret, field: field, arg: i})
			x++
		}
	}
	mark(len(units))
	hits = append(hits, hit{y: y, x0: x, x1: w, kind: hitCaret, field: field, arg: len(units)})
	line.WriteString(look.Dim.Render(".ext")) // the file's own extension, always last
	return append(lines, line.String()), hits, bounds
}

// palette draws the blocks that can be added to the pattern, as grey buttons.
func (m Model) palette(field, y int) (string, []hit) {
	var row strings.Builder
	var hits []hit
	x := gutterWidth
	for j, b := range prefsform.TemplateBlocks {
		label := " + " + b.Name + " "
		style := chipStyle
		if m.hy-tabRows+m.scroll == y && m.hx-1 >= x && m.hx-1 < x+ansi.StringWidth(label) {
			style = chipHover // under the pointer
		}
		row.WriteString(style.Render(label) + " ")
		hits = append(hits, hit{x0: x, x1: x + ansi.StringWidth(label), kind: hitBlock, field: field, arg: j})
		x += ansi.StringWidth(label) + 1
	}
	return row.String(), hits
}

// nearest is the boundary closest to the cell (x, y) of the page.
func nearest(bounds []boundary, x, y int) (int, bool) {
	best, found, bestCost := 0, false, 0
	for _, b := range bounds {
		cost := abs(b.y-y)*1000 + abs(b.x-x)
		if !found || cost < bestCost {
			best, found, bestCost = b.index, true, cost
		}
	}
	return best, found
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
