// Written by the Go export of Cuppa and maintained by hand since: change it here.

package colours

import (
	tea "charm.land/bubbletea/v2"
	master "youtube-downloader/libs/mvd-core/internal/components/master"
	_ "youtube-downloader/libs/mvd-core/internal/components/widgets"
	cuppa "youtube-downloader/libs/mvd-core/internal/cuppa"
)

const (
	width      = 120
	height     = 36
	background = ""
)

// coloursEnv is what the layout expressions and conditions of the screen read.
func env(p Props, w, h int) cuppa.FitEnv {
	_ = p
	return cuppa.FitEnv{W: w, H: h, Inputs: func(name string) float64 {
		switch name {
		case "Show keys":
			if p.ShowKeys {
				return 1
			}
			return 0
		case "Show status":
			if p.ShowStatus {
				return 1
			}
			return 0
		}
		return 0
	}}
}

var parts = append(master.PartsFor(masterProps), []cuppa.ScreenPart[Props]{
	{
		Placed: cuppa.Placed{Kind: "lipgloss.rows", Name: "Slots", X: 0, Y: 1, W: 100, H: 8, Props: map[string]string{"blurColor": "", "cellstyles": "", "color": "#3b0f2a", "columns": "Mark:2:fg=@Accent,Name:25:selbg=@Selected:selfg=@Text:selbold,Gap:2,Swatch:2:glyph,Space:1,Value:fg=@Focus", "focused": "true", "gap": "0", "rows": "▸ ,Accent (title),,#ff007f,,#ff007f;  ,Focus (active border),,#00f0ff,,#00f0ff;  ,Highlight (keys\\, bars),,#ffe600,,#ffe600;  ,Success,,#3ddc84,,#3ddc84;  ,Error,,#ff4d4d,,#ff4d4d;  ,Dim (borders\\, hints),,#6b7280,,#6b7280;  ,Text,,#e5e7eb,,#e5e7eb;  ,Selected row (background),,#3b0f2a,,#3b0f2a", "styles": "selected,normal,normal,normal,normal,normal,normal,normal", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a"}, MinW: 8, MinH: 1, Fit: func(w, h int) (x, y, cw, ch int) {
			return 0, 1, cuppa.Cells((float64(w) * 100.0 / 100.0)), 8
		}},
		Event: "Pick",
		Late: func(p Props, v map[string]string, w, h int) {
			if p.SlotsRows != nil {
				rs, _ := p.SlotsRows(w, h)
				{
					rows := make([][]string, len(rs))
					cellStyles := make([][]string, len(rs))
					styles := make([]string, len(rs))
					for i, r := range rs {
						rows[i] = []string{r.Mark, r.Name, r.Gap, r.Swatch, r.Space, r.Value}
						cellStyles[i] = []string{cuppa.CellStyleCode(r.MarkStyle), cuppa.CellStyleCode(r.NameStyle), cuppa.CellStyleCode(r.GapStyle), cuppa.CellStyleCode(r.SwatchStyle), cuppa.CellStyleCode(r.SpaceStyle), cuppa.CellStyleCode(r.ValueStyle)}
						styles[i] = cuppa.RowStyleName(r.Style)
					}
					v["rows"] = cuppa.JoinRows(rows)
					v["cellstyles"] = cuppa.JoinRows(cellStyles)
					v["styles"] = cuppa.JoinList(styles)
				}
			}
		},
		Apply: func(p Props, v map[string]string, w, h int) bool {
			v["theme.palette"] = cuppa.PaletteCode("Accent", p.Palette.Accent, "#ff007f", "Focus", p.Palette.Focus, "#00f0ff", "Highlight", p.Palette.Highlight, "#ffe600", "Success", p.Palette.Success, "#3ddc84", "Error", p.Palette.Error, "#ff4d4d", "Dim", p.Palette.Dim, "#6b7280", "Text", p.Palette.Text, "#e5e7eb", "Selected", p.Palette.Selected, "#3b0f2a")
			if p.Palette.Selected != "" {
				v["color"] = p.Palette.Selected
			}
			if p.SlotsRows == nil {
				rows := make([][]string, len(p.Slots))
				cellStyles := make([][]string, len(p.Slots))
				styles := make([]string, len(p.Slots))
				for i, r := range p.Slots {
					rows[i] = []string{r.Mark, r.Name, r.Gap, r.Swatch, r.Space, r.Value}
					cellStyles[i] = []string{cuppa.CellStyleCode(r.MarkStyle), cuppa.CellStyleCode(r.NameStyle), cuppa.CellStyleCode(r.GapStyle), cuppa.CellStyleCode(r.SwatchStyle), cuppa.CellStyleCode(r.SpaceStyle), cuppa.CellStyleCode(r.ValueStyle)}
					styles[i] = cuppa.RowStyleName(r.Style)
				}
				v["rows"] = cuppa.JoinRows(rows)
				v["cellstyles"] = cuppa.JoinRows(cellStyles)
				v["styles"] = cuppa.JoinList(styles)
			}
			return true
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.label", Name: "Hint 1", X: 0, Y: 10, W: 100, H: 1, Props: map[string]string{"align": "left", "background": "", "bold": "false", "color": "#6b7280", "text": "Colours are #rgb or #rrggbb and apply as soon as you accept them.", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a"}, MinW: 1, MinH: 1, Fit: func(w, h int) (x, y, cw, ch int) {
			return 0, 10, cuppa.Cells((float64(w) * 100.0 / 100.0)), 1
		}},
		Apply: func(p Props, v map[string]string, w, h int) bool {
			if p.Palette.Dim != "" {
				v["color"] = p.Palette.Dim
			}
			return true
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.label", Name: "Hint 2", X: 0, Y: 11, W: 100, H: 1, Props: map[string]string{"align": "left", "background": "", "bold": "false", "color": "#6b7280", "text": "Save on the preferences screen to keep them.", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a"}, MinW: 1, MinH: 1, Fit: func(w, h int) (x, y, cw, ch int) {
			return 0, 11, cuppa.Cells((float64(w) * 100.0 / 100.0)), 1
		}},
		Apply: func(p Props, v map[string]string, w, h int) bool {
			if p.Palette.Dim != "" {
				v["color"] = p.Palette.Dim
			}
			return true
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.label", Name: "Status", X: 0, Y: 12, W: 100, H: 1, Props: map[string]string{"align": "left", "background": "", "bold": "false", "color": "#ff4d4d", "text": "Use a colour like #ff007f or #f07.", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a"}, MinW: 1, MinH: 1, Fit: func(w, h int) (x, y, cw, ch int) {
			return 0, 12, cuppa.Cells((float64(w) * 100.0 / 100.0)), 1
		}},
		Apply: func(p Props, v map[string]string, w, h int) bool {
			if p.Palette.Error != "" {
				v["color"] = p.Palette.Error
			}
			v["text"] = p.Status
			return p.ShowStatus
		},
	},
}...)

// coloursMasterProps is the props of the master's parts, from the screen's.
func masterProps(p Props) master.Props {
	return master.Props{Title: p.Title, Hints: p.Hints, ShowKeys: p.ShowKeys, Palette: master.Palette(p.Palette)}
}

// Colours draws the screen in an area of w by h cells for the values in p.
func View(p Props, w, h int) cuppa.Frame {
	return cuppa.DrawScreen(parts, p, p.Theme, w, h, background, env)
}

// ColoursSize is the size of area the screen was designed for.
func Size() (w, h int) { return width, height }

// ColoursLayout is the screen without its drawing: the frame has the regions of the
// visible components, so a program can size a model of its own before it is
// drawn into one of them.
func Layout(p Props, w, h int) cuppa.Frame {
	return cuppa.LayoutScreen(parts, p, p.Theme, w, h, env)
}

// ColoursHandle turns a click or a key into the screen's event. ok is false when
// the message raises none. f is the frame the screen last returned.
func Handle(f cuppa.Frame, msg tea.Msg) (Event, bool) {
	switch m := msg.(type) {
	case tea.MouseClickMsg:
		if r, ok := f.EventAt(m.X, m.Y); ok {
			switch r.Event {
			case "Hint":
				return HintEvent{X: m.X - r.X, Y: m.Y - r.Y}, true
			case "Pick":
				return PickEvent{X: m.X - r.X, Y: m.Y - r.Y}, true
			}
		}
	}
	_ = f
	return nil, false
}

// ColoursProps is everything the Colours screen takes from the program that owns it.
type Props struct {
	// Title is the Text of "Title".
	Title string
	// Hints is the Hints (key:label, comma separated) of "Key bar".
	Hints []HintsHint
	// ShowKeys is the whether "Key bar" is shown.
	ShowKeys bool
	// Slots is the Sample rows (; between rows, , between cells) of "Slots".
	Slots []SlotsRow
	// SlotsRows is the function that gives the rows of "Slots" for the size its place has (w, h cells) and the index of the first one; use Window to keep the selected row in view. It is used when it is not nil.
	SlotsRows func(w, h int) ([]SlotsRow, int)
	// Status is the Text of "Status".
	Status string
	// ShowStatus is the whether "Status" is shown.
	ShowStatus bool
	// Palette is the design's named colours; set one and every component that
	// uses it follows. An empty one keeps the design's colour.
	Palette Palette
	// Theme is the colours components follow; an empty one keeps the design's.
	Theme cuppa.Theme
}

// ColoursPalette is a field per named colour of the design.
type Palette struct {
	// Accent is the colour named "Accent".
	Accent string
	// Focus is the colour named "Focus".
	Focus string
	// Highlight is the colour named "Highlight".
	Highlight string
	// Success is the colour named "Success".
	Success string
	// Error is the colour named "Error".
	Error string
	// Dim is the colour named "Dim".
	Dim string
	// Text is the colour named "Text".
	Text string
	// Selected is the colour named "Selected".
	Selected string
}

// DefaultColoursPalette is the palette as it was designed.
func DefaultPalette() Palette {
	return Palette{
		Accent:    "#ff007f",
		Focus:     "#00f0ff",
		Highlight: "#ffe600",
		Success:   "#3ddc84",
		Error:     "#ff4d4d",
		Dim:       "#6b7280",
		Text:      "#e5e7eb",
		Selected:  "#3b0f2a",
	}
}

// ColoursSlotsRow is one row of Slots. Style says how the row is drawn.
type SlotsRow struct {
	// Mark is a cell.
	Mark string
	// Name is a cell.
	Name string
	// Gap is a cell.
	Gap string
	// Swatch is a colour.
	Swatch string
	// Space is a cell.
	Space string
	// Value is a cell.
	Value string
	// MarkStyle restyles the Mark cell of this row; the zero value keeps the column's style.
	MarkStyle cuppa.CellStyle
	// NameStyle restyles the Name cell of this row; the zero value keeps the column's style.
	NameStyle cuppa.CellStyle
	// GapStyle restyles the Gap cell of this row; the zero value keeps the column's style.
	GapStyle cuppa.CellStyle
	// SwatchStyle restyles the Swatch cell of this row; the zero value keeps the column's style.
	SwatchStyle cuppa.CellStyle
	// SpaceStyle restyles the Space cell of this row; the zero value keeps the column's style.
	SpaceStyle cuppa.CellStyle
	// ValueStyle restyles the Value cell of this row; the zero value keeps the column's style.
	ValueStyle cuppa.CellStyle
	// Style is how the row is drawn: RowNormal, RowSelected, RowDim or RowAccent.
	Style cuppa.RowStyle
}

// ColoursHintsHint is one hint of the key bar Hints, the master's.
type HintsHint = master.HintsHint

// ColoursHintsHintAt is the index of the hint of Hints under column x of the bar (X of the
// click's event), and false between hints.
func HintsHintAt(p Props, x int) (int, bool) {
	return master.HintsHintAt(masterProps(p), x)
}

// ColoursSlotsAt is the index in Slots of the row under screen row y, for a
// screen of w by h cells, and false where there is none. It reads the rows
// the way the drawing does, so a click is mapped with no arithmetic of the
// program's own.
func SlotsAt(p Props, w, h, y int) (int, bool) {
	r, ok := Layout(p, w, h).RegionNamed("Slots")
	if !ok || y < r.Y || y >= r.Y+r.H {
		return 0, false
	}
	rows, first := p.Slots, 0
	if p.SlotsRows != nil {
		rows, first = p.SlotsRows(r.W, r.H)
	}
	k := y - r.Y
	if k >= len(rows) {
		return 0, false
	}
	return first + k, true
}

// DefaultColoursProps is the screen as it was designed: each input holds the value
// it had in the design, unless the design gives the input an app default.
func Default() Props {
	return Props{
		Title:      " MVD · Colours",
		Hints:      []HintsHint{{Key: "↑↓", Label: "move"}, {Key: "enter", Label: "edit"}, {Key: "s", Label: "save"}, {Key: "esc", Label: "back"}},
		ShowKeys:   true,
		Slots:      []SlotsRow{{Mark: "▸ ", Name: "Accent (title)", Gap: "", Swatch: "#ff007f", Space: "", Value: "#ff007f", Style: cuppa.RowSelected}, {Mark: "  ", Name: "Focus (active border)", Gap: "", Swatch: "#00f0ff", Space: "", Value: "#00f0ff", Style: cuppa.RowNormal}, {Mark: "  ", Name: "Highlight (keys, bars)", Gap: "", Swatch: "#ffe600", Space: "", Value: "#ffe600", Style: cuppa.RowNormal}, {Mark: "  ", Name: "Success", Gap: "", Swatch: "#3ddc84", Space: "", Value: "#3ddc84", Style: cuppa.RowNormal}, {Mark: "  ", Name: "Error", Gap: "", Swatch: "#ff4d4d", Space: "", Value: "#ff4d4d", Style: cuppa.RowNormal}, {Mark: "  ", Name: "Dim (borders, hints)", Gap: "", Swatch: "#6b7280", Space: "", Value: "#6b7280", Style: cuppa.RowNormal}, {Mark: "  ", Name: "Text", Gap: "", Swatch: "#e5e7eb", Space: "", Value: "#e5e7eb", Style: cuppa.RowNormal}, {Mark: "  ", Name: "Selected row (background)", Gap: "", Swatch: "#3b0f2a", Space: "", Value: "#3b0f2a", Style: cuppa.RowNormal}},
		SlotsRows:  nil,
		Status:     "Use a colour like #ff007f or #f07.",
		ShowStatus: true,
		Palette:    DefaultPalette(),
	}
}

// ColoursEvent is what a click on a component or a key raises. The concrete types
// are the Colours... structs below.
type Event interface{ isEvent() }

// ColoursHint is raised by a click on "Key bar". X and Y are the cell inside the component that was
// clicked, and 0 for a key.
type HintEvent struct{ X, Y int }

func (HintEvent) isEvent() {}

// ColoursPick is raised by a click on "Slots". X and Y are the cell inside the component that was
// clicked, and 0 for a key.
type PickEvent struct{ X, Y int }

func (PickEvent) isEvent() {}

// ColoursInputs is the props by the input names of the design, as the live
// connection publishes them.
func Inputs(p Props) map[string]any {
	return map[string]any{
		"Title":       p.Title,
		"Hints":       p.Hints,
		"Show keys":   p.ShowKeys,
		"Slots":       p.Slots,
		"Status":      p.Status,
		"Show status": p.ShowStatus,
	}
}

// ColoursEventNamed is the event the design calls name, with the cell x, y inside
// the component that raised it; false when the screen has no such event.
func EventNamed(name string, x, y int) (Event, bool) {
	switch name {
	case "Hint":
		return HintEvent{X: x, Y: y}, true
	case "Pick":
		return PickEvent{X: x, Y: y}, true
	}
	return nil, false
}

// ColoursKeys are the keys of the screen, for a key bar.
var Keys = []cuppa.KeyHint{}

// The runtime's types a page's Props and Frame are made of, so the program that
// holds the page does not import the runtime.
type (
	CellStyle = cuppa.CellStyle
	RowStyle  = cuppa.RowStyle
	Frame     = cuppa.Frame
	Region    = cuppa.Region
	Theme     = cuppa.Theme
)

// The styles of a row.
const (
	RowNormal   = cuppa.RowNormal
	RowSelected = cuppa.RowSelected
	RowDim      = cuppa.RowDim
	RowAccent   = cuppa.RowAccent
)
