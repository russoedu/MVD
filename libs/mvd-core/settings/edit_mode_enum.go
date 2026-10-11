// Package settings has what the screens that edit settings share: the modes an item can
// be edited in, and how a switch reads.
package settings

// Mode is how a setting is being changed.
type Mode int

const (
	EditNone       Mode = iota // nothing is being edited
	EditText                   // a text field
	EditNumber                 // a number field
	EditRadio                  // a choice among options
	EditFolder                 // the folder picker
	EditPicking                // the operating system's folder chooser is open
	EditConfirming             // the questions about removing the app are being asked
)
