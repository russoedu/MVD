package tui

// AppInput seeds an app model.
type AppInput struct {
	Setup SetupInput
	Start RunStarter
	// Editor is what the playlist editor needs; its zero value offers an editor that can
	// only open files.
	Editor EditorHost
	// ReviewOnStart opens the playlist editor on the list at once, instead of the list.
	ReviewOnStart bool
}
