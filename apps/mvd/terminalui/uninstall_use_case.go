package terminalui

import (
	"errors"

	"youtube-downloader/apps/mvd/uninstall"
	"youtube-downloader/libs/mvd-core/tui"
)

// Remover asks the person, on the machine's own screen, whether to remove the app and
// returns the removal if they agree; it is what the app's uninstall service offers.
type Remover interface {
	Confirm(deletePreferences *bool) (remove func(), err error)
}

// Uninstall adapts the app's uninstaller to what the preferences offer. The questions
// are put on the machine's own screen, and the removal runs after this has returned its
// answer, because it ends with the app quitting.
func Uninstall(remover Remover) tui.Uninstaller {
	return func() (tui.RemovalOutcome, error) {
		remove, err := remover.Confirm(nil)
		switch {
		case errors.Is(err, uninstall.ErrDeclined):
			return tui.RemovalDeclined, nil
		case errors.Is(err, uninstall.ErrNoDialog):
			return tui.RemovalUnavailable, nil
		case err != nil:
			return tui.RemovalDeclined, err
		}
		go remove()

		return tui.RemovalStarted, nil
	}
}
