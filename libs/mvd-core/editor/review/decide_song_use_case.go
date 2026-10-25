package review

import (
	"fmt"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlisteditor"
	"youtube-downloader/libs/mvd-core/playlistfile"
)

// Move moves the cursor by delta songs, and keeps it on the list.
func (s *Session) Move(delta int) {
	if len(s.Rows) == 0 {
		return
	}
	s.Selected = min(max(s.Selected+delta, 0), len(s.Rows)-1)
}

// Select puts the cursor on song i, when there is one.
func (s *Session) Select(i int) {
	if i >= 0 && i < len(s.Rows) {
		s.Selected = i
	}
}

// Open shows a video in the browser: the one that would be downloaded, or with original the upload
// that is in the playlist.
func (s *Session) Open(original bool) []Effect {
	row, ok := s.Current()
	if !ok {
		return nil
	}
	id, _, has := playlisteditor.Target(row)
	if original || !has {
		id = row.Upload.ID
	}
	if id == "" {
		s.Flash("there is no video to open")

		return nil
	}
	if !s.opts.CanOpen {
		s.Flash(playlisteditor.VideoAddress(id))

		return nil
	}

	return []Effect{OpenURL{URL: playlisteditor.VideoAddress(id)}}
}

// Opened is the host's answer to OpenURL: err is nil when the browser opened.
func (s *Session) Opened(url string, err error) {
	if err != nil {
		s.Flash("cannot open the browser: " + err.Error())

		return
	}
	s.Flash("opened " + url)
}

// Confirm accepts what was proposed for the song and moves on to the next one.
func (s *Session) Confirm() []Effect {
	row, ok := s.Current()
	if !ok {
		return nil
	}
	if _, _, has := playlisteditor.Target(playlistfile.Entry{Upload: row.Upload, TargetID: row.TargetID, Kind: row.Kind}); !has {
		s.Flash("nothing was found for this song: replace it, take the original or skip it")

		return nil
	}

	return s.Decide(playlistfile.DecisionConfirmed, "", "")
}

// ConfirmAll accepts what was proposed for every song nobody has decided on.
func (s *Session) ConfirmAll() []Effect {
	if s.live != nil {
		n := s.live.ReleaseAll()
		for i := range s.Rows {
			if s.Rows[i].Decision == playlistfile.DecisionNone && i < len(s.States) && s.States[i] == engine.StatePlanned {
				s.Rows[i].Decision = playlistfile.DecisionConfirmed
				s.remember(i)
			}
		}
		s.Refresh()
		s.Flash(fmt.Sprintf("%d songs sent to download as proposed", n))

		return nil
	}
	n := 0
	for i := range s.Rows {
		row := s.Rows[i]
		if row.Decision != playlistfile.DecisionNone || row.Downloaded {
			continue
		}
		if _, _, has := playlisteditor.Target(row); has {
			s.Rows[i].Decision = playlistfile.DecisionConfirmed
			n++
		}
	}
	s.autosave()
	s.Flash(fmt.Sprintf("confirmed %d songs", n))

	return nil
}

// Decide records what the person chose for the selected song. In a build made for testing a
// decision that goes against the proposal asks why.
func (s *Session) Decide(decision, chosenID, chosenTitle string) []Effect {
	row, ok := s.Current()
	if !ok {
		return nil
	}
	if row.Downloaded {
		s.Flash("this song has been downloaded already")

		return nil
	}
	if decision == playlistfile.DecisionOriginal && row.Upload.ID == "" {
		s.Flash("this song has no original upload")

		return nil
	}
	if s.live != nil && !s.releaseFor(decision, chosenID) {
		return nil
	}
	s.Rows[s.Selected].Decision = decision
	s.Rows[s.Selected].ChosenID, s.Rows[s.Selected].ChosenTitle = chosenID, chosenTitle
	if s.live != nil {
		s.remember(s.Selected)
	}
	s.autosave()
	// A decision hands the person the next song, as confirming always has; undoing one stays where
	// it is.
	if decision != playlistfile.DecisionNone && s.Selected < len(s.Rows)-1 {
		s.Move(1)
	}

	var effects []Effect
	if decision == playlistfile.DecisionReplaced && s.opts.CanDescribe {
		effects = append(effects, DescribeVideo{Key: row.Key(), VideoID: chosenID})
	}

	against := decision == playlistfile.DecisionReplaced || decision == playlistfile.DecisionOriginal || decision == playlistfile.DecisionSkipped
	if playlisteditor.DebugBuild && against {
		s.noteFor = row.Key()
		s.ask(inputNote, "Why? (Enter to leave it blank): ", "")
	}

	return effects
}

// Described is the host's answer to DescribeVideo: the title of the video that was picked.
func (s *Session) Described(key, title string) {
	for i := range s.Rows {
		if s.Rows[i].Key() == key && s.Rows[i].Decision == playlistfile.DecisionReplaced {
			s.Rows[i].ChosenTitle = title
		}
	}
	s.autosave()
}
