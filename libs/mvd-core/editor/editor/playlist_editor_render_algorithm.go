package editor

import (
	"fmt"
	"strings"
	"time"
	"youtube-downloader/libs/mvd-core/cells"
	editorplanningpage "youtube-downloader/libs/mvd-core/editor/editorplanning"
	"youtube-downloader/libs/mvd-core/editor/review"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/outline"
	"youtube-downloader/libs/mvd-core/pointer"
	"youtube-downloader/libs/mvd-core/screenview"
	"youtube-downloader/libs/mvd-core/songlist"

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
	if m.session.Planning {
		return m.renderPlanning(width)
	}

	props := Default()
	props.Palette = Palette(look.CurrentPalette())
	props.Title = " " + m.session.Title()
	props.Songs = m.songRows(width)
	props.ShowEmpty = len(m.session.Rows) == 0

	detail := m.detail(width)
	props.Detail1, props.Detail2, props.Detail3, props.Detail4 = cells.CellText(detail[0]), cells.CellText(detail[1]), cells.CellText(detail[2]), cells.CellText(detail[3])

	line, keys := m.bottomParts(width)
	props.Hints = keybar.Of[HintsHint](m.hints())
	props.ShowKeys, props.ShowLine = keys, !keys
	props.Bottom = line
	return screenview.Padded(View(props, width, m.height).View, width)
}

func (m Model) renderPlanning(width int) string {
	done, total := m.session.PlanProgress()
	body := []string{
		" ",
		" " + look.Cyan.Render(look.SpinnerFrames[m.spin]) + " " + look.Text.Render("Looking up the best version of each song, nothing is downloaded."),
		" ",
	}
	if total == 0 {
		body = append(body, look.Dim.Render(" Reading the playlists…"))
	} else {
		body = append(body, look.Dim.Render(fmt.Sprintf(" %d of %d songs looked up", done, total)))
	}
	for _, pl := range m.session.PlanState.Playlists {
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
	return screenview.Padded(editorplanningpage.View(props, width, m.height).View, width)
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
	s := m.session
	anyReplaced := false
	for _, r := range s.Rows {
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

	for i := m.top; i < min(len(s.Rows), m.top+m.listHeight()); i++ {
		r := s.Rows[i]
		col := playlisteditor.ColumnsOf(r)
		kind := s.RowLabel(i)
		mark := " "
		if i == s.Selected {
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

		if i == s.Selected {
			row.Style = RowSelected
		} else {
			dimmed := s.Dimmed(i)
			row.MarkStyle = CellStyle{Fg: look.Colours.Dim}
			row.GlyphStyle = CellStyle{Fg: toneColour(review.DecisionTone(r))}
			row.ReviewStyle = CellStyle{Fg: toneColour(review.KindTone(r, dimmed))}
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

// toneColour is the colour of the current theme that a tone is drawn in.
func toneColour(t songlist.Tone) string {
	c := look.Colours
	switch t {
	case songlist.ToneDim:
		return c.Dim
	case songlist.ToneFocus:
		return c.Focus
	case songlist.ToneHighlight:
		return c.Highlight
	case songlist.ToneSuccess:
		return c.Success
	case songlist.ToneError:
		return c.Error
	}
	return c.Text
}

// link makes an address clickable in terminals that support it, and plain text elsewhere.
func link(url string) string {
	return ansi.SetHyperlink(url) + url + ansi.ResetHyperlink()
}

func (m Model) detail(width int) []string {
	r, ok := m.session.Current()
	if !ok {
		return make([]string, editorDetailLines)
	}
	lines := make([]string, 0, editorDetailLines)

	pl := cells.Display(r.Playlist)
	if pl == "" {
		pl = cells.Display(r.PlaylistURL)
	}
	lines = append(lines, cells.Fit(" "+look.Dim.Render("Playlist  ")+look.Text.Render(pl)+look.Dim.Render(fmt.Sprintf("  ·  %d of %d", m.session.Selected+1, len(m.session.Rows))), width))

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
		prop += look.Cyan.Render(review.KindLabel(r))
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

// decisionText is what the person's pick for a song says, in the colour of its tone.
func (m Model) decisionText(r playlistfile.Entry) string {
	d := review.DecisionOf(r)
	style := look.Yellow
	switch d.Tone {
	case songlist.ToneSuccess:
		style = look.Green
	case songlist.ToneFocus:
		style = look.Cyan
	case songlist.ToneDim:
		style = look.Dim
	}
	text := style.Render(d.Lead)
	if d.Video != "" {
		text += cells.Display(d.Title) + "  " + look.Dim.Render(link(playlisteditor.VideoAddress(d.Video)))
	}
	return text
}

func (m Model) hints() []keybar.Hint {
	if m.session.Planning {
		return []keybar.Hint{{Key: "esc", Desc: "cancel"}}
	}
	if m.session.IsLive() {
		return []keybar.Hint{
			{Key: "o", Desc: "open"}, {Key: "c", Desc: "download"}, {Key: "r", Desc: "swap video"}, {Key: "u", Desc: "original"}, {Key: "x", Desc: "cancel song"},
			{Key: "a", Desc: "download all"}, {Key: "esc", Desc: "progress"},
		}
	}
	return []keybar.Hint{
		{Key: "o", Desc: "open"}, {Key: "c", Desc: "confirm"}, {Key: "r", Desc: "replace"}, {Key: "u", Desc: "original"}, {Key: "x", Desc: "skip"},
		{Key: "a", Desc: "confirm all"}, {Key: "d", Desc: "download"}, {Key: "s", Desc: "save"}, {Key: "l", Desc: "open file"}, {Key: "esc", Desc: "back"},
	}
}

// bottomParts is what the bottom line says: a question or a message, or (keys true) the
// keys, which the design draws itself.
func (m Model) bottomParts(width int) (line string, keys bool) {
	switch m.session.Mode {
	case review.AskPrevious:
		return cells.Fit(" "+look.Yellow.Render(m.session.PreviousQuestion())+
			look.Key.Render("y")+look.Dim.Render(" yes  ")+look.Key.Render("n")+look.Dim.Render(" no"), width), false
	case review.AskDownload:
		return cells.Fit(" "+look.Yellow.Render(m.session.DownloadQuestion())+
			look.Key.Render("y")+look.Dim.Render(" skip them and download  ")+
			look.Key.Render("a")+look.Dim.Render(" download them as proposed  ")+
			look.Key.Render("esc")+look.Dim.Render(" back"), width), false
	case review.Input:
		return cells.Fit(" "+m.input.View(), width), false
	}
	if notice, ok := m.session.NoticeShown(time.Now()); ok {
		return cells.Fit(" "+look.Yellow.Render(notice), width), false
	}
	return "", true
}

// mouse handles a click or the wheel: a click on a song selects it, a click on the key
// bar does what the key does.
func (m Model) mouse(msg tea.MouseMsg) (Model, tea.Cmd) {
	if m.session.Planning || m.session.Mode != review.Browsing {
		return m, nil
	}
	if dir := pointer.WheelDirection(msg); dir != 0 {
		m.session.Move(dir * pointer.WheelStep)
		return m.after(nil)
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
	if row := y - 2; row >= 0 && row < m.listHeight() && m.top+row < len(m.session.Rows) {
		m.session.Select(m.top + row)
	}
	return m, nil
}

// Outline describes the editor in plain terms, for a screen reader.
func (m Model) Outline() outline.ScreenOutline {
	if m.session.Planning {
		done, total := m.session.PlanProgress()
		return outline.ScreenOutline{
			Title:  "MVD · Review before downloading",
			Prompt: fmt.Sprintf("Looking up the best version of each song: %d of %d", done, total),
			Keys:   outline.KeysOf(m.hints()),
		}
	}
	s := m.session
	out := outline.ScreenOutline{Title: s.Title(), ItemsLabel: "Songs", Keys: outline.KeysOf(m.hints())}
	h := m.listHeight()
	for i := m.top; i < min(len(s.Rows), m.top+h); i++ {
		r := s.Rows[i]
		value := s.RowLabel(i) + ", " + strings.TrimSpace(strings.ReplaceAll(stripANSI(m.decisionText(r)), "\n", " "))
		out.Items = append(out.Items, outline.Item{Label: rowName(r), Value: value, Selected: i == s.Selected})
	}
	switch s.Mode {
	case review.AskPrevious:
		out.Prompt = "A review from last time is saved. Bring it along? y or n"
	case review.AskDownload:
		out.Prompt = fmt.Sprintf("%d songs were not reviewed. y to skip them and download, a to download them as proposed, esc to go back", s.Unreviewed())
	case review.Input:
		out.Prompt = strings.TrimSpace(m.input.Prompt)
		out.Text, out.HasText, out.TextLabel = m.input.Value(), true, strings.TrimSuffix(strings.TrimSpace(m.input.Prompt), ":")
	}
	return out
}

func stripANSI(s string) string { return ansi.Strip(s) }
