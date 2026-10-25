package review

import (
	"strings"

	"youtube-downloader/libs/mvd-core/playlisteditor"
	"youtube-downloader/libs/mvd-core/playlistfile"
)

// DownloadAnswer is what the person said to the question about the songs nobody has looked at.
type DownloadAnswer int

const (
	// DownloadReviewed downloads only the songs that were decided on.
	DownloadReviewed DownloadAnswer = iota
	// DownloadAsProposed downloads the songs nobody looked at too, as proposed.
	DownloadAsProposed
	// DownloadBack goes back to the list.
	DownloadBack
)

// ask opens the one-line question.
func (s *Session) ask(purpose inputPurpose, prompt, value string) {
	s.Prompt, s.InputText = prompt, value
	s.purpose, s.Mode = purpose, Input
	s.Asks++
}

// AskReplace asks for the address of the video to download instead.
func (s *Session) AskReplace() {
	s.ask(inputReplace, "Address of the video to download instead: ", "")
}

// SetInput is what has been typed in the one-line question that is open.
func (s *Session) SetInput(text string) { s.InputText = text }

// CancelInput closes the one-line question.
func (s *Session) CancelInput() {
	if s.purpose == inputNote {
		s.logDecision("")
	}
	s.Mode = Browsing
}

// SubmitInput answers the one-line question with what has been typed.
func (s *Session) SubmitInput() []Effect {
	text := strings.TrimSpace(s.InputText)
	s.Mode = Browsing
	switch s.purpose {
	case inputReplace:
		id, ok := playlisteditor.VideoID(text)
		if !ok {
			s.Flash("that is not the address of a YouTube video")

			return nil
		}

		return s.Decide(playlistfile.DecisionReplaced, id, "")

	case inputNote:
		s.logDecision(text)

	case inputSave:
		if text == "" {
			return nil
		}
		s.saveFile(text)

	case inputLoad:
		if text == "" {
			return nil
		}
		s.loadFile(text)
	}

	return nil
}

// logDecision writes the decision made on the row being asked about, with the reason.
func (s Session) logDecision(note string) {
	if !playlisteditor.DebugBuild || s.opts.DecisionLog == "" {
		return
	}
	for _, row := range s.Rows {
		if row.Key() == s.noteFor {
			_ = playlisteditor.AppendDecision(s.opts.DecisionLog, playlisteditor.RecordFor(row, note))

			return
		}
	}
}

// Download asks for what to do with the songs nobody has looked at, or starts at once. A live
// review downloads as it goes, so it does nothing here.
func (s *Session) Download() []Effect {
	if s.live != nil {
		return nil
	}
	if !s.opts.CanDownload {
		s.Flash("downloading is not available here")

		return nil
	}
	if playlisteditor.Unreviewed(s.Rows) > 0 {
		s.Mode = AskDownload

		return nil
	}

	return s.startDownload(false)
}

// AnswerDownload answers the question about the songs nobody has looked at.
func (s *Session) AnswerDownload(answer DownloadAnswer) []Effect {
	s.Mode = Browsing
	switch answer {
	case DownloadReviewed:
		return s.startDownload(false)
	case DownloadAsProposed:
		return s.startDownload(true)
	}

	return nil
}

func (s *Session) startDownload(includeUnreviewed bool) []Effect {
	planned, keys := playlisteditor.Selection(s.Rows, includeUnreviewed)
	if len(planned) == 0 {
		s.Flash("nothing to download: confirm, replace or take the original of some songs first")

		return nil
	}

	return []Effect{Download{Plan: planned, Keys: keys}}
}
