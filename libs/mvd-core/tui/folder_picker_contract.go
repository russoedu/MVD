package tui

// FolderPicker shows the operating system's own folder chooser, starting at start,
// and waits for the answer: the folder chosen, or chosen false when the person closed
// the chooser without choosing. It may take as long as the person does. A host passes
// one to the setup screens when it can show a chooser; a screen that gets an error from
// it (no desktop, no tool) falls back to its own folder browser.
type FolderPicker func(start string) (path string, chosen bool, err error)

// folderPickedMsg is the answer of a FolderPicker, for the setting that asked.
type folderPickedMsg struct {
	item   int
	path   string
	chosen bool
	err    error
}
