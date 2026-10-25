package preferencessession

import (
	"strconv"
	"strings"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/settings"
)

// Move puts the cursor on another setting, delta settings away.
func (s *Session) Move(delta int) {
	s.Cursor = min(max(s.Cursor+delta, 0), ItemCount-1)
}

// Select puts the cursor on setting row, when none is being changed.
func (s *Session) Select(row int) {
	if !s.Editing() && row >= 0 && row < len(Labels) {
		s.Cursor = row
	}
}

// ClearStatus forgets what was said of what happened last.
func (s *Session) ClearStatus() { s.Status = "" }

// Activate does what Enter on the setting of the cursor means: a switch turns over, a folder asks
// for a folder, any other setting begins to be changed from the value it holds.
func (s *Session) Activate() Effect {
	switch s.Cursor {
	case itemSaveNotFound:
		s.Cfg.SaveNotFound = !s.Cfg.SaveNotFound
	case itemLogFile:
		s.Cfg.CreateLogFile = !s.Cfg.CreateLogFile
	case itemVideoQuality, itemAudioQuality, itemMergeFormat, itemOfficial, itemCookies:
		s.RadioOpts, s.RadioIdx = s.radioState()
		s.Mode = settings.EditRadio
	case itemConcurrent:
		s.Text = strconv.Itoa(s.Cfg.MaxConcurrentDownloads)
		s.Mode = settings.EditNumber
	case itemOutputDir, itemLogDir:
		return s.ChooseFolder()
	default:
		s.Text = s.textValue()
		s.Mode = settings.EditText
	}

	return Effect{}
}

// SetText is what the person has typed so far.
func (s *Session) SetText(text string) { s.Text = text }

// CancelEdit gives up the change being made.
func (s *Session) CancelEdit() { s.Mode = settings.EditNone }

// MoveChoice moves to another choice, delta choices away, while a choice is being made.
func (s *Session) MoveChoice(delta int) {
	s.RadioIdx = min(max(s.RadioIdx+delta, 0), len(s.RadioOpts)-1)
}

// StepNumber makes the number typed bigger or smaller by one; it does not go below one.
func (s *Session) StepNumber(delta int) {
	n := s.typedNumber()
	if delta < 0 && n <= 1 {
		return
	}
	s.Text = strconv.Itoa(n + delta)
}

// typedNumber is the number typed, or the one the setting holds when what is typed is not one.
func (s Session) typedNumber() int {
	n, err := strconv.Atoi(strings.TrimSpace(s.Text))
	if err != nil || n < 1 {
		return s.Cfg.MaxConcurrentDownloads
	}

	return n
}

// CommitEdit takes the value typed or chosen when the setting accepts it, and ends the change.
func (s *Session) CommitEdit() {
	s.commit()
	s.Mode = settings.EditNone
}

func (s *Session) commit() {
	text := strings.TrimSpace(s.Text)
	switch s.Cursor {
	case itemOutputDir:
		if text != "" {
			s.Cfg.OutputDir = text
		}
	case itemTemplate:
		if text != "" {
			s.Cfg.OutputTemplate = text
		}
	case itemLogDir:
		if text != "" {
			s.Cfg.LogDir = text
		}
	case itemConcurrent:
		if n, err := strconv.Atoi(text); err == nil && n > 0 {
			s.Cfg.MaxConcurrentDownloads = n
		}
	case itemVideoQuality:
		s.Cfg.VideoQuality = s.RadioOpts[s.RadioIdx]
	case itemAudioQuality:
		s.Cfg.AudioQuality = s.RadioOpts[s.RadioIdx]
	case itemMergeFormat:
		s.Cfg.MergeOutputFormat = s.RadioOpts[s.RadioIdx]
	case itemOfficial:
		s.Cfg.OfficialVideo = config.OfficialMode(s.RadioOpts[s.RadioIdx])
	case itemCookies:
		ApplyCookieChoice(&s.Cfg, s.RadioOpts[s.RadioIdx])
	}
}

// Keep asks the host to save the settings and leave.
func (s Session) Keep() Effect { return Effect{Kind: Save} }

// Leave asks the host to leave the screen.
func (s Session) Leave() Effect { return Effect{Kind: Cancel} }

// OpenAdvanced asks the host to show the advanced settings.
func (s Session) OpenAdvanced() Effect { return Effect{Kind: OpenAdvanced} }

// OpenColours asks the host to show the colours.
func (s Session) OpenColours() Effect { return Effect{Kind: OpenColours} }
