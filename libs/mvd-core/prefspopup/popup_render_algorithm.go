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
	lines []string
	hits  []hit
	spans []span
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
	w := max(m.w, 20)
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
			if editingHere(i) {
				shown = m.input.View()
			}
			swatch := lipgloss.NewStyle().Background(lipgloss.Color(value)).Render("  ")
			y := add(focused, label+" "+swatch+" "+shown)
			l.hits = append(l.hits, hit{y: y, x0: gutterWidth + 29, x1: w, kind: hitValue, field: i})
		case kindChoice:
			add(focused, look.Title.Render(f.label))
			current := value
			for j, option := range f.options {
				text := "  " + look.Text.Render(f.optionText(option))
				if option == current {
					text = look.Yellow.Render("› ") + look.Text.Bold(true).Render(f.optionText(option))
				}
				y := add(focused, text)
				l.hits = append(l.hits, hit{y: y, x0: gutterWidth, x1: w, kind: hitOption, field: i, arg: j})
			}
		case kindStepper:
			add(focused, look.Title.Render(f.label))
			if f.hint != "" {
				add(focused, look.Dim.Render(f.hint))
			}
			shown := f.numberText(value)
			y := add(focused, look.Yellow.Render("-")+" "+look.Text.Bold(true).Render(shown)+" "+look.Yellow.Render("+"))
			vw := width(shown)
			l.hits = append(l.hits,
				hit{y: y, x0: gutterWidth, x1: gutterWidth + 2, kind: hitMinus, field: i},
				hit{y: y, x0: gutterWidth + 2 + vw, x1: gutterWidth + 5 + vw, kind: hitPlus, field: i})
		case kindText, kindFolder, kindTemplate:
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
				add(focused, look.Dim.Render(f.hint))
			}
			if f.kind == kindTemplate {
				addBlocks(&l, add, i, focused, m.cfg.MergeOutputFormat, value)
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

// blocks draws the blocks that add a field to the template, and the example it makes.
func addBlocks(l *pageLayout, add func(bool, string) int, field int, focused bool, ext, template string) {
	badge := lipgloss.NewStyle().Bold(true).
		Background(lipgloss.Color(look.Colours.Accent)).Foreground(lipgloss.Color("#000000"))
	var row strings.Builder
	x := gutterWidth
	type placed struct{ x0, x1, index int }
	var at []placed
	for j, b := range templateBlocks {
		label := " " + b.name + " "
		at = append(at, placed{x, x + width(label), j})
		row.WriteString(badge.Render(label) + " ")
		x += width(label) + 1
	}
	y := add(focused, row.String())
	for _, p := range at {
		l.hits = append(l.hits, hit{y: y, x0: p.x0, x1: p.x1, kind: hitBlock, field: field, arg: p.index})
	}
	add(focused, look.Dim.Render("Click a block to add it to the template"))
	add(focused, "")
	add(focused, look.Title.Render("Output example"))
	add(focused, look.Text.Render(cells.Display(ExampleName(template, ext))))
}

// tabSpans are the columns each tab takes in the row of tabs.
func (m Model) tabSpans() []span {
	spans := make([]span, len(m.pages))
	x := 0
	for i, p := range m.pages {
		width := width(p.name) + 4
		spans[i] = span{x, x + width}
		x += width + 1
	}
	return spans
}

// tabLines draws the row of tabs: a box for each, the one shown lit.
func (m Model) tabLines() [tabRows]string {
	var rows [tabRows]string
	for i, p := range m.pages {
		inner := " " + p.name + " "
		edge := strings.Repeat("─", width(inner))
		style := look.Dim
		label := look.Dim.Render(inner)
		if i == m.tab {
			style = look.BorderOn.Bold(true)
			label = look.BorderOn.Bold(true).Render(inner)
		}
		rows[0] += style.Render("╭"+edge+"╮") + " "
		rows[1] += style.Render("│") + label + style.Render("│") + " "
		rows[2] += style.Render("╰"+edge+"╯") + " "
	}
	return rows
}

// bodyRows is how many lines of the page show at once.
func (m Model) bodyRows() int { return max(m.h-tabRows-1, 1) }

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

// View draws the popup: the tabs, the page of the tab, and a line of keys or the last
// thing that happened.
func (m Model) View() string {
	w := max(m.w, 20)
	out := make([]string, 0, m.h)
	for _, row := range m.tabLines() {
		out = append(out, cells.Fit(row, w))
	}
	l := m.layout()
	rows := m.bodyRows()
	for i := 0; i < rows; i++ {
		if line := m.scroll + i; line < len(l.lines) {
			out = append(out, l.lines[line])
		} else {
			out = append(out, strings.Repeat(" ", w))
		}
	}
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
