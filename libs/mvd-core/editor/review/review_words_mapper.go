package review

import (
	"fmt"
	"strings"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlisteditor"
	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/songlist"
)

// Title is the title of the review: how many songs there are and how they stand.
func (s Session) Title() string {
	if s.live != nil {
		return s.liveTitle()
	}
	var downloaded, revised int
	for _, r := range s.Rows {
		switch {
		case r.Downloaded:
			downloaded++
		case playlisteditor.Revised(r) && r.Decision != playlistfile.DecisionSkipped:
			revised++
		}
	}
	parts := []string{fmt.Sprintf("%d songs", len(s.Rows))}
	if revised > 0 {
		parts = append(parts, fmt.Sprintf("%d to download", revised))
	}
	if n := playlisteditor.Unreviewed(s.Rows); n > 0 {
		parts = append(parts, fmt.Sprintf("%d not reviewed", n))
	}
	if downloaded > 0 {
		parts = append(parts, fmt.Sprintf("%d downloaded", downloaded))
	}

	return "MVD · Review before downloading  ·  " + strings.Join(parts, " · ")
}

// liveTitle is the title of a review that goes on while the songs are looked up: how many there
// are, how many are still being looked up, and how many have gone to download.
func (s Session) liveTitle() string {
	var waiting, sent, done int
	for i := range s.Rows {
		switch {
		case s.Waiting(i):
			waiting++
		case s.Rows[i].Downloaded:
			done++
		case s.Sent(i):
			sent++
		}
	}
	parts := []string{fmt.Sprintf("%d songs", len(s.Rows))}
	if waiting > 0 {
		parts = append(parts, fmt.Sprintf("%d being looked up", waiting))
	}
	if n := playlisteditor.Unreviewed(s.Rows) - waiting; n > 0 {
		parts = append(parts, fmt.Sprintf("%d to look at", n))
	}
	if sent > 0 {
		parts = append(parts, fmt.Sprintf("%d downloading", sent))
	}
	if done > 0 {
		parts = append(parts, fmt.Sprintf("%d downloaded", done))
	}

	return "MVD · Choose what to download  ·  " + strings.Join(parts, " · ")
}

// Waiting is true for a song of a live review that has not been looked up yet.
func (s Session) Waiting(i int) bool {
	if s.live == nil || i >= len(s.States) {
		return false
	}

	return (s.States[i] == engine.StateQueued || s.States[i] == engine.StateResolving) && s.Rows[i].TargetID == "" && s.Rows[i].Decision == playlistfile.DecisionNone
}

// Sent is true for a song of a live review that has gone to download and is not done.
func (s Session) Sent(i int) bool {
	if s.live == nil || i >= len(s.States) {
		return false
	}
	switch s.States[i] {
	case engine.StateQueued, engine.StateDownloading, engine.StateMerging, engine.StateResolving:
		return !s.Waiting(i)
	}

	return false
}

// Dimmed is true for a song that needs no more of the person's attention.
func (s Session) Dimmed(i int) bool {
	r := s.Rows[i]

	return r.Downloaded || r.Decision == playlistfile.DecisionSkipped || s.Waiting(i) || s.Sent(i)
}

// RowLabel is what the review column says of song i: what is proposed for it, or in a live review
// where the song has got to.
func (s Session) RowLabel(i int) string {
	r := s.Rows[i]
	if s.live != nil {
		switch {
		case s.Waiting(i):
			return "looking up…"
		case r.Downloaded:
			return "downloaded"
		case s.Sent(i) && i < len(s.States) && (s.States[i] == engine.StateDownloading || s.States[i] == engine.StateMerging):
			return "downloading"
		case s.Sent(i):
			return "queued"
		case i < len(s.States) && s.States[i] == engine.StateFailed:
			return "failed"
		}
	}
	switch r.Decision {
	case playlistfile.DecisionReplaced:
		return "chosen"
	case playlistfile.DecisionOriginal:
		return "original"
	}

	return KindLabel(r)
}

// KindLabel is what is proposed for a song, in a word.
func KindLabel(r playlistfile.Entry) string {
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

// KindTone is the tone of what is proposed for a song; a song that is dimmed has none to speak of.
func KindTone(r playlistfile.Entry, dimmed bool) songlist.Tone {
	if dimmed {
		return songlist.ToneDim
	}
	switch engine.PlanKind(r.Kind) {
	case engine.KindOfficial, engine.KindOwnOfficial:
		return songlist.ToneSuccess
	case engine.KindBetter:
		return songlist.ToneHighlight
	case engine.KindNone:
		return songlist.ToneError
	}

	return songlist.ToneText
}

// DecisionTone is the tone of the mark of a song: success once decided, focus when it was changed,
// dim otherwise.
func DecisionTone(r playlistfile.Entry) songlist.Tone {
	switch {
	case r.Downloaded, r.Decision == playlistfile.DecisionConfirmed:
		return songlist.ToneSuccess
	case r.Decision == playlistfile.DecisionReplaced || r.Decision == playlistfile.DecisionOriginal:
		return songlist.ToneFocus
	}

	return songlist.ToneDim
}

// DecisionWords is what the person's pick for a song says: a lead in a tone, and for a video they
// picked its id and its title ("another video" while the title is not known).
type DecisionWords struct {
	Lead         string
	Tone         songlist.Tone
	Video, Title string
}

// DecisionOf says what the person's pick for a song is.
func DecisionOf(r playlistfile.Entry) DecisionWords {
	switch {
	case r.Downloaded:
		return DecisionWords{Lead: "downloaded: it is skipped from now on", Tone: songlist.ToneSuccess}
	case r.Decision == playlistfile.DecisionConfirmed:
		return DecisionWords{Lead: "confirmed", Tone: songlist.ToneSuccess}
	case r.Decision == playlistfile.DecisionReplaced:
		title := r.ChosenTitle
		if title == "" {
			title = "another video"
		}

		return DecisionWords{Lead: "replaced by ", Tone: songlist.ToneFocus, Video: r.ChosenID, Title: title}
	case r.Decision == playlistfile.DecisionOriginal:
		return DecisionWords{Lead: "the upload that is in the list", Tone: songlist.ToneFocus}
	case r.Decision == playlistfile.DecisionSkipped:
		return DecisionWords{Lead: "skipped", Tone: songlist.ToneDim}
	}

	return DecisionWords{Lead: "not reviewed yet", Tone: songlist.ToneHighlight}
}

// PreviousQuestion is the question about the review of last time.
func (s Session) PreviousQuestion() string {
	return fmt.Sprintf("A review from last time is saved (%d songs, %d not downloaded). Bring it along? ", len(s.Saved), s.PendingCount())
}

// DownloadQuestion is what is said of the songs nobody has looked at, before asking what to do
// with them.
func (s Session) DownloadQuestion() string {
	return fmt.Sprintf("%d songs were not reviewed. ", playlisteditor.Unreviewed(s.Rows))
}
