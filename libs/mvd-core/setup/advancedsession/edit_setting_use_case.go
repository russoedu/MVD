package advancedsession

import (
	"strconv"
	"strings"

	"youtube-downloader/libs/mvd-core/settings"
)

// Move puts the cursor on another setting, delta settings away.
func (s *Session) Move(delta int) {
	s.Cursor = min(max(s.Cursor+delta, 0), len(Labels)-1)
}

// Select puts the cursor on setting row, when none is being typed.
func (s *Session) Select(row int) {
	if !s.Editing() && row >= 0 && row < len(Labels) {
		s.Cursor = row
	}
}

// Activate does what Enter on the setting of the cursor means: a switch turns over, any other
// setting begins to be typed from the value it holds.
func (s *Session) Activate() {
	switch s.Cursor {
	case itemExtraArgs:
		s.Text = strings.Join(s.Cfg.ExtraArgs, " ")
		s.Mode = settings.EditText
	case itemFragments:
		s.Text = strconv.Itoa(s.Cfg.ConcurrentFragments)
		s.Mode = settings.EditNumber
	case itemAutoRetry:
		s.Cfg.AutoRetry = !s.Cfg.AutoRetry
	case itemRetryCycles, itemRetryWait, itemStartTimeout:
		s.Text = strconv.Itoa(s.current())
		s.Mode = settings.EditNumber
	}
}

// SetText is what the person has typed so far.
func (s *Session) SetText(text string) { s.Text = text }

// StepNumber makes the number typed bigger or smaller by one; it does not go below zero.
func (s *Session) StepNumber(delta int) {
	n := s.typedNumber()
	if delta < 0 && n == 0 {
		return
	}
	s.Text = strconv.Itoa(n + delta)
}

// CancelEdit gives up the value being typed.
func (s *Session) CancelEdit() { s.Mode = settings.EditNone }

// CommitEdit takes the value typed when the setting accepts it, and ends the typing. A value the
// setting does not accept leaves it as it was.
func (s *Session) CommitEdit() {
	s.commit()
	s.Mode = settings.EditNone
}

func (s *Session) commit() {
	number := func(least int) (int, bool) {
		n, err := strconv.Atoi(strings.TrimSpace(s.Text))

		return n, err == nil && n >= least
	}
	switch s.Cursor {
	case itemExtraArgs:
		s.Cfg.ExtraArgs = strings.Fields(s.Text)
	case itemFragments:
		if n, ok := number(0); ok {
			s.Cfg.ConcurrentFragments = n
		}
	case itemRetryCycles:
		if n, ok := number(1); ok {
			s.Cfg.RetryCycles = n
		}
	case itemRetryWait:
		if n, ok := number(1); ok {
			s.Cfg.RetryWaitSeconds = n
		}
	case itemStartTimeout:
		if n, ok := number(1); ok {
			s.Cfg.StartTimeoutSeconds = n
		}
	}
}

// Keep asks the host to save the settings and leave.
func (s Session) Keep() Effect { return Save }

// Leave asks the host to leave the screen.
func (s Session) Leave() Effect { return Cancel }
