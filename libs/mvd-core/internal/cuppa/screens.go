// Written by the Go export of Cuppa and maintained by hand since: change it here.

package cuppa

// The shared part of the screens Cuppa generated: what a screen returns
// (Frame), the colours a program can hand it (Theme) and the drawing that every
// screen's view uses.

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Frame is one drawn screen: the text to show and where its components are.
type Frame struct {
	// View is the screen, w by h cells, ready to print.
	View string
	// Regions are the visible components, back to front.
	Regions []Region
}

// Region is the rectangle of one visible component. Event is the name of the
// event a click on it raises ("" for none).
type Region struct {
	Name, Event string
	X, Y, W, H  int
}

// eventAt is the topmost region under the cell that raises an event.
func (f Frame) EventAt(x, y int) (Region, bool) {
	for i := len(f.Regions) - 1; i >= 0; i-- {
		r := f.Regions[i]
		if r.Event != "" && x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H {
			return r, true
		}
	}
	return Region{}, false
}

// RegionNamed is the rectangle of the visible component with that name, for a
// program that places something over it.
func (f Frame) RegionNamed(name string) (Region, bool) {
	for _, r := range f.Regions {
		if r.Name == name {
			return r, true
		}
	}
	return Region{}, false
}

// Theme is the colours a program hands to a screen. A component colour that
// the design did not set follows these; an empty value keeps the design's own.
type Theme struct {
	Text, Muted, Border, Secondary, Background string
}

// KeyHint is a key of a screen and what it does, for a key bar.
type KeyHint struct {
	Key, Label, Event string
}

// mutedKey is the property under which painters read the muted colour.
const mutedKey = "theme.muted"

func (t Theme) role(name string) string {
	switch name {
	case "text":
		return t.Text
	case "border":
		return t.Border
	case "secondary":
		return t.Secondary
	case "background":
		return t.Background
	}
	return ""
}

// screenPart is one component of a screen design.
type ScreenPart[P any] struct {
	Placed
	// Roles maps a colour property to the theme role it follows.
	Roles map[string]string
	// Event is raised by a click on the component.
	Event string
	// Apply puts the program's values into the properties and reports whether
	// the component is shown, for a window of w by h cells.
	Apply func(p P, v map[string]string, w, h int) bool
	// Late sets properties that depend on the size the layout gave the
	// component: w and h are its own.
	Late func(p P, v map[string]string, w, h int)
}

// drawScreen draws the parts for the values p in an area of w by h cells.
func DrawScreen[P any](parts []ScreenPart[P], p P, th Theme, w, h int, background string, env func(P, int, int) FitEnv) Frame {
	w, h = max(w, 1), max(h, 1)
	e := env(p, w, h)
	e.rects = map[string][4]int{}
	if th.Background != "" {
		background = th.Background
	}
	layers := []*lipgloss.Layer{lipgloss.NewLayer(backdrop(w, h, background)).Z(0)}
	var regions []Region
	for i, Part := range parts {
		c := Part.Placed
		c.Props = make(map[string]string, len(Part.Props)+1)
		for k, v := range Part.Props {
			c.Props[k] = v
		}
		for k, role := range Part.Roles {
			if v := th.role(role); v != "" {
				c.Props[k] = v
			}
		}
		if th.Muted != "" {
			c.Props[mutedKey] = th.Muted
		}
		if Part.Apply != nil && !Part.Apply(p, c.Props, w, h) {
			continue
		}
		switch {
		case c.FitEnv != nil:
			x, y, cw, ch := c.FitEnv(e)
			c.X, c.Y, c.W, c.H = x, y, max(cw, c.MinW, 1), max(ch, c.MinH, 1)
		case c.Fit != nil:
			x, y, cw, ch := c.Fit(w, h)
			c.X, c.Y, c.W, c.H = x, y, max(cw, c.MinW, 1), max(ch, c.MinH, 1)
		}
		c.Y += containerDY(c.Props)
		e.rects[c.Name] = [4]int{c.X, c.Y, c.W, c.H}
		if Part.Late != nil {
			Part.Late(p, c.Props, c.W, c.H)
		}
		layers = append(layers, lipgloss.NewLayer(build(c).view()).X(c.X).Y(c.Y).Z(i+1))
		regions = append(regions, Region{Name: c.Name, Event: Part.Event, X: c.X, Y: c.Y, W: c.W, H: c.H})
	}
	return Frame{View: lipgloss.NewCompositor(layers...).Render(), Regions: regions}
}

// layoutScreen is drawScreen without the drawing: the regions of the visible
// components for the values p in an area of w by h cells.
func LayoutScreen[P any](parts []ScreenPart[P], p P, th Theme, w, h int, env func(P, int, int) FitEnv) Frame {
	w, h = max(w, 1), max(h, 1)
	e := env(p, w, h)
	e.rects = map[string][4]int{}
	var regions []Region
	for _, Part := range parts {
		c := Part.Placed
		c.Props = make(map[string]string, len(Part.Props)+1)
		for k, v := range Part.Props {
			c.Props[k] = v
		}
		if Part.Apply != nil && !Part.Apply(p, c.Props, w, h) {
			continue
		}
		switch {
		case c.FitEnv != nil:
			x, y, cw, ch := c.FitEnv(e)
			c.X, c.Y, c.W, c.H = x, y, max(cw, c.MinW, 1), max(ch, c.MinH, 1)
		case c.Fit != nil:
			x, y, cw, ch := c.Fit(w, h)
			c.X, c.Y, c.W, c.H = x, y, max(cw, c.MinW, 1), max(ch, c.MinH, 1)
		}
		c.Y += containerDY(c.Props)
		e.rects[c.Name] = [4]int{c.X, c.Y, c.W, c.H}
		regions = append(regions, Region{Name: c.Name, Event: Part.Event, X: c.X, Y: c.Y, W: c.W, H: c.H})
	}
	return Frame{Regions: regions}
}

// Window cuts the rows a list of h rows shows out of all of them, keeping the
// selected row centred while there is room on both sides and pinned to the top
// or the bottom near the ends. first is the index of visible[0] in rows, so a
// click on visible row k is rows[first+k]. A selected index outside the list is
// clamped; h <= 0 or a list that fits returns every row with first 0.
func Window[T any](rows []T, selected, h int) (visible []T, first int) {
	n := len(rows)
	if h <= 0 || n <= h {
		return rows, 0
	}
	selected = min(max(selected, 0), n-1)
	first = selected - h/2
	if first < 0 {
		first = 0
	}
	if first+h > n {
		first = n - h
	}
	return rows[first : first+h], first
}

// hintIndexAt is the index of the key bar hint under column x of the bar (the
// x of the click's event), counting the one space before the first hint and
// gap spaces between hints; ok is false between hints and after the last.
func HintIndexAt(keys [][2]string, gap, x int) (int, bool) {
	at := 1
	for i, k := range keys {
		if i > 0 {
			at += gap
		}
		w := lipgloss.Width(k[0])
		if k[1] != "" {
			w += 1 + lipgloss.Width(k[1])
		}
		if x >= at && x < at+w {
			return i, true
		}
		at += w
	}
	return 0, false
}

// CellStyle restyles one cell of a row: a colour is #rrggbb or a palette
// number; the zero value keeps the column's style.
type CellStyle struct {
	Fg, Bg             string
	Bold, Dim, Reverse bool
	// Plain keeps the column's colours off this cell: the terminal's own.
	Plain bool
}

// cellStyleCode is the style as the Rows component reads it.
func CellStyleCode(s CellStyle) string {
	if s == (CellStyle{}) {
		return "-"
	}
	flags := ""
	if s.Bold {
		flags += "b"
	}
	if s.Dim {
		flags += "d"
	}
	if s.Reverse {
		flags += "r"
	}
	if s.Plain {
		flags += "p"
	}
	return s.Fg + "|" + s.Bg + "|" + flags
}

// paletteCode is the named colours the screen is drawn with, for the
// components that read a colour by its name: name, value and the design's own
// colour for each, in turn.
func PaletteCode(parts ...string) string {
	var entries []string
	for i := 0; i+2 < len(parts); i += 3 {
		value := parts[i+1]
		if value == "" {
			value = parts[i+2]
		}
		entries = append(entries, parts[i]+"="+value)
	}
	return strings.Join(entries, ";")
}

// backdrop is the area itself: its size and background colour.
func backdrop(w, h int, background string) string {
	line := strings.Repeat(" ", w)
	if background != "" {
		line = lipgloss.NewStyle().Background(lipgloss.Color(background)).Render(line)
	}
	lines := make([]string, h)
	for i := range lines {
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

// RowStyle is how a row of a Rows component is drawn.
type RowStyle string

// The row styles. The zero value draws a normal row.
const (
	RowNormal   RowStyle = "normal"
	RowSelected RowStyle = "selected"
	RowDim      RowStyle = "dim"
	RowAccent   RowStyle = "accent"
)

// rowStyleName is the style as the component reads it.
func RowStyleName(s RowStyle) string {
	if s == "" {
		return string(RowNormal)
	}
	return string(s)
}

// escaper makes an item or a cell safe to put in a list: a backslash before a
// comma, a semicolon or a backslash is that character itself.
var escaper = strings.NewReplacer(`\`, `\\`, ",", `\,`, ";", `\;`)

// joinList is a list of items as the components read it.
func JoinList(items []string) string {
	escaped := make([]string, len(items))
	for i, s := range items {
		escaped[i] = escaper.Replace(s)
	}
	return strings.Join(escaped, ",")
}

// joinRows is the rows of a table as the components read them.
func JoinRows(rows [][]string) string {
	lines := make([]string, len(rows))
	for i, r := range rows {
		lines[i] = JoinList(r)
	}
	return strings.Join(lines, ";")
}
