package downloadsession

import (
	"time"

	"youtube-downloader/libs/mvd-core/runstate"
	"youtube-downloader/libs/mvd-core/songlist"
)

// Pane is the part of the screen the person is in.
type Pane int

// The panes, in the order Tab goes through them.
const (
	PaneLists Pane = iota
	PaneEntries
	PaneOutput
	paneCount
)

// Session is the state of the download screen. It is a value: a copy keeps its own state, so a map
// or a slice in it is replaced and never changed in place. A renderer reads the fields; it changes
// them only through the methods.
type Session struct {
	// State is what the run has reported so far.
	State *runstate.State
	// Ctrl is the run behind the screen, nil when none is attached.
	Ctrl Controller

	Pane Pane
	// SelPlaylist and SelEntry are the playlist and the song (an index into the songs shown) the
	// cursor is on.
	SelPlaylist, SelEntry int
	// OutScroll is how many lines the output is scrolled up from the bottom.
	OutScroll int
	// Follow keeps the cursor on the song being downloaded.
	Follow bool
	// Loaded is set once every playlist has been listed; until then the screen is a loading screen.
	Loaded bool

	Order  songlist.Order
	Filter songlist.Filter
	// ColumnOrder is the order the person gave the columns they can arrange, none for the default.
	ColumnOrder []string
	// HiddenCols are the columns the person hid.
	HiddenCols map[string]bool

	FullLog bool
	// Technical shows the log of the tool in the details instead of the plain account.
	Technical bool
	// Paused is whether the downloads are held.
	Paused      bool
	ShowHelp    bool
	ConfirmQuit bool

	// Picked are the songs ticked in the table; the actions work on them, else on the song the
	// cursor is on. Anchor is the song the last tick was on, where a range with Shift begins.
	Picked    map[int]bool
	Anchor    int
	AnchorSet bool

	// Adding is whether the box for adding links is open, and LinksText what is in it.
	Adding    bool
	LinksText string
	// Swapping is whether the address of another video is being asked for song SwapEntry, and
	// SwapText what has been typed.
	Swapping  bool
	SwapText  string
	SwapEntry int

	// Notice is a message for the person, shown until NoticeUntil.
	Notice      string
	NoticeUntil time.Time

	// Editable is whether the list of links can be changed from the screen. LinkEditing is whether
	// the links of a run that goes on can be changed or taken out.
	Editable, LinkEditing bool
	// Hosted is whether the screen is the main part of a workspace, which owns quitting and adding
	// links.
	Hosted bool
}

// New is the session of a run: the screen waits for the run's playlists to be listed.
func New(c Controller) Session {
	state := runstate.New(c.Sources())

	return Session{
		Ctrl:   c,
		State:  state,
		Follow: true,
		// A run with no playlists to list (the download of a plan) has nothing to wait for.
		Loaded: len(state.Playlists) == 0,
	}
}

// NewHosted is the session of the main part of a workspace: it begins with no run, showing whatever
// state it is given, takes a run with Attach and lets it go when it ends.
func NewHosted() Session {
	return Session{Hosted: true, Loaded: true, Follow: true, State: runstate.New(nil)}
}
