// Written by the Go export of Cuppa and maintained by hand since: change it here.

package download

import (
	"strconv"
	master "youtube-downloader/libs/mvd-core/internal/components/master"
	_ "youtube-downloader/libs/mvd-core/internal/components/widgets"
	cuppa "youtube-downloader/libs/mvd-core/internal/cuppa"

	tea "charm.land/bubbletea/v2"
)

const (
	width      = 120
	height     = 36
	background = ""
)

// downloadEnv is what the layout expressions and conditions of the screen read.
func env(p Props, w, h int) cuppa.FitEnv {
	_ = p
	return cuppa.FitEnv{W: w, H: h, Inputs: func(name string) float64 {
		switch name {
		case "Show keys":
			if p.ShowKeys {
				return 1
			}
			return 0
		case "Playlists focus":
			if p.PlaylistsFocus {
				return 1
			}
			return 0
		case "Entries focus":
			if p.EntriesFocus {
				return 1
			}
			return 0
		case "Output focus":
			if p.OutputFocus {
				return 1
			}
			return 0
		case "BarWidth":
			return float64(p.BarWidth)
		case "Percent":
			return p.Percent
		case "Show progress":
			if p.ShowProgress {
				return 1
			}
			return 0
		case "Show note":
			if p.ShowNote {
				return 1
			}
			return 0
		case "Confirm quit":
			if p.ConfirmQuit {
				return 1
			}
			return 0
		case "Show status":
			if p.ShowStatus {
				return 1
			}
			return 0
		case "LeftWidth":
			return float64(p.LeftWidth)
		case "MidWidth":
			return float64(p.MidWidth)
		case "Columns":
			return float64(p.Columns)
		case "LeftReserve":
			return float64(p.LeftReserve)
		case "PlaylistCount":
			return float64(p.PlaylistCount)
		case "HasNotes":
			return float64(p.HasNotes)
		case "FullLog":
			return float64(p.FullLog)
		case "FooterLines":
			return float64(p.FooterLines)
		}
		return 0
	}}
}

var parts = append(master.PartsFor(masterProps, "Title"), []cuppa.ScreenPart[Props]{
	{
		Placed: cuppa.Placed{Kind: "lipgloss.box", Name: "Header", X: 0, Y: 0, W: 120, H: 3, Props: map[string]string{"blurColor": "240", "border": "rounded", "color": "#6b7280", "focused": "true", "gradient": "", "right": "00:00:00", "rightColor": "#6b7280", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a", "title": "", "titleBold": "true", "titleColor": "#ff007f"}, MinW: 3, MinH: 3, Fit: func(w, h int) (x, y, cw, ch int) {
			return 0, 0, cuppa.Cells((float64(w) * 100.0 / 100.0)), 3
		}},
		Apply: func(p Props, v map[string]string, w, h int) bool {
			if p.Palette.Dim != "" {
				v["color"] = p.Palette.Dim
			}
			if p.Palette.Dim != "" {
				v["rightColor"] = p.Palette.Dim
			}
			if p.Palette.Accent != "" {
				v["titleColor"] = p.Palette.Accent
			}
			v["right"] = p.HeaderRight
			return true
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.label", Name: "Counters", X: 1, Y: 1, W: 118, H: 1, Props: map[string]string{"align": "left", "background": "", "bold": "false", "color": "#e5e7eb", "text": "Queue 0   Running 0   Done 0", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a"}, MinW: 1, MinH: 1, Fit: func(w, h int) (x, y, cw, ch int) {
			return 1, 1, cuppa.Cells(((float64(w) * 100.0 / 100.0) - 2.0)), 1
		}},
		Apply: func(p Props, v map[string]string, w, h int) bool {
			if p.Palette.Text != "" {
				v["color"] = p.Palette.Text
			}
			v["text"] = p.Counters
			return true
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.box", Name: "Playlists box", X: 0, Y: 3, W: 45, H: 8, Props: map[string]string{"blurColor": "#6b7280", "border": "rounded", "color": "#00f0ff", "focused": "false", "gradient": "", "right": "0", "rightColor": "#6b7280", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a", "title": "Links", "titleBold": "true", "titleColor": "#ff007f"}, MinW: 3, MinH: 3, FitEnv: func(e cuppa.FitEnv) (x, y, cw, ch int) {
			w, h := e.W, e.H
			_, _ = w, h
			if e.In("Columns") > 0.0 {
				return 0, 3, cuppa.Cells(e.In("LeftWidth")), cuppa.Cells(float64(h) - 4.0 - e.In("LeftReserve"))
			}
			return 0, 3, cuppa.Cells(e.In("LeftWidth")), cuppa.Cells(min(min((e.In("PlaylistCount")+2.0), max(5.0, cuppa.Div(((float64(h)*100.0/100.0)-4.0), 3.0))), ((float64(h) * 100.0 / 100.0) - 9.0)))
		}},
		Apply: func(p Props, v map[string]string, w, h int) bool {
			if p.Palette.Dim != "" {
				v["blurColor"] = p.Palette.Dim
			}
			if p.Palette.Focus != "" {
				v["color"] = p.Palette.Focus
			}
			if p.Palette.Dim != "" {
				v["rightColor"] = p.Palette.Dim
			}
			if p.Palette.Accent != "" {
				v["titleColor"] = p.Palette.Accent
			}
			v["focused"] = strconv.FormatBool(p.PlaylistsFocus)
			v["right"] = p.PlaylistsRight
			e := env(p, w, h)
			return (e.In("LeftWidth") > 0.0)
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.rows", Name: "Playlists rows", X: 1, Y: 4, W: 43, H: 6, Props: map[string]string{"blurColor": "", "cellstyles": "", "color": "#3b0f2a", "columns": "Mark:1:fg=@Text:selbg=@Selected:selfg=@Text:selbold,Sp:1:fg=@Text:selbg=@Selected:selfg=@Text:selbold,Title:fill:ellipsis:fg=@Text:selbg=@Selected:selfg=@Text:selbold,Sp2:1:fg=@Text:selbg=@Selected:selfg=@Text:selbold,Count:auto:fg=@Text:selbg=@Selected:selfg=@Text:selbold,Sp3:1:fg=@Text:selbg=@Selected:selfg=@Text:selbold,Glyph:1,End:1", "focused": "true", "gap": "0", "rows": "▸, ,90s UK Dance Hits, ,3/6, ,⠋,", "styles": "selected", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a"}, MinW: 8, MinH: 1, FitEnv: func(e cuppa.FitEnv) (x, y, cw, ch int) {
			w, h := e.W, e.H
			_, _ = w, h
			return cuppa.Cells((e.Rect("Playlists box", "x") + 1.0)), cuppa.Cells((e.Rect("Playlists box", "y") + 1.0)), cuppa.Cells((e.Rect("Playlists box", "w") - 2.0)), cuppa.Cells((e.Rect("Playlists box", "h") - 2.0))
		}},
		Event: "PickPlaylist",
		Late: func(p Props, v map[string]string, w, h int) {
			if p.PlaylistsRows != nil {
				rs, _ := p.PlaylistsRows(w, h)
				{
					rows := make([][]string, len(rs))
					cellStyles := make([][]string, len(rs))
					styles := make([]string, len(rs))
					for i, r := range rs {
						rows[i] = []string{r.Mark, r.Sp, r.Title, r.Sp2, r.Count, r.Sp3, r.Glyph, r.End}
						cellStyles[i] = []string{cuppa.CellStyleCode(r.MarkStyle), cuppa.CellStyleCode(r.SpStyle), cuppa.CellStyleCode(r.TitleStyle), cuppa.CellStyleCode(r.Sp2Style), cuppa.CellStyleCode(r.CountStyle), cuppa.CellStyleCode(r.Sp3Style), cuppa.CellStyleCode(r.GlyphStyle), cuppa.CellStyleCode(r.EndStyle)}
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
			if p.PlaylistsRows == nil {
				rows := make([][]string, len(p.Playlists))
				cellStyles := make([][]string, len(p.Playlists))
				styles := make([]string, len(p.Playlists))
				for i, r := range p.Playlists {
					rows[i] = []string{r.Mark, r.Sp, r.Title, r.Sp2, r.Count, r.Sp3, r.Glyph, r.End}
					cellStyles[i] = []string{cuppa.CellStyleCode(r.MarkStyle), cuppa.CellStyleCode(r.SpStyle), cuppa.CellStyleCode(r.TitleStyle), cuppa.CellStyleCode(r.Sp2Style), cuppa.CellStyleCode(r.CountStyle), cuppa.CellStyleCode(r.Sp3Style), cuppa.CellStyleCode(r.GlyphStyle), cuppa.CellStyleCode(r.EndStyle)}
					styles[i] = cuppa.RowStyleName(r.Style)
				}
				v["rows"] = cuppa.JoinRows(rows)
				v["cellstyles"] = cuppa.JoinRows(cellStyles)
				v["styles"] = cuppa.JoinList(styles)
			}
			e := env(p, w, h)
			return (e.In("LeftWidth") > 0.0)
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.box", Name: "Entries box", X: 0, Y: 11, W: 45, H: 25, Props: map[string]string{"blurColor": "#6b7280", "border": "rounded", "color": "#00f0ff", "focused": "true", "gradient": "", "right": "0", "rightColor": "#6b7280", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a", "title": "Entries", "titleBold": "true", "titleColor": "#ff007f"}, MinW: 3, MinH: 3, FitEnv: func(e cuppa.FitEnv) (x, y, cw, ch int) {
			w, h := e.W, e.H
			_, _ = w, h
			if e.In("Columns") > 0.0 {
				return cuppa.Cells(e.In("LeftWidth")), 3, cuppa.Cells(e.In("MidWidth")), cuppa.Cells(float64(h) - 4.0)
			}
			return 0, cuppa.Cells(e.Rect("Playlists box", "bottom")), cuppa.Cells(e.In("LeftWidth")), cuppa.Cells((((float64(h) * 100.0 / 100.0) - e.Rect("Playlists box", "bottom")) - 1.0))
		}},
		Apply: func(p Props, v map[string]string, w, h int) bool {
			if p.Palette.Dim != "" {
				v["blurColor"] = p.Palette.Dim
			}
			if p.Palette.Focus != "" {
				v["color"] = p.Palette.Focus
			}
			if p.Palette.Dim != "" {
				v["rightColor"] = p.Palette.Dim
			}
			if p.Palette.Accent != "" {
				v["titleColor"] = p.Palette.Accent
			}
			v["focused"] = strconv.FormatBool(p.EntriesFocus)
			v["right"] = p.EntriesRight
			v["title"] = p.EntriesTitle
			e := env(p, w, h)
			return (e.In("LeftWidth") > 0.0)
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.rows", Name: "Entries rows", X: 1, Y: 12, W: 43, H: 23, Props: map[string]string{"blurColor": "#ff007f", "cellstyles": "", "color": "#3b0f2a", "columns": "Glyph:1,Sp:1:fg=@Text:selbg=@Selected:selfg=@Text:selbold,Idx:2:fg=@Text:selbg=@Selected:selfg=@Text:selbold,Sp2:1:fg=@Text:selbg=@Selected:selfg=@Text:selbold,Title:fill:ellipsis:fg=@Text:selbg=@Selected:selfg=@Text:selbold,Sp3:1:fg=@Text:selbg=@Selected:selfg=@Text:selbold,Tag:auto", "focused": "true", "gap": "0", "rows": "✓, ,01, ,You Don't Know Me, ,⇄", "styles": "normal", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a"}, MinW: 8, MinH: 1, FitEnv: func(e cuppa.FitEnv) (x, y, cw, ch int) {
			w, h := e.W, e.H
			_, _ = w, h
			return cuppa.Cells((e.Rect("Entries box", "x") + 1.0)), cuppa.Cells((e.Rect("Entries box", "y") + 1.0)), cuppa.Cells((e.Rect("Entries box", "w") - 2.0)), cuppa.Cells((e.Rect("Entries box", "h") - 2.0))
		}},
		Event: "PickEntry",
		Late: func(p Props, v map[string]string, w, h int) {
			if p.EntriesRows != nil {
				rs, _ := p.EntriesRows(w, h)
				{
					rows := make([][]string, len(rs))
					cellStyles := make([][]string, len(rs))
					styles := make([]string, len(rs))
					for i, r := range rs {
						rows[i] = []string{r.Glyph, r.Sp, r.Idx, r.Sp2, r.Title, r.Sp3, r.Tag}
						cellStyles[i] = []string{cuppa.CellStyleCode(r.GlyphStyle), cuppa.CellStyleCode(r.SpStyle), cuppa.CellStyleCode(r.IdxStyle), cuppa.CellStyleCode(r.Sp2Style), cuppa.CellStyleCode(r.TitleStyle), cuppa.CellStyleCode(r.Sp3Style), cuppa.CellStyleCode(r.TagStyle)}
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
			if p.Palette.Accent != "" {
				v["blurColor"] = p.Palette.Accent
			}
			if p.Palette.Selected != "" {
				v["color"] = p.Palette.Selected
			}
			v["focused"] = strconv.FormatBool(p.EntriesFocus)
			if p.EntriesRows == nil {
				rows := make([][]string, len(p.Entries))
				cellStyles := make([][]string, len(p.Entries))
				styles := make([]string, len(p.Entries))
				for i, r := range p.Entries {
					rows[i] = []string{r.Glyph, r.Sp, r.Idx, r.Sp2, r.Title, r.Sp3, r.Tag}
					cellStyles[i] = []string{cuppa.CellStyleCode(r.GlyphStyle), cuppa.CellStyleCode(r.SpStyle), cuppa.CellStyleCode(r.IdxStyle), cuppa.CellStyleCode(r.Sp2Style), cuppa.CellStyleCode(r.TitleStyle), cuppa.CellStyleCode(r.Sp3Style), cuppa.CellStyleCode(r.TagStyle)}
					styles[i] = cuppa.RowStyleName(r.Style)
				}
				v["rows"] = cuppa.JoinRows(rows)
				v["cellstyles"] = cuppa.JoinRows(cellStyles)
				v["styles"] = cuppa.JoinList(styles)
			}
			e := env(p, w, h)
			return ((e.In("LeftWidth") > 0.0) && (e.In("HasNotes") == 0.0))
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.rows", Name: "Entry notes", X: 1, Y: 12, W: 43, H: 23, Props: map[string]string{"blurColor": "", "cellstyles": "", "color": "#3b0f2a", "columns": "Line:fill:ellipsis:fg=@Text", "focused": "true", "gap": "0", "rows": "note", "styles": "normal", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a"}, MinW: 8, MinH: 1, FitEnv: func(e cuppa.FitEnv) (x, y, cw, ch int) {
			w, h := e.W, e.H
			_, _ = w, h
			return cuppa.Cells((e.Rect("Entries box", "x") + 1.0)), cuppa.Cells((e.Rect("Entries box", "y") + 1.0)), cuppa.Cells((e.Rect("Entries box", "w") - 2.0)), cuppa.Cells((e.Rect("Entries box", "h") - 2.0))
		}},
		Late: func(p Props, v map[string]string, w, h int) {
			if p.NotesRows != nil {
				rs, _ := p.NotesRows(w, h)
				{
					rows := make([][]string, len(rs))
					cellStyles := make([][]string, len(rs))
					styles := make([]string, len(rs))
					for i, r := range rs {
						rows[i] = []string{r.Line}
						cellStyles[i] = []string{cuppa.CellStyleCode(r.LineStyle)}
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
			if p.NotesRows == nil {
				rows := make([][]string, len(p.Notes))
				cellStyles := make([][]string, len(p.Notes))
				styles := make([]string, len(p.Notes))
				for i, r := range p.Notes {
					rows[i] = []string{r.Line}
					cellStyles[i] = []string{cuppa.CellStyleCode(r.LineStyle)}
					styles[i] = cuppa.RowStyleName(r.Style)
				}
				v["rows"] = cuppa.JoinRows(rows)
				v["cellstyles"] = cuppa.JoinRows(cellStyles)
				v["styles"] = cuppa.JoinList(styles)
			}
			e := env(p, w, h)
			return ((e.In("LeftWidth") > 0.0) && (e.In("HasNotes") > 0.0))
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.box", Name: "Output box", X: 45, Y: 3, W: 75, H: 32, Props: map[string]string{"blurColor": "#6b7280", "border": "rounded", "color": "#00f0ff", "focused": "false", "gradient": "", "right": "follow", "rightColor": "#6b7280", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a", "title": "Output", "titleBold": "true", "titleColor": "#ff007f"}, MinW: 3, MinH: 3, FitEnv: func(e cuppa.FitEnv) (x, y, cw, ch int) {
			w, h := e.W, e.H
			_, _ = w, h
			return cuppa.Cells(e.In("LeftWidth") + e.In("MidWidth")), cuppa.Cells(3.0), cuppa.Cells(((float64(w) * 100.0 / 100.0) - e.In("LeftWidth") - e.In("MidWidth"))), cuppa.Cells(((float64(h) * 100.0 / 100.0) - 4.0))
		}},
		Apply: func(p Props, v map[string]string, w, h int) bool {
			if p.Palette.Dim != "" {
				v["blurColor"] = p.Palette.Dim
			}
			if p.Palette.Focus != "" {
				v["color"] = p.Palette.Focus
			}
			if p.Palette.Dim != "" {
				v["rightColor"] = p.Palette.Dim
			}
			if p.Palette.Accent != "" {
				v["titleColor"] = p.Palette.Accent
			}
			v["focused"] = strconv.FormatBool(p.OutputFocus)
			v["right"] = p.OutputRight
			v["title"] = p.OutputTitle
			e := env(p, w, h)
			return ((float64(e.W) >= 100.0) || (e.In("FullLog") > 0.0))
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.rows", Name: "Output rows", X: 46, Y: 4, W: 73, H: 30, Props: map[string]string{"blurColor": "", "cellstyles": "", "color": "#3b0f2a", "columns": "Line:fill:ellipsis:fg=@Text", "focused": "true", "gap": "0", "rows": "line", "styles": "normal", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a"}, MinW: 8, MinH: 1, FitEnv: func(e cuppa.FitEnv) (x, y, cw, ch int) {
			w, h := e.W, e.H
			_, _ = w, h
			return cuppa.Cells((e.Rect("Output box", "x") + 1.0)), cuppa.Cells((e.Rect("Output box", "y") + 1.0)), cuppa.Cells((e.Rect("Output box", "w") - 2.0)), cuppa.Cells(((e.Rect("Output box", "h") - 2.0) - e.In("FooterLines")))
		}},
		Late: func(p Props, v map[string]string, w, h int) {
			if p.OutputRows != nil {
				rs, _ := p.OutputRows(w, h)
				{
					rows := make([][]string, len(rs))
					cellStyles := make([][]string, len(rs))
					styles := make([]string, len(rs))
					for i, r := range rs {
						rows[i] = []string{r.Line}
						cellStyles[i] = []string{cuppa.CellStyleCode(r.LineStyle)}
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
			if p.OutputRows == nil {
				rows := make([][]string, len(p.Output))
				cellStyles := make([][]string, len(p.Output))
				styles := make([]string, len(p.Output))
				for i, r := range p.Output {
					rows[i] = []string{r.Line}
					cellStyles[i] = []string{cuppa.CellStyleCode(r.LineStyle)}
					styles[i] = cuppa.RowStyleName(r.Style)
				}
				v["rows"] = cuppa.JoinRows(rows)
				v["cellstyles"] = cuppa.JoinRows(cellStyles)
				v["styles"] = cuppa.JoinList(styles)
			}
			e := env(p, w, h)
			return ((float64(e.W) >= 100.0) || (e.In("FullLog") > 0.0))
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.progressline", Name: "Progress", X: 46, Y: 33, W: 73, H: 1, Props: map[string]string{"barWidth": "32", "empty": "░", "emptyColor": "#6b7280", "full": "█", "fullColor": "#ffe600", "gap": "2", "suffix": "34.2% of 112.4MiB at 8.1MiB/s ETA 0:09", "suffixColor": "#ffe600", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a", "value": "34.2"}, MinW: 4, MinH: 1, FitEnv: func(e cuppa.FitEnv) (x, y, cw, ch int) {
			w, h := e.W, e.H
			_, _ = w, h
			return cuppa.Cells((e.Rect("Output box", "x") + 1.0)), cuppa.Cells((e.Rect("Output box", "bottom") - 2.0)), cuppa.Cells((e.Rect("Output box", "w") - 2.0)), 1
		}},
		Apply: func(p Props, v map[string]string, w, h int) bool {
			if p.Palette.Dim != "" {
				v["emptyColor"] = p.Palette.Dim
			}
			if p.Palette.Highlight != "" {
				v["fullColor"] = p.Palette.Highlight
			}
			if p.Palette.Highlight != "" {
				v["suffixColor"] = p.Palette.Highlight
			}
			v["barWidth"] = strconv.Itoa(p.BarWidth)
			v["suffix"] = p.ProgressText
			v["value"] = strconv.FormatFloat(p.Percent, 'f', -1, 64)
			return p.ShowProgress
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.label", Name: "Note", X: 46, Y: 33, W: 73, H: 1, Props: map[string]string{"align": "left", "background": "", "bold": "false", "color": "#e5e7eb", "text": "✓ downloaded", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a"}, MinW: 1, MinH: 1, FitEnv: func(e cuppa.FitEnv) (x, y, cw, ch int) {
			w, h := e.W, e.H
			_, _ = w, h
			return cuppa.Cells((e.Rect("Output box", "x") + 1.0)), cuppa.Cells((e.Rect("Output box", "bottom") - 2.0)), cuppa.Cells((e.Rect("Output box", "w") - 2.0)), 1
		}},
		Apply: func(p Props, v map[string]string, w, h int) bool {
			if p.Palette.Text != "" {
				v["color"] = p.Palette.Text
			}
			v["text"] = p.Note
			return p.ShowNote
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.label", Name: "Confirm", X: 0, Y: 35, W: 120, H: 1, Props: map[string]string{"align": "left", "background": "", "bold": "false", "color": "#ff4d4d", "text": " Downloads are still running. Quit? y/n", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a"}, MinW: 1, MinH: 1, Fit: func(w, h int) (x, y, cw, ch int) {
			return 0, cuppa.Cells(((float64(h) * 100.0 / 100.0) - 1.0)), cuppa.Cells((float64(w) * 100.0 / 100.0)), 1
		}},
		Apply: func(p Props, v map[string]string, w, h int) bool {
			if p.Palette.Error != "" {
				v["color"] = p.Palette.Error
			}
			v["text"] = p.Confirm
			return p.ConfirmQuit
		},
	},
	{
		Placed: cuppa.Placed{Kind: "lipgloss.label", Name: "Status", X: 0, Y: 35, W: 120, H: 1, Props: map[string]string{"align": "left", "background": "", "bold": "false", "color": "#ff007f", "text": " status", "theme.muted": "", "theme.palette": "Accent=#ff007f;Focus=#00f0ff;Highlight=#ffe600;Success=#3ddc84;Error=#ff4d4d;Dim=#6b7280;Text=#e5e7eb;Selected=#3b0f2a"}, MinW: 1, MinH: 1, Fit: func(w, h int) (x, y, cw, ch int) {
			return 0, cuppa.Cells(((float64(h) * 100.0 / 100.0) - 1.0)), cuppa.Cells((float64(w) * 100.0 / 100.0)), 1
		}},
		Apply: func(p Props, v map[string]string, w, h int) bool {
			if p.Palette.Accent != "" {
				v["color"] = p.Palette.Accent
			}
			v["text"] = p.Status
			return p.ShowStatus
		},
	},
}...)

// downloadMasterProps is the props of the master's parts, from the screen's.
func masterProps(p Props) master.Props {
	return master.Props{Hints: p.Hints, ShowKeys: p.ShowKeys, Palette: master.Palette(p.Palette)}
}

// Download draws the screen in an area of w by h cells for the values in p.
func View(p Props, w, h int) cuppa.Frame {
	return cuppa.DrawScreen(parts, p, p.Theme, w, h, background, env)
}

// DownloadSize is the size of area the screen was designed for.
func Size() (w, h int) { return width, height }

// DownloadLayout is the screen without its drawing: the frame has the regions of the
// visible components, so a program can size a model of its own before it is
// drawn into one of them.
func Layout(p Props, w, h int) cuppa.Frame {
	return cuppa.LayoutScreen(parts, p, p.Theme, w, h, env)
}

// DownloadHandle turns a click or a key into the screen's event. ok is false when
// the message raises none. f is the frame the screen last returned.
func Handle(f cuppa.Frame, msg tea.Msg) (Event, bool) {
	switch m := msg.(type) {
	case tea.MouseClickMsg:
		if r, ok := f.EventAt(m.X, m.Y); ok {
			switch r.Event {
			case "Hint":
				return HintEvent{X: m.X - r.X, Y: m.Y - r.Y}, true
			case "PickPlaylist":
				return PickPlaylistEvent{X: m.X - r.X, Y: m.Y - r.Y}, true
			case "PickEntry":
				return PickEntryEvent{X: m.X - r.X, Y: m.Y - r.Y}, true
			}
		}
	}
	_ = f
	return nil, false
}

// DownloadProps is everything the Download screen takes from the program that owns it.
type Props struct {
	// Hints is the Hints (key:label, comma separated) of "Key bar".
	Hints []HintsHint
	// ShowKeys is the whether "Key bar" is shown.
	ShowKeys bool
	// HeaderRight is the Right label (in the top edge) of "Header".
	HeaderRight string
	// Counters is the Text of "Counters".
	Counters string
	// PlaylistsFocus is the Has focus (false: the border takes the blurred colour) of "Playlists box".
	PlaylistsFocus bool
	// PlaylistsRight is the Right label (in the top edge) of "Playlists box".
	PlaylistsRight string
	// Playlists is the Sample rows (; between rows, , between cells) of "Playlists rows".
	Playlists []PlaylistsRow
	// PlaylistsRows is the function that gives the rows of "Playlists rows" for the size its place has (w, h cells) and the index of the first one; use Window to keep the selected row in view. It is used when it is not nil.
	PlaylistsRows func(w, h int) ([]PlaylistsRow, int)
	// EntriesFocus is the Has focus (false: the border takes the blurred colour) of "Entries box", Has focus (false: a selected row is accent text, not inverted) of "Entries rows".
	EntriesFocus bool
	// EntriesRight is the Right label (in the top edge) of "Entries box".
	EntriesRight string
	// EntriesTitle is the Title of "Entries box".
	EntriesTitle string
	// Entries is the Sample rows (; between rows, , between cells) of "Entries rows".
	Entries []EntriesRow
	// EntriesRows is the function that gives the rows of "Entries rows" for the size its place has (w, h cells) and the index of the first one; use Window to keep the selected row in view. It is used when it is not nil.
	EntriesRows func(w, h int) ([]EntriesRow, int)
	// Notes is the Sample rows (; between rows, , between cells) of "Entry notes".
	Notes []NotesRow
	// NotesRows is the function that gives the rows of "Entry notes" for the size its place has (w, h cells) and the index of the first one; use Window to keep the selected row in view. It is used when it is not nil.
	NotesRows func(w, h int) ([]NotesRow, int)
	// OutputFocus is the Has focus (false: the border takes the blurred colour) of "Output box".
	OutputFocus bool
	// OutputRight is the Right label (in the top edge) of "Output box".
	OutputRight string
	// OutputTitle is the Title of "Output box".
	OutputTitle string
	// Output is the Sample rows (; between rows, , between cells) of "Output rows".
	Output []OutputRow
	// OutputRows is the function that gives the rows of "Output rows" for the size its place has (w, h cells) and the index of the first one; use Window to keep the selected row in view. It is used when it is not nil.
	OutputRows func(w, h int) ([]OutputRow, int)
	// BarWidth is the Bar width (0: what the text leaves) of "Progress".
	BarWidth int
	// ProgressText is the Text after the bar of "Progress".
	ProgressText string
	// Percent is the Done (0 to 100) of "Progress".
	Percent float64
	// ShowProgress is the whether "Progress" is shown.
	ShowProgress bool
	// Note is the Text of "Note".
	Note string
	// ShowNote is the whether "Note" is shown.
	ShowNote bool
	// Confirm is the Text of "Confirm".
	Confirm string
	// ConfirmQuit is the whether "Confirm" is shown.
	ConfirmQuit bool
	// Status is the Text of "Status".
	Status string
	// ShowStatus is the whether "Status" is shown.
	ShowStatus bool
	// LeftWidth is the read by a layout expression or a condition.
	LeftWidth int
	// MidWidth is how wide the entries column is when the screen is in columns, else 0.
	MidWidth int
	// Columns is 1 when the playlists, the entries and the output are side by side.
	Columns int
	// LeftReserve is the rows the app keeps under the playlists, for its own buttons.
	LeftReserve int
	// PlaylistCount is the read by a layout expression or a condition.
	PlaylistCount int
	// HasNotes is the read by a layout expression or a condition.
	HasNotes int
	// FullLog is the read by a layout expression or a condition.
	FullLog int
	// FooterLines is the read by a layout expression or a condition.
	FooterLines int
	// Palette is the design's named colours; set one and every component that
	// uses it follows. An empty one keeps the design's colour.
	Palette Palette
	// Theme is the colours components follow; an empty one keeps the design's.
	Theme cuppa.Theme
}

// DownloadPalette is a field per named colour of the design.
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

// DefaultDownloadPalette is the palette as it was designed.
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

// DownloadPlaylistsRow is one row of Playlists. Style says how the row is drawn.
type PlaylistsRow struct {
	// Mark is a cell.
	Mark string
	// Sp is a cell.
	Sp string
	// Title is a cell.
	Title string
	// Sp2 is a cell.
	Sp2 string
	// Count is a cell.
	Count string
	// Sp3 is a cell.
	Sp3 string
	// Glyph is a cell.
	Glyph string
	// End is a cell.
	End string
	// MarkStyle restyles the Mark cell of this row; the zero value keeps the column's style.
	MarkStyle cuppa.CellStyle
	// SpStyle restyles the Sp cell of this row; the zero value keeps the column's style.
	SpStyle cuppa.CellStyle
	// TitleStyle restyles the Title cell of this row; the zero value keeps the column's style.
	TitleStyle cuppa.CellStyle
	// Sp2Style restyles the Sp2 cell of this row; the zero value keeps the column's style.
	Sp2Style cuppa.CellStyle
	// CountStyle restyles the Count cell of this row; the zero value keeps the column's style.
	CountStyle cuppa.CellStyle
	// Sp3Style restyles the Sp3 cell of this row; the zero value keeps the column's style.
	Sp3Style cuppa.CellStyle
	// GlyphStyle restyles the Glyph cell of this row; the zero value keeps the column's style.
	GlyphStyle cuppa.CellStyle
	// EndStyle restyles the End cell of this row; the zero value keeps the column's style.
	EndStyle cuppa.CellStyle
	// Style is how the row is drawn: RowNormal, RowSelected, RowDim or RowAccent.
	Style cuppa.RowStyle
}

// DownloadEntriesRow is one row of Entries. Style says how the row is drawn.
type EntriesRow struct {
	// Glyph is a cell.
	Glyph string
	// Sp is a cell.
	Sp string
	// Idx is a cell.
	Idx string
	// Sp2 is a cell.
	Sp2 string
	// Title is a cell.
	Title string
	// Sp3 is a cell.
	Sp3 string
	// Tag is a cell.
	Tag string
	// GlyphStyle restyles the Glyph cell of this row; the zero value keeps the column's style.
	GlyphStyle cuppa.CellStyle
	// SpStyle restyles the Sp cell of this row; the zero value keeps the column's style.
	SpStyle cuppa.CellStyle
	// IdxStyle restyles the Idx cell of this row; the zero value keeps the column's style.
	IdxStyle cuppa.CellStyle
	// Sp2Style restyles the Sp2 cell of this row; the zero value keeps the column's style.
	Sp2Style cuppa.CellStyle
	// TitleStyle restyles the Title cell of this row; the zero value keeps the column's style.
	TitleStyle cuppa.CellStyle
	// Sp3Style restyles the Sp3 cell of this row; the zero value keeps the column's style.
	Sp3Style cuppa.CellStyle
	// TagStyle restyles the Tag cell of this row; the zero value keeps the column's style.
	TagStyle cuppa.CellStyle
	// Style is how the row is drawn: RowNormal, RowSelected, RowDim or RowAccent.
	Style cuppa.RowStyle
}

// DownloadNotesRow is one row of Notes. Style says how the row is drawn.
type NotesRow struct {
	// Line is a cell.
	Line string
	// LineStyle restyles the Line cell of this row; the zero value keeps the column's style.
	LineStyle cuppa.CellStyle
	// Style is how the row is drawn: RowNormal, RowSelected, RowDim or RowAccent.
	Style cuppa.RowStyle
}

// DownloadOutputRow is one row of Output. Style says how the row is drawn.
type OutputRow struct {
	// Line is a cell.
	Line string
	// LineStyle restyles the Line cell of this row; the zero value keeps the column's style.
	LineStyle cuppa.CellStyle
	// Style is how the row is drawn: RowNormal, RowSelected, RowDim or RowAccent.
	Style cuppa.RowStyle
}

// DownloadHintsHint is one hint of the key bar Hints, the master's.
type HintsHint = master.HintsHint

// DownloadHintsHintAt is the index of the hint of Hints under column x of the bar (X of the
// click's event), and false between hints.
func HintsHintAt(p Props, x int) (int, bool) {
	return master.HintsHintAt(masterProps(p), x)
}

// DownloadPlaylistsAt is the index in Playlists of the row under screen row y, for a
// screen of w by h cells, and false where there is none. It reads the rows
// the way the drawing does, so a click is mapped with no arithmetic of the
// program's own.
func PlaylistsAt(p Props, w, h, y int) (int, bool) {
	r, ok := Layout(p, w, h).RegionNamed("Playlists rows")
	if !ok || y < r.Y || y >= r.Y+r.H {
		return 0, false
	}
	rows, first := p.Playlists, 0
	if p.PlaylistsRows != nil {
		rows, first = p.PlaylistsRows(r.W, r.H)
	}
	k := y - r.Y
	if k >= len(rows) {
		return 0, false
	}
	return first + k, true
}

// DownloadEntriesAt is the index in Entries of the row under screen row y, for a
// screen of w by h cells, and false where there is none. It reads the rows
// the way the drawing does, so a click is mapped with no arithmetic of the
// program's own.
func EntriesAt(p Props, w, h, y int) (int, bool) {
	r, ok := Layout(p, w, h).RegionNamed("Entries rows")
	if !ok || y < r.Y || y >= r.Y+r.H {
		return 0, false
	}
	rows, first := p.Entries, 0
	if p.EntriesRows != nil {
		rows, first = p.EntriesRows(r.W, r.H)
	}
	k := y - r.Y
	if k >= len(rows) {
		return 0, false
	}
	return first + k, true
}

// DownloadNotesAt is the index in Notes of the row under screen row y, for a
// screen of w by h cells, and false where there is none. It reads the rows
// the way the drawing does, so a click is mapped with no arithmetic of the
// program's own.
func NotesAt(p Props, w, h, y int) (int, bool) {
	r, ok := Layout(p, w, h).RegionNamed("Entry notes")
	if !ok || y < r.Y || y >= r.Y+r.H {
		return 0, false
	}
	rows, first := p.Notes, 0
	if p.NotesRows != nil {
		rows, first = p.NotesRows(r.W, r.H)
	}
	k := y - r.Y
	if k >= len(rows) {
		return 0, false
	}
	return first + k, true
}

// DownloadOutputAt is the index in Output of the row under screen row y, for a
// screen of w by h cells, and false where there is none. It reads the rows
// the way the drawing does, so a click is mapped with no arithmetic of the
// program's own.
func OutputAt(p Props, w, h, y int) (int, bool) {
	r, ok := Layout(p, w, h).RegionNamed("Output rows")
	if !ok || y < r.Y || y >= r.Y+r.H {
		return 0, false
	}
	rows, first := p.Output, 0
	if p.OutputRows != nil {
		rows, first = p.OutputRows(r.W, r.H)
	}
	k := y - r.Y
	if k >= len(rows) {
		return 0, false
	}
	return first + k, true
}

// DefaultDownloadProps is the screen as it was designed: each input holds the value
// it had in the design, unless the design gives the input an app default.
func Default() Props {
	return Props{
		Hints:          []HintsHint{{Key: "↑↓", Label: "move"}, {Key: "tab", Label: "pane"}, {Key: "⏎", Label: "follow"}, {Key: "f", Label: "failed"}, {Key: "r", Label: "retry"}, {Key: "l", Label: "full log"}, {Key: "?", Label: "help"}, {Key: "q", Label: "quit"}},
		ShowKeys:       false,
		HeaderRight:    "00:00:00",
		Counters:       "Queue 0   Running 0   Done 0",
		PlaylistsFocus: false,
		PlaylistsRight: "0",
		Playlists:      []PlaylistsRow{{Mark: "▸", Sp: " ", Title: "90s UK Dance Hits", Sp2: " ", Count: "3/6", Sp3: " ", Glyph: "⠋", End: "", Style: cuppa.RowSelected}},
		PlaylistsRows:  nil,
		EntriesFocus:   true,
		EntriesRight:   "0",
		EntriesTitle:   "Songs",
		Entries:        []EntriesRow{{Glyph: "✓", Sp: " ", Idx: "01", Sp2: " ", Title: "You Don't Know Me", Sp3: " ", Tag: "⇄", Style: cuppa.RowNormal}},
		EntriesRows:    nil,
		Notes:          []NotesRow{{Line: "note", Style: cuppa.RowNormal}},
		NotesRows:      nil,
		OutputFocus:    false,
		OutputRight:    "follow",
		OutputTitle:    "Details",
		Output:         []OutputRow{{Line: "line", Style: cuppa.RowNormal}},
		OutputRows:     nil,
		BarWidth:       32,
		ProgressText:   "34.2% of 112.4MiB at 8.1MiB/s ETA 0:09",
		Percent:        34.2,
		ShowProgress:   false,
		Note:           "✓ downloaded",
		ShowNote:       false,
		Confirm:        " Downloads are still running. Quit? y/n",
		ConfirmQuit:    false,
		Status:         " status",
		ShowStatus:     false,
		LeftWidth:      45,
		PlaylistCount:  6,
		HasNotes:       0,
		FullLog:        0,
		FooterLines:    0,
		Palette:        DefaultPalette(),
	}
}

// DownloadEvent is what a click on a component or a key raises. The concrete types
// are the Download... structs below.
type Event interface{ isEvent() }

// DownloadHint is raised by a click on "Key bar". X and Y are the cell inside the component that was
// clicked, and 0 for a key.
type HintEvent struct{ X, Y int }

func (HintEvent) isEvent() {}

// DownloadPickPlaylist is raised by a click on "Playlists rows". X and Y are the cell inside the component that was
// clicked, and 0 for a key.
type PickPlaylistEvent struct{ X, Y int }

func (PickPlaylistEvent) isEvent() {}

// DownloadPickEntry is raised by a click on "Entries rows". X and Y are the cell inside the component that was
// clicked, and 0 for a key.
type PickEntryEvent struct{ X, Y int }

func (PickEntryEvent) isEvent() {}

// DownloadInputs is the props by the input names of the design, as the live
// connection publishes them.
func Inputs(p Props) map[string]any {
	return map[string]any{
		"Hints":           p.Hints,
		"Show keys":       p.ShowKeys,
		"HeaderRight":     p.HeaderRight,
		"Counters":        p.Counters,
		"Playlists focus": p.PlaylistsFocus,
		"PlaylistsRight":  p.PlaylistsRight,
		"Playlists":       p.Playlists,
		"Entries focus":   p.EntriesFocus,
		"EntriesRight":    p.EntriesRight,
		"EntriesTitle":    p.EntriesTitle,
		"Entries":         p.Entries,
		"Notes":           p.Notes,
		"Output focus":    p.OutputFocus,
		"OutputRight":     p.OutputRight,
		"OutputTitle":     p.OutputTitle,
		"Output":          p.Output,
		"BarWidth":        p.BarWidth,
		"ProgressText":    p.ProgressText,
		"Percent":         p.Percent,
		"Show progress":   p.ShowProgress,
		"Note":            p.Note,
		"Show note":       p.ShowNote,
		"Confirm":         p.Confirm,
		"Confirm quit":    p.ConfirmQuit,
		"Status":          p.Status,
		"Show status":     p.ShowStatus,
		"LeftWidth":       p.LeftWidth,
		"MidWidth":        p.MidWidth,
		"Columns":         p.Columns,
		"LeftReserve":     p.LeftReserve,
		"PlaylistCount":   p.PlaylistCount,
		"HasNotes":        p.HasNotes,
		"FullLog":         p.FullLog,
		"FooterLines":     p.FooterLines,
	}
}

// DownloadEventNamed is the event the design calls name, with the cell x, y inside
// the component that raised it; false when the screen has no such event.
func EventNamed(name string, x, y int) (Event, bool) {
	switch name {
	case "Hint":
		return HintEvent{X: x, Y: y}, true
	case "PickPlaylist":
		return PickPlaylistEvent{X: x, Y: y}, true
	case "PickEntry":
		return PickEntryEvent{X: x, Y: y}, true
	}
	return nil, false
}

// DownloadKeys are the keys of the screen, for a key bar.
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
