package downloadsnapshot

import (
	"fmt"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/run/downloadsession"
	"youtube-downloader/libs/mvd-core/runstate"
	"youtube-downloader/libs/mvd-core/songlist"
)

// OutputTitle is what the pane of the details is called.
func OutputTitle(s downloadsession.Session) string {
	if !s.Technical || s.SelPlaylist >= len(s.State.Playlists) {
		return "Details"
	}
	en := s.SelectedEntry()
	if s.Pane == downloadsession.PaneLists || en == nil {
		return "Technical log · " + PlaylistName(s, s.State.Playlists[s.SelPlaylist])
	}

	return fmt.Sprintf("Technical log · %02d %s", en.Index, songlist.VideoTitle(en))
}

// FooterOf is what the pane keeps for the song chosen under its lines.
func FooterOf(s downloadsession.Session) Footer {
	en := s.SelectedEntry()
	if en == nil || s.Pane == downloadsession.PaneLists {
		return Footer{None: true}
	}
	switch en.State {
	case engine.StateDownloading, engine.StateMerging:
		info := runstate.ProgressLine(en)
		if en.State == engine.StateMerging {
			info = "merging streams with ffmpeg"
		}
		percent := en.Percent
		if en.State == engine.StateMerging {
			percent = 100 // the parts are all in: the files are being joined
		} else if en.Total == 0 {
			percent = 0 // the size is not known yet: an empty bar, not a full one
		}

		return Footer{Info: info, Percent: percent, HasBar: true, Tone: songlist.ToneHighlight}
	case engine.StateFailed:
		if s.Technical {
			return Footer{Info: "✗ " + runstate.PlainError(en.Err), Tone: songlist.ToneError}
		}
	case engine.StateDone:
		if !s.Technical {
			break // the details say it
		}
		what := "original " + en.TargetID
		if en.Official {
			what = "official video " + en.TargetID
		} else if en.Better {
			what = "better quality video " + en.TargetID
		}

		return Footer{Info: "✓ downloaded " + what, Tone: songlist.ToneSuccess}
	}

	return Footer{None: true}
}
