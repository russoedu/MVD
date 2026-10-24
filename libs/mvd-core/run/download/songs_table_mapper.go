package download

import (
	"fmt"
	"strings"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/internal/cuppa"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/runstate"
)

// The columns of the table of songs, and how wide each is, in cells. The song takes what is left.
const (
	columnMark     = "mark"
	columnIdx      = "idx"
	columnSong     = "song"
	columnArtist   = "artist"
	columnVideo    = "video"
	columnProgress = "progress"
	columnStatus   = "status"

	widthMark     = 1
	widthIdx      = 2
	widthStatus   = 11
	widthProgress = 13 // a bar of 8, a space and "100%"
	widthArtist   = 16
	widthVideo    = 22
	minSongWidth  = 12
	progressBar   = 8
)

// tableColumn is one column of the table: what it is, what its heading says and how wide it is.
type tableColumn struct {
	name, title string
	width       int
}

// tableColumns are the columns of the table of songs in a pane w cells wide, from left to right.
// The mark, the number and the song are always there; the rest are added by importance while
// they fit, and the artist and the video name only when some song has something to say there.
func (m Model) tableColumns(w int) []tableColumn {
	widths := map[string]int{
		columnStatus: widthStatus, columnProgress: widthProgress, columnArtist: widthArtist, columnVideo: widthVideo,
	}
	used := widthMark + widthIdx + minSongWidth + 2 // and the gaps between these three
	include := map[string]bool{}
	for _, name := range []string{columnStatus, columnProgress, columnArtist, columnVideo} {
		if (name == columnArtist || name == columnVideo) && !m.tableHas(name) {
			continue
		}
		if used+widths[name]+1 <= w {
			include[name] = true
			used += widths[name] + 1
		}
	}

	all := []tableColumn{
		{columnMark, "", widthMark}, {columnIdx, "#", widthIdx}, {columnSong, "Song", 0},
		{columnArtist, "Artist", widthArtist}, {columnVideo, "Video", widthVideo},
		{columnProgress, "Progress", widthProgress}, {columnStatus, "Status", widthStatus},
	}
	var out []tableColumn
	for _, c := range all {
		if c.name == columnMark || c.name == columnIdx || c.name == columnSong || include[c.name] {
			out = append(out, c)
		}
	}
	taken := len(out) - 1 // the gaps
	for _, c := range out {
		taken += c.width
	}
	for i := range out {
		if out[i].name == columnSong {
			out[i].width = max(w-taken, minSongWidth)
		}
	}

	return out
}

// tableHas says whether any song shown has something for the artist or the video column.
func (m Model) tableHas(column string) bool {
	for _, id := range m.visibleEntries() {
		en := m.state.Entries[id]
		switch column {
		case columnArtist:
			if m.whoSings(en) != "" {
				return true
			}
		case columnVideo:
			if songName(en) != entryTitle(en) {
				return true
			}
		}
	}

	return false
}

// songName is the song as it is called: as a music database named it, else the name of its video.
func songName(en *runstate.Entry) string {
	if en.NamedTitle != "" {
		return en.NamedTitle
	}

	return entryTitle(en)
}

// columnsSpec is the columns of the rows widget for these table columns, with a cell of one space
// between two, so a selected row is coloured all the way across.
func columnsSpec(cols []tableColumn) string {
	const selectable = ":fg=@Text:selbg=@Selected:selfg=@Text:selbold"
	var parts []string
	for i, c := range cols {
		if i > 0 {
			parts = append(parts, "Sp:1"+selectable)
		}
		switch c.name {
		case columnMark:
			parts = append(parts, "Mark:1")
		default:
			parts = append(parts, fmt.Sprintf("%s:%d:ellipsis%s", strings.ToUpper(c.name[:1])+c.name[1:], c.width, selectable))
		}
	}

	return strings.Join(parts, ",")
}

// songsTable is the layout of the songs in a pane w cells wide, with the heading of each column.
// It is empty while there are no songs to show, and the pane then keeps its rows of text.
func (m Model) songsTable(w int) Table {
	if len(m.visibleEntries()) == 0 {
		return Table{}
	}
	cols := m.tableColumns(w)
	dim := cuppa.CellStyleCode(CellStyle{Fg: look.Colours.Dim})
	var header, styles []string
	for i, c := range cols {
		if i > 0 {
			header, styles = append(header, " "), append(styles, "")
		}
		header, styles = append(header, c.title), append(styles, dim)
	}

	return Table{Columns: columnsSpec(cols), Header: header, HeaderStyles: styles}
}

// songCells are the cells of the row of a song in the table, and their styles as codes.
func (m Model) songCells(en *runstate.Entry, cols []tableColumn, selected bool) (texts, styles []string) {
	glyph, glyphColour := m.entryGlyph(en)
	status, statusColour := entryStatus(en)
	progress, progressColour := progressCell(en)
	plain := func(colour string) string { return cuppa.CellStyleCode(CellStyle{Fg: colour}) }

	for i, c := range cols {
		if i > 0 {
			texts, styles = append(texts, " "), append(styles, "")
		}
		var text, style string
		switch c.name {
		case columnMark:
			text, style = glyph, plain(glyphColour)
		case columnIdx:
			text = fmt.Sprintf("%02d", en.Index)
		case columnSong:
			text = songName(en)
			if selected {
				text = marquee(cells.CellText(text), c.width, m.marquee)
			}
		case columnArtist:
			text, style = m.whoSings(en), plain(look.Colours.Dim)
		case columnVideo:
			if songName(en) != entryTitle(en) {
				text = entryTitle(en)
			}
			style = plain(look.Colours.Dim)
		case columnProgress:
			text, style = progress, plain(progressColour)
		case columnStatus:
			text, style = status, plain(statusColour)
		}
		if c.name != columnSong || !selected {
			text = cells.CellText(text)
		}
		texts, styles = append(texts, text), append(styles, style)
	}

	return texts, styles
}

// entryStatus is the word that says where a song is, and its colour.
func entryStatus(en *runstate.Entry) (word, colour string) {
	c := look.Colours
	switch en.State {
	case engine.StateQueued:
		return "waiting", c.Dim
	case engine.StateResolving:
		return "finding", c.Focus
	case engine.StateDownloading:
		if en.Total > 0 {
			return "downloading", c.Highlight
		}
		return "starting", c.Highlight
	case engine.StateMerging:
		return "joining", c.Highlight
	case engine.StateDuplicate:
		return "have it", c.Dim
	case engine.StateNotFound:
		return "no official", c.Dim
	case engine.StateFailed:
		return "failed", c.Error
	case engine.StatePlanned:
		return "chosen", c.Dim
	case engine.StateDone:
		switch {
		case en.Official:
			return "official", c.Focus
		case en.Better:
			return "better", c.Focus
		}
		return "saved", c.Success
	}

	return "", c.Text
}

// progressCell is a bar and a percentage while a song comes in, a full bar once it is saved, and
// nothing for a song that has not begun or has failed.
func progressCell(en *runstate.Entry) (text, colour string) {
	c := look.Colours
	switch en.State {
	case engine.StateDownloading:
		if en.Total > 0 {
			return bar(en.Percent) + fmt.Sprintf(" %3.0f%%", en.Percent), c.Highlight
		}
	case engine.StateMerging:
		return bar(100) + " 100%", c.Highlight
	case engine.StateDone:
		return bar(100) + " 100%", c.Success
	}

	return "", c.Text
}

// bar draws percent as a bar of progressBar cells.
func bar(percent float64) string {
	filled := min(max(int(percent/100*progressBar+0.5), 0), progressBar)

	return strings.Repeat("█", filled) + strings.Repeat("░", progressBar-filled)
}
