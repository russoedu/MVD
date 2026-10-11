// Package drawing compares what a screen drew with a reference drawing, cell by cell: the
// character, the colours, bold, dim and reverse.
package drawing

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// cell is one character cell of a rendered screen: what it shows and how.
type cell struct {
	Text   string
	FG, BG string
	Attrs  string // b bold, d dim, i italic, u underline, r reverse, s strike
}

func (c cell) String() string {
	return fmt.Sprintf("%q fg=%s bg=%s %s", c.Text, orDefault(c.FG), orDefault(c.BG), c.Attrs)
}

func orDefault(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// gridOf reads a string with ANSI styling into rows of cells, w cells wide. Hyperlinks
// and other non-style sequences are skipped. A wide character takes its cell and leaves
// the next one empty, as a terminal does.
func gridOf(s string, w int) [][]cell {
	var rows [][]cell
	cur := make([]cell, 0, w)
	var fg, bg string
	attrs := map[byte]bool{}
	attrString := func() string {
		out := ""
		for _, a := range []byte("bdiursc") {
			if attrs[a] {
				out += string(a)
			}
		}
		return out
	}
	flush := func() {
		for len(cur) < w {
			cur = append(cur, cell{Text: " ", FG: fg, BG: bg, Attrs: attrString()})
		}
		rows = append(rows, cur)
		cur = make([]cell, 0, w)
	}

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch {
		case r == '\n':
			flush()
		case r == '\r':
		case r == 0x1b && i+1 < len(runes) && runes[i+1] == '[':
			j := i + 2
			for j < len(runes) && (runes[j] < 0x40 || runes[j] > 0x7e) {
				j++
			}
			if j < len(runes) && runes[j] == 'm' {
				applySGR(string(runes[i+2:j]), &fg, &bg, attrs)
			}
			i = j
		case r == 0x1b && i+1 < len(runes) && runes[i+1] == ']':
			j := i + 2
			for j < len(runes) && runes[j] != 0x07 && (runes[j] != 0x1b || j+1 >= len(runes) || runes[j+1] != '\\') {
				j++
			}
			if j < len(runes) && runes[j] == 0x1b {
				j++
			}
			i = j
		default:
			width := ansi.StringWidth(string(r))
			if width == 0 && len(cur) > 0 {
				cur[len(cur)-1].Text += string(r)
				continue
			}
			cur = append(cur, cell{Text: string(r), FG: fg, BG: bg, Attrs: attrString()})
			for k := 1; k < width; k++ {
				cur = append(cur, cell{Text: "", FG: fg, BG: bg, Attrs: attrString()})
			}
		}
	}
	if len(cur) > 0 || len(rows) == 0 {
		flush()
	}
	return rows
}

// applySGR applies the parameters of one "ESC [ ... m" sequence.
func applySGR(params string, fg, bg *string, attrs map[byte]bool) {
	if params == "" {
		params = "0"
	}
	parts := strings.Split(strings.ReplaceAll(params, ":", ";"), ";")
	num := func(i int) int {
		if i >= len(parts) {
			return 0
		}
		n, _ := strconv.Atoi(parts[i])
		return n
	}
	for i := 0; i < len(parts); i++ {
		switch n := num(i); {
		case n == 0:
			*fg, *bg = "", ""
			for k := range attrs {
				delete(attrs, k)
			}
		case n == 1:
			attrs['b'] = true
		case n == 2:
			attrs['d'] = true
		case n == 3:
			attrs['i'] = true
		case n == 4:
			attrs['u'] = true
		case n == 7:
			attrs['r'] = true
		case n == 9:
			attrs['s'] = true
		case n == 22:
			delete(attrs, 'b')
			delete(attrs, 'd')
		case n == 23:
			delete(attrs, 'i')
		case n == 24:
			delete(attrs, 'u')
		case n == 27:
			delete(attrs, 'r')
		case n == 29:
			delete(attrs, 's')
		case n >= 30 && n <= 37:
			*fg = "c" + strconv.Itoa(n-30)
		case n == 39:
			*fg = ""
		case n >= 40 && n <= 47:
			*bg = "c" + strconv.Itoa(n-40)
		case n == 49:
			*bg = ""
		case n >= 90 && n <= 97:
			*fg = "c" + strconv.Itoa(n-90+8)
		case n >= 100 && n <= 107:
			*bg = "c" + strconv.Itoa(n-100+8)
		case n == 38 || n == 48:
			target := fg
			if n == 48 {
				target = bg
			}
			switch num(i + 1) {
			case 5:
				*target = "c" + strconv.Itoa(num(i+2))
				i += 2
			case 2:
				*target = fmt.Sprintf("#%02x%02x%02x", num(i+2), num(i+3), num(i+4))
				i += 4
			}
		}
	}
}

// visible drops what cannot be seen of a cell: the foreground colour and the weight of
// a blank cell. Its background, reverse video, underline and strike stay, since they show.
func (c cell) visible() cell {
	if c.Text == " " || c.Text == "" {
		c.FG = ""
		c.Attrs = strings.Map(func(r rune) rune {
			if r == 'b' || r == 'd' || r == 'i' {
				return -1
			}
			return r
		}, c.Attrs)
	}
	return c
}

// sameCells compares two renders cell by cell, as they look, and returns what differs,
// at most limit differences, as readable lines ("row 3 col 12: want ... got ...").
// SameCells lists, at most limit of them, the cells where got differs from want (both w cells
// wide).
func SameCells(want, got string, w int, limit int) []string {
	a, b := gridOf(want, w), gridOf(got, w)
	var diffs []string
	add := func(format string, args ...interface{}) {
		if len(diffs) < limit {
			diffs = append(diffs, fmt.Sprintf(format, args...))
		}
	}
	if len(a) != len(b) {
		add("height: want %d rows, got %d", len(a), len(b))
	}
	for y := 0; y < len(a) && y < len(b); y++ {
		if len(a[y]) != len(b[y]) {
			add("row %d: want %d cells, got %d", y, len(a[y]), len(b[y]))
			continue
		}
		for x := range a[y] {
			if a[y][x].visible() != b[y][x].visible() {
				add("row %d col %d: want %s, got %s", y, x, a[y][x], b[y][x])
			}
		}
	}
	return diffs
}
