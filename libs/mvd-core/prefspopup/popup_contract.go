package prefspopup

import "youtube-downloader/libs/mvd-core/prefsform"

// ExampleName is what a file would be called by a yt-dlp output template, for a sample song (see
// prefsform.ExampleName).
func ExampleName(template, ext string) string { return prefsform.ExampleName(template, ext) }

// Host is what the popup cannot do itself. Each one is optional.
type Host struct {
	// PickFolder shows the operating system's folder chooser, starting at start.
	PickFolder func(start string) (path string, chosen bool, err error)
	// Move asks to move the app to its own folder and says what came of it.
	Move func() (message string, err error)
	// Uninstall asks to remove the app and says what came of it.
	Uninstall func() (message string, err error)
	// StylesWindow says the host draws its window in the classic systems' looks too, so the
	// preferences offer the choice.
	StylesWindow bool
}

// CloseMsg is sent when the person leaves the popup with esc.
type CloseMsg struct{}

// SaveMsg is sent when the person asks to save with the keyboard.
type SaveMsg struct{}

// pickedMsg is the answer of the folder chooser.
type pickedMsg struct {
	field  int
	path   string
	chosen bool
	err    error
}

// doneMsg is what an action came to.
type doneMsg struct {
	message string
	err     error
}
