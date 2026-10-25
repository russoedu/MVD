package downloadsnapshot

import (
	"fmt"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/run/downloadsession"
	"youtube-downloader/libs/mvd-core/runstate"
	"youtube-downloader/libs/mvd-core/songlist"
)

// PlaylistRowsOf are the links of the list.
func PlaylistRowsOf(s downloadsession.Session) []PlaylistRow {
	rows := make([]PlaylistRow, 0, len(s.State.Playlists))
	for _, pl := range s.State.Playlists {
		finished, _, _ := s.State.PlaylistTally(pl)
		finished -= s.WaitingIn(pl) // a song that waits to be chosen is not done
		count := fmt.Sprintf("%d/%d", finished, len(pl.Entries))
		if pl.Err != "" {
			count = "can't read"
		} else if !pl.Listed {
			count = "reading"
		}
		rows = append(rows, PlaylistRow{Name: PlaylistName(s, pl), Count: count, Glyph: playlistGlyph(s, pl)})
	}

	return rows
}

// playlistGlyph is the mark at the end of a link.
func playlistGlyph(s downloadsession.Session, pl *runstate.Playlist) Glyph {
	if pl.Err != "" {
		return Glyph{GlyphFailed, songlist.ToneError}
	}
	if !pl.Listed {
		return Glyph{GlyphBusy, songlist.ToneDim}
	}
	finished, failed, active := s.State.PlaylistTally(pl)
	switch {
	case active > 0:
		return Glyph{GlyphBusy, songlist.ToneHighlight}
	case s.WaitingIn(pl) > 0:
		return Glyph{GlyphChoose, songlist.ToneHighlight}
	case finished == len(pl.Entries) && failed > 0:
		return Glyph{GlyphFailed, songlist.ToneError}
	case finished == len(pl.Entries):
		return Glyph{GlyphDone, songlist.ToneSuccess}
	}

	return Glyph{GlyphIdle, songlist.ToneDim}
}

// SongRowOf is a song of the table.
func SongRowOf(s downloadsession.Session, en *runstate.Entry) SongRow {
	status, tone := songlist.StatusWord(en)
	video := ""
	if songlist.Name(en) != songlist.VideoTitle(en) {
		video = songlist.VideoTitle(en)
	}

	return SongRow{
		ID: en.ID, Index: en.Index,
		Name: songlist.Name(en), Artist: songlist.Artist(en), Video: video, Album: en.NamedAlbum,
		Quality:  QualityText(en),
		Progress: ProgressOf(en),
		Status:   status, Tone: tone,
		Glyph:  SongGlyph(en),
		Ticked: s.Picked[en.ID],
	}
}

// ColumnHasContent says whether any song shown has something for the artist, video, album or quality
// column, which are left out of the table while none has.
func ColumnHasContent(s downloadsession.Session, column string) bool {
	for _, id := range s.VisibleEntries() {
		en := s.State.Entries[id]
		switch column {
		case downloadsession.ColumnArtist:
			if songlist.Artist(en) != "" {
				return true
			}
		case downloadsession.ColumnVideo:
			if songlist.Name(en) != songlist.VideoTitle(en) {
				return true
			}
		case downloadsession.ColumnAlbum:
			if en.NamedAlbum != "" {
				return true
			}
		case downloadsession.ColumnQuality:
			if QualityText(en) != "" {
				return true
			}
		}
	}

	return false
}

// QualityText is how good the version of a song is: the height of its picture ("1080p"), or for one
// with sound only the bitrate ("160k"); empty until yt-dlp has said.
func QualityText(en *runstate.Entry) string {
	switch {
	case en.Height > 0:
		return fmt.Sprintf("%dp", en.Height)
	case en.AudioKbps > 0:
		return fmt.Sprintf("%dk", en.AudioKbps)
	}

	return ""
}

// ProgressOf is how far a song is: a bar while it comes in, a full one once it is saved, and nothing
// for a song that has not begun or has failed.
func ProgressOf(en *runstate.Entry) Progress {
	switch en.State {
	case engine.StateDownloading:
		if en.Total > 0 {
			return Progress{Visible: true, Percent: en.Percent, Tone: songlist.ToneHighlight}
		}
	case engine.StateMerging:
		return Progress{Visible: true, Percent: 100, Tone: songlist.ToneHighlight}
	case engine.StateDone:
		return Progress{Visible: true, Percent: 100, Tone: songlist.ToneSuccess}
	}

	return Progress{Tone: songlist.ToneText}
}

// SongGlyph is the mark in front of a song.
func SongGlyph(en *runstate.Entry) Glyph {
	switch en.State {
	case engine.StateQueued:
		return Glyph{GlyphIdle, songlist.ToneDim}
	case engine.StateResolving:
		return Glyph{GlyphResolving, songlist.ToneFocus}
	case engine.StateDownloading:
		return Glyph{GlyphBusy, songlist.ToneHighlight}
	case engine.StateMerging:
		return Glyph{GlyphMerging, songlist.ToneHighlight}
	case engine.StateDone:
		return Glyph{GlyphDone, songlist.ToneSuccess}
	case engine.StateDuplicate:
		return Glyph{GlyphHad, songlist.ToneDim}
	case engine.StateNotFound:
		return Glyph{GlyphMissing, songlist.ToneDim}
	case engine.StateFailed:
		return Glyph{GlyphFailed, songlist.ToneError}
	case engine.StatePlanned:
		if songlist.FoundOfficial(en) {
			return Glyph{GlyphDone, songlist.ToneSuccess}
		}

		return Glyph{GlyphChoose, songlist.ToneHighlight}
	case engine.StateAccepted:
		return Glyph{GlyphAccepted, songlist.ToneFocus}
	case engine.StateSkipped:
		return Glyph{GlyphRejected, songlist.ToneDim}
	}

	return Glyph{GlyphNone, songlist.ToneText}
}
