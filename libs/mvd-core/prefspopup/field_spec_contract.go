package prefspopup

import (
	"youtube-downloader/libs/mvd-core/config"
)

// fieldKind is how a setting is drawn and changed.
type fieldKind int

const (
	kindChoice   fieldKind = iota // one of a few options, all listed
	kindStepper                   // a number changed by one step at a time
	kindText                      // a line of text
	kindFolder                    // a folder, chosen in the host's chooser
	kindTemplate                  // the name pattern of the files, with blocks to add to it
	kindColour                    // a #rgb or #rrggbb colour, with its swatch
	kindAction                    // a button that asks the host to do something
)

// actionKind is which of the host's actions a button asks for.
type actionKind int

const (
	actionNone actionKind = iota
	actionMove
	actionUninstall
)

// field is one setting of a page.
type field struct {
	label       string
	kind        fieldKind
	placeholder string // shown dim when a text is empty
	hint        string // a dim line of help under the label
	options     []string
	show        func(option string) string // how an option reads; nil capitalises it
	min, max    int
	step        int
	off         string // what a stepper at 0 reads, when it is not "0"
	get         func(config.Config) string
	set         func(*config.Config, string)
	action      actionKind
}

// Host is what the popup cannot do itself. Each one is optional.
type Host struct {
	// PickFolder shows the operating system's folder chooser, starting at start.
	PickFolder func(start string) (path string, chosen bool, err error)
	// Move asks to move the app to its own folder and says what came of it.
	Move func() (message string, err error)
	// Uninstall asks to remove the app and says what came of it.
	Uninstall func() (message string, err error)
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
