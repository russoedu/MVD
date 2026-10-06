package terminalui

import (
	"errors"

	"youtube-downloader/libs/mvd-core/tui"
	"youtube-downloader/libs/mvd-server/api"
)

// Uninstall adapts the app's uninstaller to what the preferences offer. The
// questions are put on the machine's own screen, as for the page, and the removal
// runs after this has returned its answer, because it ends with the app quitting.
func Uninstall(uninstaller api.Uninstaller) tui.Uninstaller {
	return func() (tui.RemovalOutcome, error) {
		remove, err := uninstaller.Confirm(nil)
		switch {
		case errors.Is(err, api.ErrUninstallDeclined):
			return tui.RemovalDeclined, nil
		case errors.Is(err, api.ErrNoDialog):
			return tui.RemovalUnavailable, nil
		case err != nil:
			return tui.RemovalDeclined, err
		}
		go remove()

		return tui.RemovalStarted, nil
	}
}
