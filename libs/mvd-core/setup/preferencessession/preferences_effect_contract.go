package preferencessession

// EffectKind is what the host has to do after an intent.
type EffectKind int

const (
	// NoEffect needs nothing from the host.
	NoEffect EffectKind = iota
	// Save asks the host to keep the settings and leave the screen.
	Save
	// Cancel asks the host to leave the screen.
	Cancel
	// OpenAdvanced asks the host to show the advanced settings.
	OpenAdvanced
	// OpenColours asks the host to show the colours.
	OpenColours
	// PickFolder asks the host to show the machine's own folder chooser, starting at Folder, and to
	// give the answer back with FolderPicked, naming Item.
	PickFolder
	// BrowseFolders asks the host to show its own folder browser, starting at Folder, and to say
	// what was chosen with FolderChosen or FolderCancelled.
	BrowseFolders
	// AskUninstall asks the host to ask the person about removing the app, and to give the answer
	// back with UninstallAnswered.
	AskUninstall
	// AskMove asks the host to ask the person about moving the app, and to give the answer back
	// with MoveAnswered.
	AskMove
)

// Effect is what the host has to do, with what it needs to do it.
type Effect struct {
	Kind EffectKind
	// Folder is where a chooser or a browser starts, and Item the setting it is for.
	Folder string
	Item   int
}
