package editor

import (
	"youtube-downloader/libs/mvd-core/connect"
	"fmt"
	"strings"
	"time"
	"youtube-downloader/libs/mvd-core/cells"
	editorplanningpage "youtube-downloader/libs/mvd-core/editor/editorplanning"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/outline"
	"youtube-downloader/libs/mvd-core/pointer"
	"youtube-downloader/libs/mvd-core/screenview"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlisteditor"
	"youtube-downloader/libs/mvd-core/playlistfile"
)

// editorDetailLines is how many lines the panel under the list takes.
const editorDetailLines = 4

// listHeight is how many songs fit on the screen: the title, the column names, a rule, the panel and the
// key bar take the rest.
func (m Model) listHeight() int {
	return max(3, m.height-(editorDetailLines+4))
}

// Render draws the editor.
func (m Model) Render() string {
	width := max(m.width, 20)
	if m.planning {
		return screenview.Padded(editorplanningpage.View(m.planningProps(width), width, m.height).View, width)
	}
	return screenview.Padded(View(m.editorProps(width), width, m.height).View, width)
}

// Live is the page the editor draws now and its inputs, by the input names of the design,
// for the live connection to the designer.
func (m Model) Live() (page string, inputs map[string]any) {
	width := max(m.width, 20)
	if m.planning {
		return "EditorPlanning", connect.Plain(editorplanningpage.Inputs(m.planningProps(width)))
	}
	return "Editor", connect.Plain(Inputs(m.editorProps(width)))
}

// editorProps fills the page of the song list.
func (m Model) editorProps(width int) Props {
	props := Default()
	props.Palette = Palette(look.CurrentPalette())
	props.Title = " " + m.title()
	props.Songs = m.songRows(width)
	props.ShowEmpty = len(m.rows) == 0

	detail := m.detail(width)
	props.Detail1, props.Detail2, props.Detail3, props.Detail4 = cells.CellText(detail[0]), cells.CellText(detail[1]), cells.CellText(detail[2]), cells.CellText(detail[3])

	line, keys := m.bottomParts(width)
	props.Hints = keybar.Of[HintsHint](m.hints())
	props.ShowKeys, props.ShowLine = keys, !keys
	props.Bottom = line
	return props
}

// planningProps fills the page shown while the plan runs.
func (m Model) planningProps(width int) editorplanningpage.Props {
	t := m.state.Tally()
	done := t.Planned + t.NotFound + t.Failed
	body := []string{
		" ",
		" " + look.Cyan.Render(look.SpinnerFrames[m.spin]) + " " + look.Text.Render("Looking up the best version of each song, nothing is downloaded."),
		" ",
	}
	if t.Total == 0 {
		body = append(body, look.Dim.Render(" Reading the playlists…"))
	} else {
		body = append(body, look.Dim.Render(fmt.Sprintf(" %d of %d songs looked up", done, t.Total)))
	}
	for _, pl := range m.state.Playlists {
		if pl.Err != "" {
			body = append(body, " "+look.Red.Render(cells.Display(pl.Title)+": "+pl.Err))
		}
	}

	props := editorplanningpage.Default()
	props.Palette = editorplanningpage.Palette(look.CurrentPalette())
	props.Body = nil
	for _, l := range body {
		props.Body = append(props.Body, editorplanningpage.BodyRow{Line: l})
	}
	line, keys := m.bottomParts(width)
	props.Hints = keybar.Of[editorplanningpage.HintsHint](m.hints())
	props.ShowKeys, props.ShowLine = keys, !keys
	props.Bottom = line
	return props
}

func (m Model) title() string {
	var downloaded, revised int
	for _, r := range m.rows {
		switch {
		case r.Downloaded:
			downloaded++
		case playlisteditor.Revised(r) && r.Decision != playlistfile.DecisionSkipped:
			revised++
		}
	}
	parts := []string{fmt.Sprintf("%d songs", len(m.rows))}
	if revised > 0 {
		parts = append(parts, fmt.Sprintf("%d to download", revised))
	}
	if n := playlisteditor.Unreviewed(m.rows); n > 0 {
		parts = append(parts, fmt.Sprintf("%d not reviewed", n))
	}
	if downloaded > 0 {
		parts = append(parts, fmt.Sprintf("%d downloaded", downloaded))
	}
	return "MVD · Review before downloading  ·  " + strings.Join(parts, " · ")
}

// rowName is what a song is called on its line: the artist and the title a music database
// gave it, else the upload's own title.
func rowName(r playlistfile.Entry) string {
	name := cells.Display(r.Upload.Title)
	if r.Artist != "" && r.Title != "" {
		name = cells.Display(r.Artist + " - " + r.Title)
	}
	if name == "" {
		name = "(" + r.Upload.ID + ")"
	}
	return name
}

func kindLabel(r playlistfile.Entry) string {
	switch engine.PlanKind(r.Kind) {
	case engine.KindOfficial, engine.KindOwnOfficial:
		return "official"
	case engine.KindBetter:
		return "better"
	case engine.KindOriginal:
		return "original"
	case engine.KindNone:
		return "not found"
	}
	return r.Kind
}

// decisionGlyph is the mark at the start of a line.
func decisionGlyph(r playlistfile.Entry) string {
	switch {
	case r.Downloaded:
		return "↓"
	case r.Decision == playlistfile.DecisionConfirmed:
		return "✓"
	case r.Decision == playlistfile.DecisionReplaced:
		return "↻"
	case r.Decision == playlistfile.DecisionOriginal:
		return "="
	case r.Decision == playlistfile.DecisionSkipped:
		return "✗"
	}
	return "·"
}

// songRows are the rows of the list: the header, and the songs from the top of the
// window. The design owns the columns, their widths and the look of the selected row; what
// is here is the text of each cell and the colour it takes from the state of its song.
func (m Model) songRows(width int) []SongsRow {
	anyReplaced := false
	for _, r := range m.rows {
		anyReplaced = anyReplaced || r.Decision == playlistfile.DecisionReplaced
	}
	showAddress := width >= 90

	header := SongsRow{Review: "Review", Artist: "Artist", Song: "Song", Video: "Video", Style: RowDim}
	if anyReplaced {
		header.Replaced = "Replaced by"
		header.S6, header.S7 = "  ", ""
	}
	if showAddress {
		header.Address = "Address"
		header.S7 = "  "
	}
	dim := CellStyle{Fg: look.Colours.Dim}
	header.ReviewStyle, header.ArtistStyle, header.SongStyle, header.VideoStyle = dim, dim, dim, dim
	header.ReplacedStyle, header.AddressStyle = dim, dim
	rows := []SongsRow{header}

	for i := m.top; i < min(len(m.rows), m.top+m.listHeight()); i++ {
		r := m.rows[i]
		col := playlisteditor.ColumnsOf(r)
		kind := kindLabel(r)
		switch r.Decision {
		case playlistfile.DecisionReplaced:
			kind = "chosen"
		case playlistfile.DecisionOriginal:
			kind = "original"
		}
		mark := " "
		if i == m.sel {
			mark = "▸"
		}
		row := SongsRow{
			Mark: mark, Glyph: decisionGlyph(r), Review: kind,
			Artist: cells.Display(col.Artist), Song: cells.Display(col.Song), Video: cells.Display(col.Video),
		}
		if anyReplaced {
			row.S6, row.Replaced = "  ", cells.Display(col.Replaced)
		}
		if target, _, has := playlisteditor.Target(r); has && showAddress {
			row.S7, row.Address = "  ", "youtube.com/watch?v="+target
		}

		if i == m.sel {
			row.Style = RowSelected
		} else {
			dimmed := r.Downloaded || r.Decision == playlistfile.DecisionSkipped
			row.MarkStyle = CellStyle{Fg: look.Colours.Dim}
			row.GlyphStyle = CellStyle{Fg: m.glyphColour(r)}
			row.ReviewStyle = CellStyle{Fg: kindColour(r, dimmed)}
			row.ArtistStyle = CellStyle{Fg: look.Colours.Focus}
			row.SongStyle = CellStyle{Fg: look.Colours.Text}
			if dimmed {
				row.ArtistStyle = CellStyle{Fg: look.Colours.Dim}
				row.SongStyle = CellStyle{Fg: look.Colours.Dim}
			}
			row.VideoStyle = CellStyle{Fg: look.Colours.Dim}
			row.ReplacedStyle = CellStyle{Fg: look.Colours.Focus}
			row.AddressStyle = CellStyle{Fg: look.Colours.Dim}
		}
		rows = append(rows, row)
	}
	return rows
}

// glyphColour is the colour of the mark of a song: green once decided, cyan when it was
// changed, dim otherwise.
func (m Model) glyphColour(r playlistfile.Entry) string {
	switch {
	case r.Downloaded, r.Decision == playlistfile.DecisionConfirmed:
		return look.Colours.Success
	case r.Decision == playlistfile.DecisionReplaced || r.Decision == playlistfile.DecisionOriginal:
		return look.Colours.Focus
	}
	return look.Colours.Dim
}

// kindColour is the colour of what is proposed for a song.
func kindColour(r playlistfile.Entry, dimmed bool) string {
	if dimmed {
		return look.Colours.Dim
	}
	switch engine.PlanKind(r.Kind) {
	case engine.KindOfficial, engine.KindOwnOfficial:
		return look.Colours.Success
	case engine.KindBetter:
		return look.Colours.Highlight
	case engine.KindNone:
		return look.Colours.Error
	}
	return look.Colours.Text
}

// link makes an address clickable in terminals that support it, and plain text elsewhere.
func link(url string) string {
	return ansi.SetHyperlink(url) + url + ansi.ResetHyperlink()
}

func (m Model) detail(width int) []string {
	r, ok := m.current()
	if !ok {
		return make([]string, editorDetailLines)
	}
	lines := make([]string, 0, editorDetailLines)

	pl := cells.Display(r.Playlist)
	if pl == "" {
		pl = cells.Display(r.PlaylistURL)
	}
	lines = append(lines, cells.Fit(" "+look.Dim.Render("Playlist  ")+look.Text.Render(pl)+look.Dim.Render(fmt.Sprintf("  ·  %d of %d", m.sel+1, len(m.rows))), width))

	in := " " + look.Dim.Render("In list   ") + look.Text.Render(cells.Display(r.Upload.Title))
	if r.Upload.Channel != "" {
		in += look.Dim.Render("  ·  " + cells.Display(r.Upload.Channel))
	}
	if r.Upload.ID != "" {
		in += "  " + look.Dim.Render(link(playlisteditor.VideoAddress(r.Upload.ID)))
	}
	lines = append(lines, cells.Fit(in, width))

	prop := " " + look.Dim.Render("Proposed ") + " "
	if engine.PlanKind(r.Kind) == engine.KindNone {
		prop += look.Red.Render("nothing found")
	} else {
		prop += look.Cyan.Render(kindLabel(r))
		if r.TargetID != "" {
			prop += "  " + look.Dim.Render(link(playlisteditor.VideoAddress(r.TargetID)))
		}
	}
	if r.Reason != "" {
		prop += look.Dim.Render("  ·  " + r.Reason)
	}
	lines = append(lines, cells.Fit(prop, width))

	lines = append(lines, cells.Fit(" "+look.Dim.Render("Your pick ")+m.decisionText(r), width))
	return lines
}

func (m Model) decisionText(r playlistfile.Entry) string {
	switch {
	case r.Downloaded:
		return look.Green.Render("downloaded: it is skipped from now on")
	case r.Decision == playlistfile.DecisionConfirmed:
		return look.Green.Render("confirmed")
	case r.Decision == playlistfile.DecisionReplaced:
		text := "another video  " + look.Dim.Render(link(playlisteditor.VideoAddress(r.ChosenID)))
		if r.ChosenTitle != "" {
			text = cells.Display(r.ChosenTitle) + "  " + look.Dim.Render(link(playlisteditor.VideoAddress(r.ChosenID)))
		}
		return look.Cyan.Render("replaced by ") + text
	case r.Decision == playlistfile.DecisionOriginal:
		return look.Cyan.Render("the upload that is in the list")
	case r.Decision == playlistfile.DecisionSkipped:
		return look.Dim.Render("skipped")
	}
	return look.Yellow.Render("not reviewed yet")
}

func (m Model) hints() []keybar.Hint {
	if m.planning {
		return []keybar.Hint{{Key: "esc", Desc: "cancel"}}
	}
	return []keybar.Hint{
		{Key: "o", Desc: "open"}, {Key: "c", Desc: "confirm"}, {Key: "r", Desc: "replace"}, {Key: "u", Desc: "original"}, {Key: "x", Desc: "skip"},
		{Key: "a", Desc: "confirm all"}, {Key: "d", Desc: "download"}, {Key: "s", Desc: "save"}, {Key: "l", Desc: "open file"}, {Key: "esc", Desc: "back"},
	}
}

// bottomParts is what the bottom line says: a question or a message, or (keys true) the
// keys, which the design draws itself.
func (m Model) bottomParts(width int) (line string, keys bool) {
	switch m.mode {
	case AskPrevious:
		pending := 0
		for _, r := range m.saved {
			if !r.Downloaded {
				pending++
			}
		}
		return cells.Fit(" "+look.Yellow.Render(fmt.Sprintf("A review from last time is saved (%d songs, %d not downloaded). Bring it along? ", len(m.saved), pending))+
			look.Key.Render("y")+look.Dim.Render(" yes  ")+look.Key.Render("n")+look.Dim.Render(" no"), width), false
	case AskDownload:
		return cells.Fit(" "+look.Yellow.Render(fmt.Sprintf("%d songs were not reviewed. ", playlisteditor.Unreviewed(m.rows)))+
			look.Key.Render("y")+look.Dim.Render(" skip them and download  ")+
			look.Key.Render("a")+look.Dim.Render(" download them as proposed  ")+
			look.Key.Render("esc")+look.Dim.Render(" back"), width), false
	case Input:
		return cells.Fit(" "+m.input.View(), width), false
	}
	if m.status != "" && !m.statusUntil.IsZero() && time.Now().Before(m.statusUntil) {
		return cells.Fit(" "+look.Yellow.Render(m.status), width), false
	}
	return "", true
}

// mouse handles a click or the wheel: a click on a song selects it, a click on the key
// bar does what the key does.
func (m Model) mouse(msg tea.MouseMsg) (Model, tea.Cmd) {
	if m.planning || m.mode != Browsing {
		return m, nil
	}
	if dir := pointer.WheelDirection(msg); dir != 0 {
		m.move(dir * pointer.WheelStep)
		return m, nil
	}
	if !pointer.IsLeftClick(msg) {
		return m, nil
	}
	x, y := msg.Mouse().X, msg.Mouse().Y
	if y == m.height-1 {
		if h, found := keybar.At(m.hints(), x); found {
			if key, ok := keybar.KeyMsg(h.Key); ok {
				return m.handleKey(key)
			}
		}
		return m, nil
	}
	if row := y - 2; row >= 0 && row < m.listHeight() && m.top+row < len(m.rows) {
		m.sel = m.top + row
	}
	return m, nil
}

// ApplyDesigner does what a click on the designer's drawing does: a key bar entry presses
// its key and a song of the list is selected. x and y are the cell inside the component.
func (m Model) ApplyDesigner(name string, x, y int) (Model, tea.Cmd) {
	if m.planning || m.mode != Browsing {
		return m, nil
	}
	ev, ok := EventNamed(name, x, y)
	if !ok {
		return m, nil
	}
	switch ev.(type) {
	case HintEvent:
		if h, found := keybar.At(m.hints(), x); found {
			if key, clickable := keybar.KeyMsg(h.Key); clickable {
				return m.handleKey(key)
			}
		}
	case PickEvent:
		if y >= 0 && y < m.listHeight() && m.top+y < len(m.rows) {
			m.sel = m.top + y
		}
	}
	return m, nil
}

// Outline describes the editor in plain terms, for a screen reader.
func (m Model) Outline() outline.ScreenOutline {
	if m.planning {
		t := m.state.Tally()
		return outline.ScreenOutline{
			Title:  "MVD · Review before downloading",
			Prompt: fmt.Sprintf("Looking up the best version of each song: %d of %d", t.Planned+t.NotFound+t.Failed, t.Total),
			Keys:   outline.KeysOf(m.hints()),
		}
	}
	out := outline.ScreenOutline{Title: m.title(), ItemsLabel: "Songs", Keys: outline.KeysOf(m.hints())}
	h := m.listHeight()
	for i := m.top; i < min(len(m.rows), m.top+h); i++ {
		r := m.rows[i]
		value := kindLabel(r) + ", " + strings.TrimSpace(strings.ReplaceAll(stripANSI(m.decisionText(r)), "\n", " "))
		out.Items = append(out.Items, outline.Item{Label: rowName(r), Value: value, Selected: i == m.sel})
	}
	switch m.mode {
	case AskPrevious:
		out.Prompt = "A review from last time is saved. Bring it along? y or n"
	case AskDownload:
		out.Prompt = fmt.Sprintf("%d songs were not reviewed. y to skip them and download, a to download them as proposed, esc to go back", playlisteditor.Unreviewed(m.rows))
	case Input:
		out.Prompt = strings.TrimSpace(m.input.Prompt)
		out.Text, out.HasText, out.TextLabel = m.input.Value(), true, strings.TrimSuffix(strings.TrimSpace(m.input.Prompt), ":")
	}
	return out
}

func stripANSI(s string) string { return ansi.Strip(s) }
