package prefsform

// The preferences as a window draws them, in any look: the tabs, each setting with what it is and
// what it holds, and what the person does back. The shapes and their JSON names are the ones of
// apps/mvd-web's preferences-window contract, which reads them; the two change together.

// WindowKind is how a window draws a setting.
type WindowKind string

// The kinds, one per FieldKind.
const (
	WindowChoice   WindowKind = "choice"
	WindowNumber   WindowKind = "number"
	WindowText     WindowKind = "text"
	WindowFolder   WindowKind = "folder"
	WindowTemplate WindowKind = "template"
	WindowColour   WindowKind = "colour"
	WindowAction   WindowKind = "action"
)

// WindowOption is an option of a choice, with how it reads and the colours shown beside it.
type WindowOption struct {
	Value    string   `json:"value"`
	Label    string   `json:"label"`
	Swatches []string `json:"swatches,omitempty"`
}

// WindowUnit is a piece of the name pattern: a block (Block true) or text typed between blocks.
type WindowUnit struct {
	Text  string `json:"text"`
	Block bool   `json:"block,omitempty"`
}

// WindowField is a setting as a window draws it.
type WindowField struct {
	Label       string         `json:"label"`
	Hint        string         `json:"hint,omitempty"`
	Kind        WindowKind     `json:"kind"`
	Value       string         `json:"value"`
	Options     []WindowOption `json:"options,omitempty"`
	Min         int            `json:"min,omitempty"`
	Max         int            `json:"max,omitempty"`
	Step        int            `json:"step,omitempty"`
	Off         string         `json:"off,omitempty"`
	Placeholder string         `json:"placeholder,omitempty"`
	Optional    bool           `json:"optional,omitempty"`
	// Hidden is a setting that does not apply to the settings as they are; it keeps its place, so
	// the intents of the others keep their numbers.
	Hidden bool `json:"hidden,omitempty"`
	// Units, Blocks and Example are the name pattern's: its pieces, the blocks that can be added,
	// and the name it gives a sample song.
	Units   []WindowUnit `json:"units,omitempty"`
	Blocks  []string     `json:"blocks,omitempty"`
	Example string       `json:"example,omitempty"`
}

// WindowColours are the colours of the terminal look, for a window drawn in it.
type WindowColours struct {
	Background string `json:"background"`
	Accent     string `json:"accent"`
	Focus      string `json:"focus"`
	Highlight  string `json:"highlight"`
	Success    string `json:"success"`
	Error      string `json:"error"`
	Dim        string `json:"dim"`
	Text       string `json:"text"`
	Selected   string `json:"selected"`
}

// Window is the preferences as a window draws them.
type Window struct {
	Title string `json:"title"`
	// Welcome is what the first start says, until the person does something.
	Welcome string        `json:"welcome,omitempty"`
	Tabs    []string      `json:"tabs"`
	Tab     int           `json:"tab"`
	Fields  []WindowField `json:"fields"`
	// Status is the last thing that happened, Bad when it went wrong; Busy while a window of the
	// host's (the folder chooser, a question) has the answer.
	Status  string        `json:"status,omitempty"`
	Bad     bool          `json:"bad,omitempty"`
	Busy    bool          `json:"busy,omitempty"`
	Colours WindowColours `json:"colours"`
}

// WindowIntentKind is what the person did in the preferences window.
type WindowIntentKind string

// The kinds of intent.
const (
	// IntentTab shows tab Index.
	IntentTab WindowIntentKind = "prefs-tab"
	// IntentSet gives setting Field the value Text: an option, a number or a text.
	IntentSet WindowIntentKind = "prefs-set"
	// IntentActivate does what the setting's button does: choose a folder, or run the action.
	IntentActivate WindowIntentKind = "prefs-activate"
	// IntentDefaultColour puts a colour back to its default.
	IntentDefaultColour WindowIntentKind = "prefs-default-colour"
	// IntentAddBlock adds block Index of the palette at the end of the name pattern;
	// IntentRemoveUnit takes piece Index out; IntentAddText types Text at its end.
	IntentAddBlock   WindowIntentKind = "prefs-add-block"
	IntentRemoveUnit WindowIntentKind = "prefs-remove-unit"
	IntentAddText    WindowIntentKind = "prefs-add-text"
	// IntentSave keeps the settings and closes; IntentClose closes.
	IntentSave  WindowIntentKind = "prefs-save"
	IntentClose WindowIntentKind = "prefs-close"
)

// WindowIntent is what the person did, with the setting it was on.
type WindowIntent struct {
	Kind  WindowIntentKind `json:"kind"`
	Field int              `json:"field,omitempty"`
	Index int              `json:"index,omitempty"`
	Text  string           `json:"text,omitempty"`
}
