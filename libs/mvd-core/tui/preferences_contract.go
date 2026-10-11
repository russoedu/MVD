package tui

import preferencespage "youtube-downloader/libs/mvd-core/setup/preferences"

// The contracts of the preferences screen are owned by its slice; a host reaches them
// here, with the setup input.

// FolderPicker shows the operating system's own folder chooser (see the preferences page).
type FolderPicker = preferencespage.FolderPicker

// MoveOutcome is how asking to move the app to its own folder went.
type MoveOutcome = preferencespage.MoveOutcome

const (
	MoveStarted      = preferencespage.MoveStarted
	MoveDeclined     = preferencespage.MoveDeclined
	MoveAlreadyThere = preferencespage.MoveAlreadyThere
	MoveUnavailable  = preferencespage.MoveUnavailable
)

// Mover asks the person whether to move the app to its own folder, and moves it.
type Mover = preferencespage.Mover

// RemovalOutcome is how asking to remove the app went.
type RemovalOutcome = preferencespage.RemovalOutcome

const (
	RemovalStarted     = preferencespage.RemovalStarted
	RemovalDeclined    = preferencespage.RemovalDeclined
	RemovalUnavailable = preferencespage.RemovalUnavailable
)

// Uninstaller asks the person whether to remove the app, and starts the removal.
type Uninstaller = preferencespage.Uninstaller
