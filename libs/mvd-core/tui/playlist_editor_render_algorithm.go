package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlisteditor"
	"youtube-downloader/libs/mvd-core/playlistfile"
)

// editorDetailLines is how many lines the panel under the list takes.
const editorDetailLines = 4

// listHeight is how many songs fit on the screen: the title, a rule, the panel and the
// key bar take the rest.
func (m editorModel) listHeight() int {
	return max(3, m.height-(editorDetailLines+3))
}

func (m editorModel) render() string {
	width := max(m.width, 20)
	if m.planning {
		return m.renderPlanning(width)
	}

	var lines []string
	lines = append(lines, styTitle.Render(" "+m.title()))

	h := m.listHeight()
	for i := m.top; i < m.top+h; i++ {
		if i >= len(m.rows) {
			lines = append(lines, "")
			continue
		}
		lines = append(lines, m.rowLine(i, width))
	}
	if len(m.rows) == 0 {
		lines[1] = styDim.Render(" There are no songs here. Open a file with l, or leave and give a playlist.")
	}

	lines = append(lines, styBorder.Render(strings.Repeat("─", width)))
	lines = append(lines, m.detail(width)...)
	lines = append(lines, m.bottomLine(width))
	return strings.Join(lines, "\n")
}

func (m editorModel) renderPlanning(width int) string {
	t := m.state.Tally()
	done := t.Planned + t.NotFound + t.Failed
	body := []string{
		"",
		" " + styCyan.Render(spinnerFrames[m.spin]) + " " + styText.Render("Looking up the best version of each song, nothing is downloaded."),
		"",
	}
	if t.Total == 0 {
		body = append(body, styDim.Render(" Reading the playlists…"))
	} else {
		body = append(body, styDim.Render(fmt.Sprintf(" %d of %d songs looked up", done, t.Total)))
	}
	for _, pl := range m.state.Playlists {
		if pl.Err != "" {
			body = append(body, " "+styRed.Render(display(pl.Title)+": "+pl.Err))
		}
	}

	var lines []string
	lines = append(lines, styTitle.Render(" MVD · Review before downloading"))
	for i := 0; i < max(3, m.height-2); i++ {
		if i < len(body) {
			lines = append(lines, fit(body[i], width))
		} else {
			lines = append(lines, "")
		}
	}
	lines = append(lines, m.bottomLine(width))
	return strings.Join(lines, "\n")
}

func (m editorModel) title() string {
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
	name := display(r.Upload.Title)
	if r.Artist != "" && r.Title != "" {
		name = display(r.Artist + " - " + r.Title)
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

func (m editorModel) rowLine(i, width int) string {
	r := m.rows[i]
	cursor := " "
	if i == m.sel {
		cursor = "▸"
	}
	target, _, has := playlisteditor.Target(r)
	short := ""
	if has && width >= 90 {
		short = "youtube.com/watch?v=" + target
	}
	kind := kindLabel(r)
	if r.Decision == playlistfile.DecisionReplaced {
		kind = "chosen"
	}
	if r.Decision == playlistfile.DecisionOriginal {
		kind = "original"
	}

	nameWidth := width - 2 - 2 - 11 - len(short) - 2
	name := truncate(rowName(r), max(8, nameWidth))
	if r.Decision == playlistfile.DecisionReplaced && r.ChosenTitle != "" {
		name = truncate(rowName(r)+"  →  "+display(r.ChosenTitle), max(8, nameWidth))
	}

	plain := fmt.Sprintf("%s %s %-9s %s", cursor, decisionGlyph(r), kind, padRight(name, max(8, nameWidth)))
	if short != "" {
		plain += "  " + short
	}
	if i == m.sel {
		return stySelected.Render(fit(plain, width))
	}

	glyph := decisionGlyph(r)
	switch {
	case r.Downloaded:
		glyph = styGreen.Render(glyph)
	case r.Decision == playlistfile.DecisionConfirmed:
		glyph = styGreen.Render(glyph)
	case r.Decision == playlistfile.DecisionReplaced || r.Decision == playlistfile.DecisionOriginal:
		glyph = styCyan.Render(glyph)
	default:
		glyph = styDim.Render(glyph)
	}
	kindStyle := styText
	switch engine.PlanKind(r.Kind) {
	case engine.KindOfficial, engine.KindOwnOfficial:
		kindStyle = styGreen
	case engine.KindBetter:
		kindStyle = styYellow
	case engine.KindNone:
		kindStyle = styRed
	}
	nameStyle := styText
	if r.Downloaded || r.Decision == playlistfile.DecisionSkipped {
		nameStyle, kindStyle = styDim, styDim
	}
	line := " " + styDim.Render(cursor) + " " + glyph + " " + kindStyle.Render(fmt.Sprintf("%-9s", kind)) + " " + nameStyle.Render(padRight(name, max(8, nameWidth)))
	if short != "" {
		line += "  " + styDim.Render(short)
	}
	return fit(line, width)
}

// link makes an address clickable in terminals that support it, and plain text elsewhere.
func link(url string) string {
	return ansi.SetHyperlink(url) + url + ansi.ResetHyperlink()
}

func (m editorModel) detail(width int) []string {
	r, ok := m.current()
	if !ok {
		return make([]string, editorDetailLines)
	}
	lines := make([]string, 0, editorDetailLines)

	pl := display(r.Playlist)
	if pl == "" {
		pl = display(r.PlaylistURL)
	}
	lines = append(lines, fit(" "+styDim.Render("Playlist  ")+styText.Render(pl)+styDim.Render(fmt.Sprintf("  ·  %d of %d", m.sel+1, len(m.rows))), width))

	in := " " + styDim.Render("In list   ") + styText.Render(display(r.Upload.Title))
	if r.Upload.Channel != "" {
		in += styDim.Render("  ·  " + display(r.Upload.Channel))
	}
	if r.Upload.ID != "" {
		in += "  " + styDim.Render(link(playlisteditor.VideoAddress(r.Upload.ID)))
	}
	lines = append(lines, fit(in, width))

	prop := " " + styDim.Render("Proposed ") + " "
	if engine.PlanKind(r.Kind) == engine.KindNone {
		prop += styRed.Render("nothing found")
	} else {
		prop += styCyan.Render(kindLabel(r))
		if r.TargetID != "" {
			prop += "  " + styDim.Render(link(playlisteditor.VideoAddress(r.TargetID)))
		}
	}
	if r.Reason != "" {
		prop += styDim.Render("  ·  " + r.Reason)
	}
	lines = append(lines, fit(prop, width))

	lines = append(lines, fit(" "+styDim.Render("Your pick ")+" "+m.decisionText(r), width))
	return lines
}

func (m editorModel) decisionText(r playlistfile.Entry) string {
	switch {
	case r.Downloaded:
		return styGreen.Render("downloaded: it is skipped from now on")
	case r.Decision == playlistfile.DecisionConfirmed:
		return styGreen.Render("confirmed")
	case r.Decision == playlistfile.DecisionReplaced:
		text := "another video  " + styDim.Render(link(playlisteditor.VideoAddress(r.ChosenID)))
		if r.ChosenTitle != "" {
			text = display(r.ChosenTitle) + "  " + styDim.Render(link(playlisteditor.VideoAddress(r.ChosenID)))
		}
		return styCyan.Render("replaced by ") + text
	case r.Decision == playlistfile.DecisionOriginal:
		return styCyan.Render("the upload that is in the list")
	case r.Decision == playlistfile.DecisionSkipped:
		return styDim.Render("skipped")
	}
	return styYellow.Render("not reviewed yet")
}

func (m editorModel) hints() []keyHint {
	if m.planning {
		return []keyHint{{"esc", "cancel"}}
	}
	return []keyHint{
		{"o", "open"}, {"c", "confirm"}, {"r", "replace"}, {"u", "original"}, {"x", "skip"},
		{"a", "confirm all"}, {"d", "download"}, {"s", "save"}, {"l", "open file"}, {"esc", "back"},
	}
}

func (m editorModel) bottomLine(width int) string {
	switch m.mode {
	case editorAskPrevious:
		pending := 0
		for _, r := range m.saved {
			if !r.Downloaded {
				pending++
			}
		}
		return fit(" "+styYellow.Render(fmt.Sprintf("A review from last time is saved (%d songs, %d not downloaded). Bring it along? ", len(m.saved), pending))+
			styKey.Render("y")+styDim.Render(" yes  ")+styKey.Render("n")+styDim.Render(" no"), width)
	case editorAskDownload:
		return fit(" "+styYellow.Render(fmt.Sprintf("%d songs were not reviewed. ", playlisteditor.Unreviewed(m.rows)))+
			styKey.Render("y")+styDim.Render(" skip them and download  ")+
			styKey.Render("a")+styDim.Render(" download them as proposed  ")+
			styKey.Render("esc")+styDim.Render(" back"), width)
	case editorInput:
		return fit(" "+m.input.View(), width)
	}
	if m.status != "" && !m.statusUntil.IsZero() && time.Now().Before(m.statusUntil) {
		return fit(" "+styYellow.Render(m.status), width)
	}
	return keyBar(width, m.hints())
}

// mouse handles a click or the wheel: a click on a song selects it, a click on the key
// bar does what the key does.
func (m editorModel) mouse(msg tea.MouseMsg) (editorModel, tea.Cmd) {
	if m.planning || m.mode != editorBrowsing {
		return m, nil
	}
	if dir := wheelDirection(msg); dir != 0 {
		m.move(dir * wheelStep)
		return m, nil
	}
	if !isLeftClick(msg) {
		return m, nil
	}
	x, y := msg.Mouse().X, msg.Mouse().Y
	if y == m.height-1 {
		if h, found := hintAt(m.hints(), x); found {
			if key, ok := hintKeyMsg(h.key); ok {
				return m.handleKey(key)
			}
		}
		return m, nil
	}
	if row := y - 1; row >= 0 && row < m.listHeight() && m.top+row < len(m.rows) {
		m.sel = m.top + row
	}
	return m, nil
}

// Outline describes the editor in plain terms, for a screen reader.
func (m editorModel) Outline() ScreenOutline {
	if m.planning {
		t := m.state.Tally()
		return ScreenOutline{
			Title:  "MVD · Review before downloading",
			Prompt: fmt.Sprintf("Looking up the best version of each song: %d of %d", t.Planned+t.NotFound+t.Failed, t.Total),
			Keys:   outlineKeys(m.hints()),
		}
	}
	out := ScreenOutline{Title: m.title(), ItemsLabel: "Songs", Keys: outlineKeys(m.hints())}
	h := m.listHeight()
	for i := m.top; i < min(len(m.rows), m.top+h); i++ {
		r := m.rows[i]
		value := kindLabel(r) + ", " + strings.TrimSpace(strings.ReplaceAll(stripANSI(m.decisionText(r)), "\n", " "))
		out.Items = append(out.Items, OutlineItem{Label: rowName(r), Value: value, Selected: i == m.sel})
	}
	switch m.mode {
	case editorAskPrevious:
		out.Prompt = "A review from last time is saved. Bring it along? y or n"
	case editorAskDownload:
		out.Prompt = fmt.Sprintf("%d songs were not reviewed. y to skip them and download, a to download them as proposed, esc to go back", playlisteditor.Unreviewed(m.rows))
	case editorInput:
		out.Prompt = strings.TrimSpace(m.input.Prompt)
		out.Text, out.HasText, out.TextLabel = m.input.Value(), true, strings.TrimSuffix(strings.TrimSpace(m.input.Prompt), ":")
	}
	return out
}

func stripANSI(s string) string { return ansi.Strip(s) }
