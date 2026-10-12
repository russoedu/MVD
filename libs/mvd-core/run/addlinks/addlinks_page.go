// Written by the Go export of Cuppa and maintained by hand since: change it here.

package addlinks

import (
	tea "charm.land/bubbletea/v2"
	_ "youtube-downloader/libs/mvd-core/internal/components/widgets"
	cuppa "youtube-downloader/libs/mvd-core/internal/cuppa"
)

const (
	width      = 120
	height     = 36
	background = ""
)

// addlinksEnv is what the layout expressions and conditions of the screen read.
func env(p Props, w, h int) cuppa.FitEnv {
	_ = p
	return cuppa.FitEnv{W: w, H: h, Inputs: func(name string) float64 {
		switch name {
		}
		return 0
	}}
}

var parts = []cuppa.ScreenPart[Props]{
	{
		Placed: cuppa.Placed{Kind: "lipgloss.box", Name: "Box", X: 8, Y: 12, W: 104, H: 12, Props: map[string]string{"blurColor": "240", "border": "rounded", "color": "#00f0ff", "focused": "true", "gradient": "", "right": "", "rightColor": "240", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a", "title": "Add links", "titleBold": "true", "titleColor": "#ff007f"}, MinW: 3, MinH: 3, Fit: func(w, h int) (x, y, cw, ch int) {
			return cuppa.Cells(cuppa.Div(((float64(w) * 100.0 / 100.0) - max(24.0, min(104.0, ((float64(w)*100.0/100.0)-4.0)))), 2.0)), cuppa.Cells(cuppa.Div(((float64(h) * 100.0 / 100.0) - 12.0), 2.0)), cuppa.Cells(max(24.0, min(104.0, ((float64(w)*100.0/100.0)-4.0)))), cuppa.Cells(12.0)
		}},
		Apply: func(p Props, v map[string]string, w, h int) bool {
			if p.Palette.Focus != "" {
				v["color"] = p.Palette.Focus
			}
			if p.Palette.Accent != "" {
				v["titleColor"] = p.Palette.Accent
			}
			return true
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.slot", Name: "Links view", X: 9, Y: 13, W: 102, H: 8, Props: map[string]string{"theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a", "view": "Paste or type playlist / video URLs, one per line"}, MinW: 1, MinH: 1, FitEnv: func(e cuppa.FitEnv) (x, y, cw, ch int) {
			w, h := e.W, e.H
			_, _ = w, h
			return cuppa.Cells((e.Rect("Box", "x") + 1.0)), cuppa.Cells((e.Rect("Box", "y") + 1.0)), cuppa.Cells((e.Rect("Box", "w") - 2.0)), cuppa.Cells(8.0)
		}},
		Apply: func(p Props, v map[string]string, w, h int) bool {
			v["view"] = p.LinksView
			return true
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.keybar", Name: "Key bar", X: 9, Y: 22, W: 102, H: 1, Props: map[string]string{"gap": "2", "hints": "ctrl+s:add,esc:cancel", "keyColor": "#ffe600", "labelColor": "#6b7280", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a"}, MinW: 4, MinH: 1, FitEnv: func(e cuppa.FitEnv) (x, y, cw, ch int) {
			w, h := e.W, e.H
			_, _ = w, h
			return cuppa.Cells((e.Rect("Box", "x") + 1.0)), cuppa.Cells((e.Rect("Box", "y") + 10.0)), cuppa.Cells((e.Rect("Box", "w") - 2.0)), 1
		}},
		Apply: func(p Props, v map[string]string, w, h int) bool {
			if p.Palette.Highlight != "" {
				v["keyColor"] = p.Palette.Highlight
			}
			if p.Palette.Dim != "" {
				v["labelColor"] = p.Palette.Dim
			}
			{
				items := make([]string, len(p.Hints))
				for i, h := range p.Hints {
					items[i] = h.Key + ":" + h.Label
				}
				v["hints"] = cuppa.JoinList(items)
			}
			return true
		},
	},
}

// AddLinks draws the screen in an area of w by h cells for the values in p.
func View(p Props, w, h int) cuppa.Frame {
	return cuppa.DrawScreen(parts, p, p.Theme, w, h, background, env)
}

// AddLinksSize is the size of area the screen was designed for.
func Size() (w, h int) { return width, height }

// AddLinksLayout is the screen without its drawing: the frame has the regions of the
// visible components, so a program can size a model of its own before it is
// drawn into one of them.
func Layout(p Props, w, h int) cuppa.Frame {
	return cuppa.LayoutScreen(parts, p, p.Theme, w, h, env)
}

// AddLinksHandle turns a click or a key into the screen's event. ok is false when
// the message raises none. f is the frame the screen last returned.
func Handle(f cuppa.Frame, msg tea.Msg) (Event, bool) {
	_, _ = f, msg
	return nil, false
}

// AddLinksProps is everything the AddLinks screen takes from the program that owns it.
type Props struct {
	// LinksView is the Sample view (| starts a line) of "Links view".
	LinksView string
	// Hints is the Hints (key:label, comma separated) of "Key bar".
	Hints []HintsHint
	// Palette is the design's named colours; set one and every component that
	// uses it follows. An empty one keeps the design's colour.
	Palette Palette
	// Theme is the colours components follow; an empty one keeps the design's.
	Theme cuppa.Theme
}

// AddLinksPalette is a field per named colour of the design.
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

// DefaultAddLinksPalette is the palette as it was designed.
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

// AddLinksHintsHint is one hint of the key bar Hints.
type HintsHint struct{ Key, Label string }

// AddLinksHintsHintAt is the index of the hint of Hints under column x of the bar (X of the
// click's event), and false between hints.
func HintsHintAt(p Props, x int) (int, bool) {
	keys := make([][2]string, len(p.Hints))
	for i, h := range p.Hints {
		keys[i] = [2]string{h.Key, h.Label}
	}
	return cuppa.HintIndexAt(keys, 2, x)
}

// DefaultAddLinksProps is the screen as it was designed: each input holds the value
// it had in the design, unless the design gives the input an app default.
func Default() Props {
	return Props{
		LinksView: "Paste or type playlist / video URLs, one per line",
		Hints:     []HintsHint{{Key: "ctrl+s", Label: "add"}, {Key: "esc", Label: "cancel"}},
		Palette:   DefaultPalette(),
	}
}

// AddLinksEvent is what a click on a component or a key raises. The concrete types
// are the AddLinks... structs below.
type Event interface{ isEvent() }

// AddLinksInputs is the props by the input names of the design, as the live
// connection publishes them.
func Inputs(p Props) map[string]any {
	return map[string]any{
		"LinksView": p.LinksView,
		"Hints":     p.Hints,
	}
}

// AddLinksEventNamed is the event the design calls name, with the cell x, y inside
// the component that raised it; false when the screen has no such event.
func EventNamed(name string, x, y int) (Event, bool) {
	switch name {
	}
	return nil, false
}

// AddLinksKeys are the keys of the screen, for a key bar.
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
