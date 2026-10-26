package windowsnapshot

// Snapshot is what the main window shows.
type Snapshot struct {
	Title    string   `json:"title"`
	Menus    []Menu   `json:"menus"`
	Toolbar  []Button `json:"toolbar"`
	Sources  Sources  `json:"sources"`
	Songs    Songs    `json:"songs"`
	Details  Details  `json:"details"`
	Status   Status   `json:"status"`
	Activity Activity `json:"activity"`
	// Adding is the window for adding music, when it is open.
	Adding *AddMusic `json:"adding,omitempty"`
	// Covered is a screen only the terminal look draws so far (the preferences, a question...),
	// when one is open over the window: the theme shows the terminal for it.
	Covered *Covered `json:"covered,omitempty"`
}

// Activity is what MVD is doing, for the theme's pointer.
type Activity string

// The activities.
const (
	ActivityIdle       Activity = "idle"
	ActivityBackground Activity = "background"
	ActivityWaiting    Activity = "waiting"
)

// Shortcut is a key with the modifiers held for it, so each theme writes it its own way.
type Shortcut struct {
	Ctrl  bool   `json:"ctrl,omitempty"`
	Shift bool   `json:"shift,omitempty"`
	Alt   bool   `json:"alt,omitempty"`
	Key   string `json:"key"`
}

// Menu is a heading of the menu bar and its entries.
type Menu struct {
	Label    string      `json:"label"`
	Mnemonic string      `json:"mnemonic,omitempty"`
	Entries  []MenuEntry `json:"entries"`
}

// MenuEntry is an entry of a menu. ID is what a menu intent sends back.
type MenuEntry struct {
	ID              string    `json:"id"`
	Label           string    `json:"label"`
	Mnemonic        string    `json:"mnemonic,omitempty"`
	Shortcut        *Shortcut `json:"shortcut,omitempty"`
	Enabled         bool      `json:"enabled"`
	SeparatorBefore bool      `json:"separatorBefore,omitempty"`
}

// ButtonRole is what a button means in its group: a theme places the default and cancel ones.
type ButtonRole string

// The roles.
const (
	RoleDefault ButtonRole = "default"
	RoleCancel  ButtonRole = "cancel"
	RoleOther   ButtonRole = "other"
)

// Button is a button. ID is what a button intent sends back.
type Button struct {
	ID             string     `json:"id"`
	Label          string     `json:"label"`
	Mnemonic       string     `json:"mnemonic,omitempty"`
	Role           ButtonRole `json:"role"`
	Enabled        bool       `json:"enabled"`
	DisabledReason string     `json:"disabledReason,omitempty"`
}

// Tone is how a line reads: a theme picks the colour.
type Tone string

// The tones.
const (
	ToneNormal Tone = "normal"
	ToneDim    Tone = "dim"
	ToneGood   Tone = "good"
	ToneBad    Tone = "bad"
	ToneBusy   Tone = "busy"
)

// SourceRow is a source of the list.
type SourceRow struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Detail string `json:"detail"`
	Tone   Tone   `json:"tone"`
}

// Sources is the list of sources and which one is chosen.
type Sources struct {
	Heading  string      `json:"heading"`
	Rows     []SourceRow `json:"rows"`
	Selected int         `json:"selected"`
}

// SongRow is a song of the table.
type SongRow struct {
	ID       string   `json:"id"`
	Ticked   bool     `json:"ticked"`
	Artist   string   `json:"artist"`
	Title    string   `json:"title"`
	Progress Progress `json:"progress"`
	Status   string   `json:"status"`
	Tone     Tone     `json:"tone"`
}

// Songs is the table of songs and which row is chosen.
type Songs struct {
	Heading  string    `json:"heading"`
	Rows     []SongRow `json:"rows"`
	Selected int       `json:"selected"`
}

// Details is the pane under the lists.
type Details struct {
	Heading string   `json:"heading"`
	Lines   []string `json:"lines"`
}

// KeyHint is a key on offer and what it does.
type KeyHint struct {
	Shortcut Shortcut `json:"shortcut"`
	Label    string   `json:"label"`
}

// Status is the line at the bottom.
type Status struct {
	Summary string    `json:"summary"`
	Hints   []KeyHint `json:"hints"`
}

// The ways of saying what to add, and the fields of the window for adding music.
const (
	ModeURL     = "url"
	ModeSearch  = "search"
	FieldURL    = "url"
	FieldArtist = "artist"
	FieldAlbum  = "album"
	FieldSong   = "song"
)

// AddMode is one way of saying what to add.
type AddMode struct {
	Mode     string `json:"mode"`
	Label    string `json:"label"`
	Mnemonic string `json:"mnemonic,omitempty"`
}

// AddField is a field of the window for adding music, with its text.
type AddField struct {
	Field    string `json:"field"`
	Label    string `json:"label"`
	Mnemonic string `json:"mnemonic,omitempty"`
	Text     string `json:"text"`
}

// AddMusic is the window for adding music.
type AddMusic struct {
	Heading string     `json:"heading"`
	Mode    string     `json:"mode"`
	Modes   []AddMode  `json:"modes"`
	Fields  []AddField `json:"fields"`
	Help    []string   `json:"help"`
	Problem string     `json:"problem,omitempty"`
	Buttons []Button   `json:"buttons"`
}

// Covered is a screen open over the window that only the terminal look draws so far.
type Covered struct {
	Title string `json:"title"`
}
