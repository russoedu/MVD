package download

import (
	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/run/downloadsnapshot"
	"youtube-downloader/libs/mvd-core/runstate"
)

// detailRows are the lines of the details pane, each wrapped to w cells: what the snapshot says of the
// song that is chosen, or of the link, drawn in the tones' colours.
func (m Model) detailRows(w int) []OutputRow {
	var rows []OutputRow
	for _, l := range downloadsnapshot.DetailsOf(m.session) {
		if l.Blank {
			rows = append(rows, OutputRow{Line: " "})
			continue
		}
		style := CellStyle{Fg: toneColour(l.Tone), Bold: l.Strong}
		text := l.Text
		if l.SongTitle {
			text = cells.Display(text)
		}
		for _, line := range wrapText(text, max(w-2, 8)) {
			rows = append(rows, OutputRow{Line: cells.CellText(line), LineStyle: style})
		}
	}

	return rows
}

// songSentence says what happened to a song, and what to do when it went wrong, with its colour.
func (m Model) songSentence(en *runstate.Entry) (string, CellStyle) {
	text, tone := downloadsnapshot.SongSentence(m.session, en)

	return text, CellStyle{Fg: toneColour(tone)}
}

// playlistName is what a link is called, as it can be drawn in the cells.
func (m Model) playlistName(pl *runstate.Playlist) string {
	return cells.Display(downloadsnapshot.PlaylistName(m.session, pl))
}
