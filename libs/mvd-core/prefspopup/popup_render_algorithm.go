package prefspopup

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/look"
)

// tabRows is how many rows the row of tabs takes.
const tabRows = 3

// gutterWidth is the columns before what a field draws: the bar of the focused one.
const gutterWidth = 2

// hitKind is what a click on a part of the page does.
type hitKind int

const (
	hitOption hitKind = iota // sets the option arg of the field
	hitMinus                 // steps the number of the field down
	hitPlus                  // steps the number of the field up
	hitBlock                 // adds the template block arg to the template
	hitValue                 // starts changing the value of the field
	hitCaret                 // puts the cursor of the template at the boundary arg
	hitChipHandle            // takes hold of block arg of the template, to carry it
	hitChipRemove            // takes block arg out of the template
)

// hit is a part of the page that takes clicks: line y of the page, columns x0 up to x1.
type hit struct {
	y, x0, x1 int
	kind      hitKind
	field     int
	arg       int
}

// span is the lines of the page a field takes.
type span struct{ first, last int }

// pageLayout is the page of a tab drawn once: its lines, where it takes clicks, and the
// lines of each field.
type pageLayout struct {
	lines  []string
	hits   []hit
	spans  []span
	bounds []boundary // the places between the units of the template
}

func width(s string) int { return ansi.StringWidth(s) }


func capitalise(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// optionText is how an option of a field reads.
func (f field) optionText(option string) string {
	if f.show != nil {
		return f.show(option)
	}
	return capitalise(option)
}

// numberText is how a number of a field reads.
func (f field) numberText(value string) string {
	if value == "0" && f.off != "" {
		return f.off
	}
	return value
}

// layout draws the page of the tab w columns wide.
func (m Model) layout() pageLayout {
	w := m.bodyWidth()
	var l pageLayout
	fields := m.pages[m.tab].fields
	selected := lipgloss.NewStyle().Bold(true).
		Background(lipgloss.Color(look.Colours.Selected)).Foreground(lipgloss.Color(look.Colours.Text))
	bar := look.Title.Render("┃") + " "
	add := func(focused bool, text string) int {
		gutter := "  "
		if focused {
			gutter = bar
		}
		l.lines = append(l.lines, cells.Fit(gutter+text, w))
		return len(l.lines) - 1
	}
	editingHere := func(i int) bool { return m.editing && i == m.cursor }
	// A description is as long as it needs to be: it goes on to the next line at the width of the page.
	addHint := func(focused bool, text string) {
		for _, line := range strings.Split(ansi.Wrap(text, max(w-gutterWidth-1, 10), ""), "\n") {
			add(focused, look.Dim.Render(line))
		}
	}

	for i, f := range fields {
		focused := i == m.cursor
		first := len(l.lines)
		value := ""
		if f.get != nil {
			value = f.get(m.cfg)
		}
		switch f.kind {
		case kindColour:
			label := cells.Fit(f.label, 28)
			if focused {
				label = selected.Render(label)
			} else {
				label = look.Text.Render(label)
			}
			shown := look.Cyan.Render(value)
			if value == "" {
				shown = look.Dim.Render("terminal default")
			}
			if editingHere(i) {
				shown = m.input.View()
			}
			swatch := "  "
			if value != "" {
				swatch = lipgloss.NewStyle().Background(lipgloss.Color(value)).Render("  ")
			}
			y := add(focused, label+" "+swatch+" "+shown)
			l.hits = append(l.hits, hit{y: y, x0: gutterWidth + 29, x1: w, kind: hitValue, field: i})
		case kindChoice:
			add(focused, look.Title.Render(f.label))
			if f.hint != "" {
				addHint(focused, f.hint)
			}
			current := value
			for j, option := range f.options {
				text := "  " + look.Text.Render(f.optionText(option))
				if option == current {
					text = look.Yellow.Render("› ") + look.Text.Bold(true).Render(f.optionText(option))
				}
				if f.swatches != nil {
					text += "  " + swatchRow(f.swatches(option))
				}
				y := add(focused, text)
				l.hits = append(l.hits, hit{y: y, x0: gutterWidth, x1: w, kind: hitOption, field: i, arg: j})
			}
		case kindStepper:
			add(focused, look.Title.Render(f.label))
			if f.hint != "" {
				addHint(focused, f.hint)
			}
			shown := f.numberText(value)
			y := add(focused, look.Yellow.Render("-")+" "+look.Text.Bold(true).Render(shown)+" "+look.Yellow.Render("+"))
			vw := width(shown)
			l.hits = append(l.hits,
				hit{y: y, x0: gutterWidth, x1: gutterWidth + 2, kind: hitMinus, field: i},
				hit{y: y, x0: gutterWidth + 2 + vw, x1: gutterWidth + 5 + vw, kind: hitPlus, field: i})
		case kindTemplate:
			add(focused, look.Title.Render(f.label))
			if f.hint != "" {
				addHint(focused, f.hint)
			}
			rows, hits, bounds := m.chips(i, w)
			first := len(l.lines)
			for _, row := range rows {
				add(focused, row)
			}
			for _, h := range hits {
				h.y += first
				l.hits = append(l.hits, h)
			}
			for _, b := range bounds {
				b.y += first
				l.bounds = append(l.bounds, b)
			}
			add(focused, "")
			palette, hits := m.palette(i, len(l.lines))
			y := add(focused, palette)
			for _, h := range hits {
				h.y = y
				l.hits = append(l.hits, h)
			}
			add(focused, look.Dim.Render("Click a block to add it at the cursor, drag ≡ to reorder, × to remove"))
			add(focused, "")
			add(focused, look.Title.Render("Output example"))
			add(focused, look.Text.Render(cells.Display(ExampleName(value, m.cfg.MergeOutputFormat))))
		case kindText, kindFolder:
			add(focused, look.Title.Render(f.label))
			shown := look.Text.Render(cells.Display(value))
			switch {
			case editingHere(i):
				shown = m.input.View()
			case value == "":
				shown = look.Dim.Render(f.placeholder)
			}
			y := add(focused, shown)
			l.hits = append(l.hits, hit{y: y, x0: gutterWidth, x1: w, kind: hitValue, field: i})
			if f.hint != "" {
				addHint(focused, f.hint)
			}
		case kindAction:
			text := look.Yellow.Render("[ " + f.label + " ]")
			if focused {
				text = look.Title.Reverse(true).Render("[ " + f.label + " ]")
			}
			y := add(focused, text)
			l.hits = append(l.hits, hit{y: y, x0: gutterWidth, x1: gutterWidth + width(f.label) + 4, kind: hitValue, field: i})
		}
		l.spans = append(l.spans, span{first, len(l.lines) - 1})
	}
	if fields[0].kind == kindColour {
		l.lines = append(l.lines, "",
			cells.Fit("  "+look.Dim.Render("Colours are #rgb or #rrggbb and apply as soon as you accept them."), w),
			cells.Fit("  "+look.Dim.Render("Press d for the default of the colour."), w))
	}
	return l
}

// tabSpans are the columns each tab takes in the row of tabs: the tabs touch, as Cuppa's
// tab pages draw them.
func (m Model) tabSpans() []span {
	spans := make([]span, len(m.pages))
	x := 0
	for i, p := range m.pages {
		w := width(p.name) + 4
		spans[i] = span{x, x + w}
		x += w
	}
	return spans
}

// tabLines draws the row of tabs the way Cuppa's tab pages do, with the same characters so
// every corner lines up: each tab is a box with rounded top corners, side by side, and the
// one shown is open at the bottom into the box of the page (┘ and └), the others closed
// onto its top edge (┴). The third row is that top edge.
func (m Model) tabLines() [tabRows]string {
	var rows [tabRows]string
	w := max(m.w, 20)
	edge := look.BorderOn
	var line strings.Builder
	used := 0
	for i, p := range m.pages {
		label := " " + p.name + " "
		bar := strings.Repeat("─", width(label))
		style, text := edge, look.Dim.Render(label)
		if i == m.tab {
			text = edge.Bold(true).Render(label)
		}
		if s := m.tabSpans()[i]; m.hy >= 0 && m.hy < tabRows && m.hx >= s.first && m.hx < s.last {
			text = look.Text.Underline(true).Render(label) // under the pointer
		}
		rows[0] += style.Render("╭" + bar + "╮")
		rows[1] += style.Render("│") + text + style.Render("│")
		switch {
		case i == m.tab && i == 0:
			line.WriteString(edge.Render("│" + strings.Repeat(" ", width(label)) + "└"))
		case i == m.tab:
			line.WriteString(edge.Render("┘" + strings.Repeat(" ", width(label)) + "└"))
		case i == 0:
			line.WriteString(edge.Render("├" + bar + "┴"))
		default:
			line.WriteString(edge.Render("┴" + bar + "┴"))
		}
		used += width(label) + 2
	}
	rows[2] = line.String() + edge.Render(strings.Repeat("─", max(w-used-1, 0))+"╮")
	return rows
}

// bodyRows is how many lines of the page show at once.
func (m Model) bodyRows() int { return max(m.h-tabRows-2, 1) }

// bodyWidth is the columns of the page: the inside of the box it is in.
func (m Model) bodyWidth() int { return max(m.w, 20) - 2 }

// reveal scrolls the page so the field the cursor is on shows.
func (m Model) reveal() Model {
	l := m.layout()
	if m.cursor >= len(l.spans) {
		return m
	}
	s, rows := l.spans[m.cursor], m.bodyRows()
	if s.last >= m.scroll+rows {
		m.scroll = s.last - rows + 1
	}
	if s.first < m.scroll {
		m.scroll = s.first
	}
	m.scroll = max(min(m.scroll, len(l.lines)-rows), 0)
	return m
}

// View draws the popup: the tabs, the page of the tab in a box under them, and a line of
// keys or the last thing that happened.
func (m Model) View() string {
	w := max(m.w, 20)
	out := make([]string, 0, m.h)
	for _, row := range m.tabLines() {
		out = append(out, cells.Fit(row, w))
	}
	l := m.layout()
	edge := look.BorderOn.Render("│")
	rows := m.bodyRows()
	for i := 0; i < rows; i++ {
		line := strings.Repeat(" ", w-2)
		if at := m.scroll + i; at < len(l.lines) {
			line = l.lines[at]
		}
		out = append(out, edge+line+edge)
	}
	out = append(out, look.BorderOn.Render("╰"+strings.Repeat("─", w-2)+"╯"))
	return strings.Join(append(out, cells.Fit(" "+m.hintLine(), w)), "\n")
}

// hintLine is what the bottom line says.
func (m Model) hintLine() string {
	switch {
	case m.status != "" && m.statusBad:
		return look.Red.Render(m.status)
	case m.status != "":
		return look.Dim.Render(m.status)
	case m.editing:
		return look.Dim.Render("enter apply  esc cancel")
	}
	return look.Dim.Render("↑↓ move  ←→ change  enter edit  tab page  s save  esc close")
}

// swatchRow draws a row of colours, two cells each; a colour that is "" is left blank.
func swatchRow(colours []string) string {
	var b strings.Builder
	for _, c := range colours {
		if c == "" {
			b.WriteString(look.Dim.Render("··"))
			continue
		}
		b.WriteString(lipgloss.NewStyle().Background(lipgloss.Color(c)).Render("  "))
	}
	return b.String()
}
