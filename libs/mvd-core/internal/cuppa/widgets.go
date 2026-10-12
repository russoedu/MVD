// Written by the Go export of Cuppa and maintained by hand since: change it here.

package cuppa

// widgets.go holds the components of the design that runtime.go does not draw
// itself: Lip Gloss tables, trees and tabs, the Bubbles help and file picker,
// the layout sketches (join and place) and the look of the community widgets
// (flex box, boxer, date picker, overlay, status bar, file tree), which are
// drawn here with Lip Gloss so the project needs no extra module for them.
//
// The sketches are drawn on a small grid of cells (drawing) the way the designer
// draws them. Cuppa writes them again on every export.

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	ltree "charm.land/lipgloss/v2/tree"
)

func init() {
	Extra["lipgloss.table"] = lgTablePart
	Extra["lipgloss.tree"] = lgTreePart
	Extra["lipgloss.tabs"] = tabsPart
	Extra["lipgloss.joinh"] = joinHPart
	Extra["lipgloss.joinv"] = joinVPart
	Extra["lipgloss.place"] = placePart
	Extra["bubbles.help"] = helpPart
	Extra["bubbles.filepicker"] = filePickerPart
	Extra["community.flexbox"] = flexBoxPart
	Extra["community.boxer"] = boxerPart
	Extra["community.datepicker"] = datePickerPart
	Extra["community.overlay"] = overlayPart
	Extra["community.statusbar"] = statusBarPart
	Extra["community.filetree"] = fileTreePart
	Extra["huh.form"] = huhFormPart
	Extra["draw.run"] = drawRunPart
}

// drawRunPart is a run of the drawing: some text in one style.
func drawRunPart(c Placed, p Props) *Part {
	st := lipgloss.NewStyle()
	if fg := p.Str("fg"); fg != "" {
		st = st.Foreground(lipgloss.Color(fg))
	}
	if bg := p.Str("bg"); bg != "" {
		st = st.Background(lipgloss.Color(bg))
	}
	return StaticPart(st.Render(p.Str("text")))
}

// ---- a small grid of styled cells -----------------------------------------

type sty struct {
	fg, bg string
	bold   bool
}

type cell struct {
	r rune
	s sty
}

type drawing struct {
	w, h  int
	Cells []cell
}

var dimStyle = sty{fg: "240"}

// mutedStyle is the quiet text style of the design: its muted colour, or grey.
func mutedStyle(p Props) sty {
	if c := p.Str("theme.muted"); c != "" {
		return sty{fg: c}
	}
	return dimStyle
}

func newDrawing(w, h int) *drawing {
	return &drawing{w: max(w, 0), h: max(h, 0), Cells: make([]cell, max(w, 0)*max(h, 0))}
}

func (c *drawing) set(x, y int, r rune, s sty) {
	if x >= 0 && y >= 0 && x < c.w && y < c.h {
		c.Cells[y*c.w+x] = cell{r, s}
	}
}

// text writes s from (x, y), at most limit cells (no limit when limit <= 0),
// and returns how many cells it wrote.
func (c *drawing) text(x, y int, s string, st sty, limit int) int {
	n := 0
	for _, r := range s {
		if limit > 0 && n >= limit {
			break
		}
		c.set(x+n, y, r, st)
		n++
	}
	return n
}

func (c *drawing) fill(x, y, w, h int, r rune, s sty) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			c.set(i, j, r, s)
		}
	}
}

// box clears the rectangle and draws a border of the named style around it.
func (c *drawing) box(x, y, w, h int, border string, s sty) {
	c.fill(x, y, w, h, ' ', s)
	if w < 2 || h < 2 {
		return
	}
	b := borderOf(border)
	first := func(v string) rune { return []rune(v)[0] }
	for i := x + 1; i < x+w-1; i++ {
		c.set(i, y, first(b.Top), s)
		c.set(i, y+h-1, first(b.Bottom), s)
	}
	for j := y + 1; j < y+h-1; j++ {
		c.set(x, j, first(b.Left), s)
		c.set(x+w-1, j, first(b.Right), s)
	}
	c.set(x, y, first(b.TopLeft), s)
	c.set(x+w-1, y, first(b.TopRight), s)
	c.set(x, y+h-1, first(b.BottomLeft), s)
	c.set(x+w-1, y+h-1, first(b.BottomRight), s)
}

// String draws the grid, one Lip Gloss style per run of equal cells.
func (c *drawing) String() string {
	rows := make([]string, c.h)
	for y := 0; y < c.h; y++ {
		var b strings.Builder
		var run []rune
		var cur sty
		flush := func() {
			if len(run) == 0 {
				return
			}
			st := lipgloss.NewStyle().Bold(cur.bold)
			if cur.fg != "" {
				st = st.Foreground(lipgloss.Color(cur.fg))
			}
			if cur.bg != "" {
				st = st.Background(lipgloss.Color(cur.bg))
			}
			b.WriteString(st.Render(string(run)))
			run = run[:0]
		}
		for x := 0; x < c.w; x++ {
			ce := c.Cells[y*c.w+x]
			r := ce.r
			if r == 0 {
				r = ' '
			}
			if len(run) > 0 && ce.s != cur {
				flush()
			}
			cur = ce.s
			run = append(run, r)
		}
		flush()
		rows[y] = b.String()
	}
	return strings.Join(rows, "\n")
}

// sketch makes a static component from a drawing function.
func sketch(c Placed, p Props, draw func(cv *drawing, p Props)) *Part {
	cv := newDrawing(c.W, c.H)
	draw(cv, p)
	return StaticPart(cv.String())
}

// wrapText breaks text into lines of at most width runes, at spaces where it can.
func wrapText(text string, width int) []string {
	if width < 1 {
		return nil
	}
	var out []string
	line := ""
	for _, word := range strings.Fields(text) {
		for len([]rune(word)) > width {
			r := []rune(word)
			if line != "" {
				out = append(out, line)
				line = ""
			}
			out = append(out, string(r[:width]))
			word = string(r[width:])
		}
		switch {
		case line == "":
			line = word
		case len([]rune(line))+1+len([]rune(word)) <= width:
			line += " " + word
		default:
			out = append(out, line)
			line = word
		}
	}
	if line != "" {
		out = append(out, line)
	}
	return out
}

func drawFrameWithName(cv *drawing, name string) {
	cv.box(0, 0, cv.w, cv.h, "normal", dimStyle)
	if cv.h >= 3 {
		cv.text(max((cv.w-len([]rune(name)))/2, 1), cv.h/2, name, sty{}, cv.w-2)
	} else {
		cv.text(1, 0, name, sty{}, cv.w-2)
	}
}

// ---- Lip Gloss -----------------------------------------------------------

func lgTablePart(c Placed, p Props) *Part {
	var rows [][]string
	for _, r := range splitRaw(p.Str("rows"), ";") {
		rows = append(rows, Split(r, ","))
	}
	t := table.New().
		Headers(Split(p.Str("headers"), ",")...).
		Rows(rows...).
		Border(borderOf(p.Str("border"))).
		BorderStyle(fg(p.Str("color"))).
		StyleFunc(func(row, col int) lipgloss.Style { return lipgloss.NewStyle().Padding(0, 1) })
	return StaticPart(t.String())
}

func lgTreePart(c Placed, p Props) *Part {
	var items []any
	for _, s := range Split(p.Str("items"), ",") {
		items = append(items, s)
	}
	t := ltree.Root(p.Str("root")).Child(items...).
		RootStyle(fg(p.Str("color")).Bold(true)).
		EnumeratorStyle(fg(p.Str("color"))).
		ItemStyle(fg(p.Str("color")).PaddingLeft(1))
	return StaticPart(t.String())
}

// tabsPart is the Lip Gloss "layout" tabs: a row of tabs whose active tab
// opens into the bordered window below it.
func tabsPart(c Placed, p Props) *Part {
	return sketch(c, p, func(cv *drawing, p Props) {
		if cv.h < 4 || cv.w < 8 {
			drawFrameWithName(cv, "Tabs")
			return
		}
		names := Split(p.Str("tabs"), ",")
		if len(names) == 0 {
			names = []string{"Tab"}
		}
		active := min(max(p.Integer("active", 1), 1), len(names)) - 1
		edge := sty{fg: p.Str("color")}
		strong := sty{fg: p.Str("color"), bold: true}
		put := func(x, y int, r rune) { cv.set(x, y, r, edge) }
		for x := 1; x < cv.w-1; x++ {
			put(x, 2, '─')
			put(x, cv.h-1, '─')
		}
		for y := 3; y < cv.h-1; y++ {
			put(0, y, '│')
			put(cv.w-1, y, '│')
		}
		put(0, cv.h-1, '╰')
		put(cv.w-1, cv.h-1, '╯')
		put(cv.w-1, 2, '╮')
		put(0, 2, '├')
		x := 0
		for i, name := range names {
			text := []rune(" " + name + " ")
			w := len(text)
			if x+w+2 > cv.w-1 {
				break
			}
			for dx := 1; dx <= w; dx++ {
				put(x+dx, 0, '─')
			}
			put(x, 0, '╭')
			put(x+w+1, 0, '╮')
			put(x, 1, '│')
			put(x+w+1, 1, '│')
			st := mutedStyle(p)
			if i == active {
				st = strong
			}
			cv.text(x+1, 1, string(text), st, w)
			if i == active {
				put(x, 2, '┘')
				if x == 0 {
					put(x, 2, '│')
				}
				for dx := 1; dx <= w; dx++ {
					cv.set(x+dx, 2, ' ', sty{})
				}
				put(x+w+1, 2, '└')
			} else {
				for dx := 1; dx <= w; dx++ {
					put(x+dx, 2, '─')
				}
				put(x+w+1, 2, '┴')
				if x == 0 {
					put(x, 2, '├')
				}
			}
			x += w + 2
		}
	})
}

func joinHPart(c Placed, p Props) *Part {
	return sketch(c, p, func(cv *drawing, p Props) {
		n := max(p.Integer("columns", 2), 1)
		cv.box(0, 0, cv.w, cv.h, "normal", mutedStyle(p))
		for i := 1; i < n; i++ {
			x := cv.w * i / n
			for y := 1; y < cv.h-1; y++ {
				cv.set(x, y, '┆', mutedStyle(p))
			}
		}
		cv.text(2, 0, " JoinHorizontal ", mutedStyle(p), cv.w-3)
	})
}

func joinVPart(c Placed, p Props) *Part {
	return sketch(c, p, func(cv *drawing, p Props) {
		n := max(p.Integer("rows", 2), 1)
		cv.box(0, 0, cv.w, cv.h, "normal", mutedStyle(p))
		for i := 1; i < n; i++ {
			y := cv.h * i / n
			for x := 1; x < cv.w-1; x++ {
				cv.set(x, y, '┄', mutedStyle(p))
			}
		}
		cv.text(2, 0, " JoinVertical ", mutedStyle(p), cv.w-3)
	})
}

func placePart(c Placed, p Props) *Part {
	return sketch(c, p, func(cv *drawing, p Props) {
		cv.box(0, 0, cv.w, cv.h, "normal", mutedStyle(p))
		text := p.Str("text")
		w := len([]rune(text))
		x := 1
		switch p.Str("horizontal") {
		case "center":
			x = (cv.w - w) / 2
		case "right":
			x = cv.w - w - 1
		}
		y := 1
		switch p.Str("vertical") {
		case "center":
			y = cv.h / 2
		case "bottom":
			y = cv.h - 2
		}
		cv.text(x, y, text, sty{}, cv.w-2)
		cv.text(2, 0, " Place ", mutedStyle(p), cv.w-3)
	})
}

// ---- Bubbles -------------------------------------------------------------

// helpPart is the Bubbles help view over the "key:description" list.
func helpPart(c Placed, p Props) *Part {
	var bindings []key.Binding
	for _, item := range Split(p.Str("bindings"), ",") {
		k, desc, _ := strings.Cut(item, ":")
		bindings = append(bindings, key.NewBinding(key.WithKeys(k), key.WithHelp(k, desc)))
	}
	m := help.New()
	m.SetWidth(c.W)
	if col := p.Str("color"); col != "" {
		st := help.DefaultDarkStyles()
		st.ShortKey = st.ShortKey.Foreground(lipgloss.Color(col))
		st.FullKey = st.FullKey.Foreground(lipgloss.Color(col))
		m.Styles = st
	}
	if p.Flag("expanded") {
		var columns [][]key.Binding
		for i := 0; i < len(bindings); i += 3 {
			columns = append(columns, bindings[i:min(i+3, len(bindings))])
		}
		return StaticPart(m.FullHelpView(columns))
	}
	return StaticPart(m.ShortHelpView(bindings))
}

// filePickerPart is the Bubbles file picker, browsing the folder named by the
// path property (or where the program runs).
func filePickerPart(c Placed, p Props) *Part {
	fp := filepicker.New()
	if dir := p.Str("path"); dir != "" {
		if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(dir, "~") {
			dir = home + strings.TrimPrefix(dir, "~")
		}
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			fp.CurrentDirectory = dir
		}
	}
	fp.SetHeight(max(c.H, 1))
	return &Part{
		keys:   true,
		start:  fp.Init(),
		update: func(msg tea.Msg) tea.Cmd { var cmd tea.Cmd; fp, cmd = fp.Update(msg); return cmd },
		view:   func() string { return fp.View() },
	}
}

// ---- community widgets, drawn with Lip Gloss ------------------------------

// slotOf splits length into n slots separated by one-cell gaps and returns the
// start and size of slot i.
func slotOf(length, n, i int) (start, size int) {
	size = max((length-(n-1))/n, 1)
	return i * (size + 1), size
}

func selectedStyle(color string) sty { return sty{fg: "0", bg: color, bold: true} }

func flexBoxPart(c Placed, p Props) *Part {
	return sketch(c, p, func(cv *drawing, p Props) {
		rows, cols := max(p.Integer("rows", 1), 1), max(p.Integer("columns", 1), 1)
		st := sty{fg: p.Str("color")}
		for r := 0; r < rows; r++ {
			y, h := slotOf(cv.h, rows, r)
			for col := 0; col < cols; col++ {
				x, w := slotOf(cv.w, cols, col)
				cv.box(x, y, w, h, "rounded", st)
				if h >= 3 && w >= 5 {
					cv.text(x+2, y+h/2, fmt.Sprintf("%d,%d", r+1, col+1), mutedStyle(p), w-3)
				}
			}
		}
	})
}

func boxerPart(c Placed, p Props) *Part {
	return sketch(c, p, func(cv *drawing, p Props) {
		n := max(p.Integer("children", 2), 1)
		st := sty{fg: p.Str("color")}
		vertical := p.Str("orientation") == "vertical"
		for i := 0; i < n; i++ {
			x, y, w, h := 0, 0, cv.w, cv.h
			if vertical {
				y, h = slotOf(cv.h, n, i)
			} else {
				x, w = slotOf(cv.w, n, i)
			}
			cv.box(x, y, w, h, "normal", st)
			if w >= 9 && h >= 3 {
				cv.text(x+2, y+h/2, fmt.Sprintf("model %d", i+1), mutedStyle(p), w-3)
			}
		}
	})
}

func datePickerPart(c Placed, p Props) *Part {
	return sketch(c, p, func(cv *drawing, p Props) {
		accent := sty{fg: p.Str("color"), bold: true}
		month := p.Str("month")
		cv.text(0, 0, "◂", accent, 1)
		cv.text((cv.w-len([]rune(month)))/2, 0, month, sty{bold: true}, cv.w)
		cv.text(cv.w-1, 0, "▸", accent, 1)
		cv.text(0, 1, "Su Mo Tu We Th Fr Sa", mutedStyle(p), cv.w)
		const firstWeekday = 4 // 1 October 2026 is a Thursday
		day := min(max(p.Integer("day", 1), 1), 31)
		for d := 1; d <= 31; d++ {
			slot := firstWeekday + d - 1
			x, y := (slot%7)*3, 2+slot/7
			st := sty{}
			if d == day {
				st = selectedStyle(p.Str("color"))
			}
			cv.text(x, y, fmt.Sprintf("%2d", d), st, 2)
		}
	})
}

func overlayPart(c Placed, p Props) *Part {
	return sketch(c, p, func(cv *drawing, p Props) {
		color := p.Str("color")
		cv.box(0, 0, cv.w, cv.h, "thick", sty{fg: color})
		if cv.w > 4 {
			cv.text(2, 0, " "+p.Str("title")+" ", sty{fg: color, bold: true}, cv.w-4)
		}
		for i, l := range wrapText(p.Str("body"), cv.w-4) {
			if 2+i >= cv.h-2 {
				break
			}
			cv.text(2, 1+i, l, sty{}, cv.w-4)
		}
		if cv.h >= 5 {
			n := cv.text(2, cv.h-2, " OK ", selectedStyle(color), cv.w-4)
			cv.text(3+n, cv.h-2, " Cancel ", sty{fg: "250", bg: "238"}, cv.w-5-n)
		}
	})
}

func statusBarPart(c Placed, p Props) *Part {
	return sketch(c, p, func(cv *drawing, p Props) {
		cv.fill(0, 0, cv.w, cv.h, ' ', sty{bg: "236"})
		x := cv.text(0, 0, " "+p.Str("left")+" ", selectedStyle(p.Str("color")), cv.w)
		x += cv.text(x, 0, " "+p.Str("middle")+" ", sty{fg: "252", bg: "240"}, cv.w-x)
		end := " " + p.Str("end") + " "
		right := " " + p.Str("right") + " "
		ew, rw := len([]rune(end)), len([]rune(right))
		if cv.w-ew >= x {
			cv.text(cv.w-ew, 0, end, sty{fg: "0", bg: "250", bold: true}, ew)
		}
		if cv.w-ew-rw >= x {
			cv.text(cv.w-ew-rw, 0, right, sty{fg: "252", bg: "238"}, rw)
		}
	})
}

func fileTreePart(c Placed, p Props) *Part {
	return sketch(c, p, func(cv *drawing, p Props) {
		entries := Split(p.Str("entries"), ",")
		selected := 0
		if len(entries) > 0 {
			selected = min(max(p.Integer("selected", 0), 0), len(entries)-1)
		}
		for i, e := range entries {
			if i >= cv.h {
				break
			}
			name, st := "  "+e, sty{}
			if strings.HasSuffix(e, "/") {
				name, st = "▸ "+e, sty{fg: "39"}
			}
			if i == selected {
				st = selectedStyle(p.Str("color"))
				cv.fill(0, i, cv.w, 1, ' ', st)
			}
			cv.text(0, i, name, st, cv.w)
		}
	})
}

// huhFormPart is a Huh form group: a rounded frame with its title in the top
// edge, a row of field placeholders and a footer line.
func huhFormPart(c Placed, p Props) *Part {
	return sketch(c, p, func(cv *drawing, p Props) {
		color := p.Str("color")
		cv.box(0, 0, cv.w, cv.h, "rounded", sty{fg: color})
		if t := p.Str("title"); t != "" && cv.w > 4 {
			cv.text(2, 0, " "+t+" ", sty{fg: color, bold: true}, cv.w-4)
		}
		rows := (cv.h - 3) / 2
		for i := 0; i < rows; i++ {
			y := 1 + i*2
			cv.set(2, y, '┃', sty{fg: color})
			cv.text(4, y, strings.Repeat("▁", max(cv.w-8, 0)), mutedStyle(p), cv.w-6)
		}
		if f := p.Str("footer"); f != "" && cv.h > 2 {
			cv.text(2, cv.h-2, f, mutedStyle(p), cv.w-4)
		}
	})
}
