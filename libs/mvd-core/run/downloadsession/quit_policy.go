package downloadsession

// RequestQuit is the person asking to leave: a finished run leaves at once, one that goes on asks
// first. A hosted screen leaves it to the workspace.
func (s *Session) RequestQuit() Effect {
	if s.Hosted {
		return nil
	}
	if s.State.Idle {
		return Quit{}
	}
	s.ConfirmQuit = true

	return nil
}

// LeaveNow is the person leaving without being asked.
func (s *Session) LeaveNow() Effect { return Quit{} }

// AnswerQuit takes the answer to the question whether to leave while the downloads go: yes leaves,
// any other answer stays.
func (s *Session) AnswerQuit(yes bool) Effect {
	if yes {
		return Quit{}
	}
	s.ConfirmQuit = false

	return nil
}

// OpenHelp opens the help.
func (s *Session) OpenHelp() { s.ShowHelp = true }

// CloseHelp closes the help.
func (s *Session) CloseHelp() { s.ShowHelp = false }
