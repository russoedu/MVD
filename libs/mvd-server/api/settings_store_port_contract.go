package api

import (
	"context"

	"youtube-downloader/libs/mvd-server/settings"
)

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
