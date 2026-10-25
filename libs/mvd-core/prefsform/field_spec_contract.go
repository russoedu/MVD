package prefsform

import (
	"strings"

	"youtube-downloader/libs/mvd-core/config"
)

// FieldKind is how a setting is drawn and changed.
type FieldKind int

const (
	KindChoice   FieldKind = iota // one of a few options, all listed
	KindStepper                   // a number changed by one step at a time
	KindText                      // a line of text
	KindFolder                    // a folder, chosen in the host's chooser
	KindTemplate                  // the name pattern of the files, with blocks to add to it
	KindColour                    // a #rgb or #rrggbb colour, with its swatch
	KindAction                    // a button that asks the host to do something
)

// ActionKind is which of the host's actions a button asks for.
type ActionKind int

const (
	ActionNone ActionKind = iota
	ActionMove
	ActionUninstall
)

// Field is one setting of a page.
type Field struct {
	Label       string
	Kind        FieldKind
	Placeholder string // shown dim when a text is empty
	Hint        string // a dim line of help under the label
	Options     []string
	Show        func(option string) string // how an option reads; nil capitalises it
	Min, Max    int
	Step        int
	Off         string // what a stepper at 0 reads, when it is not "0"
	Get         func(config.Config) string
	Set         func(*config.Config, string)
	Action      ActionKind
	// Slot, for a colour, points at the colour inside the theme; Optional says it may be left
	// empty, for the terminal's own.
	Slot     func(*config.ThemeColors) *string
	Optional bool
	// Swatches, for a choice, are the colours shown beside an option.
	Swatches func(option string) []string
}

// Page is a tab: its name and the settings under it.
type Page struct {
	Name   string
	Fields []Field
}

// Capabilities are what the host can do for the form, which decides what it offers. Each one is
// optional.
type Capabilities struct {
	// CanPickFolder says the host shows a folder chooser; without it a folder is typed.
	CanPickFolder bool
	// CanMove says the host can move the app to its own folder.
	CanMove bool
	// CanUninstall says the host can remove the app.
	CanUninstall bool
}

// OptionText is how an option of a field reads.
func (f Field) OptionText(option string) string {
	if f.Show != nil {
		return f.Show(option)
	}

	return capitalise(option)
}

// NumberText is how a number of a field reads.
func (f Field) NumberText(value string) string {
	if value == "0" && f.Off != "" {
		return f.Off
	}

	return value
}

func capitalise(s string) string {
	if s == "" {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}
