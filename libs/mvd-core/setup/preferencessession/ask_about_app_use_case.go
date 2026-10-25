package preferencessession

import "youtube-downloader/libs/mvd-core/settings"

// AskToUninstall starts the questions about removing the app; nothing when the host cannot remove
// the app.
func (s *Session) AskToUninstall() Effect {
	if !s.CanUninstall {
		return Effect{}
	}
	s.Mode = settings.EditConfirming
	s.Status = waitingForAnswersNotice

	return Effect{Kind: AskUninstall}
}

// AskToMove starts the questions about moving the app to its own folder; nothing when the host
// cannot move the app.
func (s *Session) AskToMove() Effect {
	if !s.CanMove {
		return Effect{}
	}
	s.Mode = settings.EditConfirming
	s.Status = waitingForAnswersNotice

	return Effect{Kind: AskMove}
}

// UninstallAnswered takes the answer about removing the app: what to tell the person.
func (s *Session) UninstallAnswered(outcome RemovalOutcome, err error) {
	s.Mode = settings.EditNone
	switch {
	case err != nil:
		s.Status = "The removal could not be asked: " + err.Error()
	case outcome == RemovalStarted:
		s.Status = "Removing MVD. This window will close."
	case outcome == RemovalDeclined:
		s.Status = "Nothing was removed."
	default:
		s.Status = "This machine cannot ask for confirmation, so nothing was removed. Start MVD with -uninstall from a terminal."
	}
}

// MoveAnswered takes the answer about moving the app: what to tell the person.
func (s *Session) MoveAnswered(outcome MoveOutcome, err error) {
	s.Mode = settings.EditNone
	switch {
	case err != nil:
		s.Status = "MVD could not be moved: " + err.Error()
	case outcome == MoveStarted:
		s.Status = "MVD was moved to its own folder. This window will close and the new copy will open."
	case outcome == MoveAlreadyThere:
		s.Status = "MVD already lives in its own folder."
	case outcome == MoveUnavailable:
		s.Status = "This machine has no place to move MVD to, or no way to ask. Start MVD with -move from a terminal."
	default:
		s.Status = "MVD was not moved."
	}
}
