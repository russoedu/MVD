package api

import (
	"context"
	"errors"

	"youtube-downloader/libs/mvd-server/settings"
)

// ErrUninstallDeclined is what an Uninstaller returns when the person said no.
var ErrUninstallDeclined = errors.New("the person declined to remove the app")

// Uninstaller removes the app from the machine, after the person has agreed.
type Uninstaller interface {
	// Confirm puts what will be removed to the person on the machine's own screen and
	// waits. deletePreferences says whether the stored preferences go too, or nil to ask
	// the person as well. It returns ErrUninstallDeclined when they say no and
	// ErrNoDialog when there is no way to ask. Otherwise it returns the removal, which
	// the caller runs once it has answered the page: it ends with the app quitting.
	Confirm(deletePreferences *bool) (remove func(), err error)
}

// SettingsStore is where the settings live. *settings.Repository satisfies it.
type SettingsStore interface {
	Load() (settings.Document, error)
	// Save returns a settings.Invalid when the values are not acceptable.
	Save(settings.Settings) error
}

// FolderPicker asks the person to choose a folder, with whatever the machine offers.
type FolderPicker interface {
	// Pick shows a folder chooser starting at start (which may be empty or missing).
	// chosen is false if the person cancelled. It returns ErrNoDialog when this
	// machine has no way to show one.
	Pick(ctx context.Context, start string) (path string, chosen bool, err error)
}
