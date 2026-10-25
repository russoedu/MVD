package coloursession

import (
	"strings"

	"youtube-downloader/libs/mvd-core/config"
)

// Move puts the cursor on another colour, delta colours away, when none is being typed.
func (s *Session) Move(delta int) {
	if s.Editing {
		return
	}
	s.Cursor = min(max(s.Cursor+delta, 0), len(Labels)-1)
}

// Select puts the cursor on colour row, when none is being typed.
func (s *Session) Select(row int) {
	if !s.Editing && row >= 0 && row < len(Labels) {
		s.Cursor = row
	}
}

// BeginEdit starts typing the colour of the cursor, from the colour it has.
func (s *Session) BeginEdit() {
	s.Text = s.Value(s.Cursor)
	s.Editing = true
	s.Status = ""
}

// SetText is what the person has typed so far.
func (s *Session) SetText(text string) { s.Text = text }

// CancelEdit gives up the colour being typed.
func (s *Session) CancelEdit() {
	s.Editing = false
	s.Status = ""
}

// AcceptEdit takes the colour typed when it is one, and says so when it is not. The host is told to
// draw with it.
func (s *Session) AcceptEdit() Effect {
	val := strings.TrimSpace(s.Text)
	if !config.ValidColour(val) {
		s.Status = InvalidColourNotice

		return NoEffect
	}
	*s.slot(s.Cursor) = val
	s.Editing = false
	s.Status = ""

	return ApplyColours
}

// RestoreDefault puts the colour of the cursor back to the one the app comes with.
func (s *Session) RestoreDefault() Effect {
	*s.slot(s.Cursor) = defaultFor(s.Cursor)

	return ApplyColours
}

// Keep asks the host to save the colours and leave.
func (s Session) Keep() Effect { return Save }

// Leave asks the host to leave the screen.
func (s Session) Leave() Effect { return Cancel }
