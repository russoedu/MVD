package terminalui

import (
	"youtube-downloader/apps/mvd/install"
	"youtube-downloader/libs/mvd-core/tui"
)

// Move adapts the app's move to its own folder to what the preferences offer. The
// questions are put on the machine's own screen. When the app was moved, the moved copy
// has started and this one is told to quit through quit, after the answer has been given.
func Move(move func() install.MoveResult, quit func()) tui.Mover {
	return func() (tui.MoveOutcome, error) {
		switch move() {
		case install.Moved:
			go quit()

			return tui.MoveStarted, nil
		case install.AlreadyThere:
			return tui.MoveAlreadyThere, nil
		case install.CannotMove:
			return tui.MoveUnavailable, nil
		}

		return tui.MoveDeclined, nil
	}
}
