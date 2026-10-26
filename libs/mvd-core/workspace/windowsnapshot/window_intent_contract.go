package windowsnapshot

// IntentKind is what the person did in a themed window.
type IntentKind string

// The kinds of intent.
const (
	IntentMenu         IntentKind = "menu"
	IntentButton       IntentKind = "button"
	IntentSelectSource IntentKind = "select-source"
	IntentSelectSong   IntentKind = "select-song"
	IntentTickSong     IntentKind = "tick-song"
	IntentWindow       IntentKind = "window"
	IntentAddingMode   IntentKind = "adding-mode"
	IntentAddingText   IntentKind = "adding-text"
	IntentAddingButton IntentKind = "adding-button"
	// IntentKey is a key the window does not use itself, such as a letter that accepts a song,
	// for MVD to act on as the terminal look does.
	IntentKey IntentKind = "key"
)

// Intent is what the person did, with what it needs: the id of a menu entry or a button, the row
// of a list, whether a song is ticked, a window button, a mode, a field and its text, or a key.
type Intent struct {
	Kind   IntentKind `json:"kind"`
	ID     string     `json:"id,omitempty"`
	Index  int        `json:"index,omitempty"`
	Ticked bool       `json:"ticked,omitempty"`
	Action string     `json:"action,omitempty"`
	Mode   string     `json:"mode,omitempty"`
	Field  string     `json:"field,omitempty"`
	Text   string     `json:"text,omitempty"`
	Key    string     `json:"key,omitempty"`
}
