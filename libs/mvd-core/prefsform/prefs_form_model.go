package prefsform

import "youtube-downloader/libs/mvd-core/config"

// Form is the preferences being edited: a tab at a time, and the setting the cursor is on. A
// change is in Cfg as soon as it is made; the form writes nothing itself. It is a value, so a
// copy keeps its own state. A renderer reads the exported fields and changes them only through
// the methods.
type Form struct {
	// Cfg is the settings as changed so far.
	Cfg   config.Config
	Pages []Page
	// Tab is the page shown and Cursor the setting of the page the cursor is on.
	Tab, Cursor int
	// Editing is true while a value is being typed; Before is what it was when typing began,
	// for Cancel to put back.
	Editing bool
	Before  string
	// Caret is the boundary of the name pattern the cursor is at, in units (see ParseTemplate).
	Caret int
	// Welcome is the first start: the renderer says what to do until a key is pressed.
	Welcome bool
	// Busy is true while a window of the host's is open and has the answer.
	Busy bool
	// Status is the last thing that happened, and StatusBad says it went wrong.
	Status    string
	StatusBad bool

	caps Capabilities
}

// New returns the form for cfg, on the first tab.
func New(cfg config.Config, caps Capabilities) Form {
	return Form{Cfg: cfg, Pages: Pages(caps), caps: caps}
}

// Current is the setting the cursor is on.
func (f Form) Current() Field { return f.Pages[f.Tab].Fields[f.Cursor] }

// TemplateLen is how many units (blocks and characters) the name pattern has.
func (f Form) TemplateLen() int {
	if f.Current().Kind != KindTemplate {
		return 0
	}

	return len(ParseTemplate(f.Current().Get(f.Cfg)))
}

// WithConfig replaces the settings being edited.
func (f Form) WithConfig(cfg config.Config) Form {
	f.Cfg = cfg

	return f
}

// WithWelcome makes the form of the first start.
func (f Form) WithWelcome() Form {
	f.Welcome = true

	return f
}
