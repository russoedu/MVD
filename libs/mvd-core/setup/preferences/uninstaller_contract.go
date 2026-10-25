package preferences

import "youtube-downloader/libs/mvd-core/setup/preferencessession"

// RemovalOutcome is how asking to remove the app went.
type RemovalOutcome = preferencessession.RemovalOutcome

const (
	// RemovalStarted: the person agreed and the removal is under way. It ends with
	// the app quitting.
	RemovalStarted = preferencessession.RemovalStarted
	// RemovalDeclined: the person said no; nothing was removed.
	RemovalDeclined = preferencessession.RemovalDeclined
	// RemovalUnavailable: this machine has no way to ask, so nothing was removed.
	RemovalUnavailable = preferencessession.RemovalUnavailable
)

// Uninstaller asks the person, on the machine's own screen, whether to remove the
// app, and starts the removal if they agree. It waits for the answers, which may take
// as long as the person does. A host that can remove itself passes one to the setup
// screens, which then offer it on the preferences.
type Uninstaller func() (RemovalOutcome, error)

// uninstallAnsweredMsg is the answer of an Uninstaller.
type uninstallAnsweredMsg struct {
	outcome RemovalOutcome
	err     error
}
