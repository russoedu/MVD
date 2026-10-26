package addsource

// Mode is how the person says what to add.
type Mode int

// The modes, in the order they are offered.
const (
	// ModeURL is a link to a playlist or a video of a platform. It is the one chosen at first.
	ModeURL Mode = iota
	// ModeSearch is music asked for by name: an artist, an album, a song.
	ModeSearch
)

// Modes are what the two modes are called, in the order they are offered.
var Modes = []string{"Platform URL", "Music search"}

// Field is a place the focus can be.
type Field int

// The places of the window.
const (
	FieldMode Field = iota
	FieldURL
	FieldArtist
	FieldAlbum
	FieldSong
	FieldAdd
	FieldCancel
)

// Label is what a text field is called.
func (f Field) Label() string {
	switch f {
	case FieldURL:
		return "Link"
	case FieldArtist:
		return "Artist"
	case FieldAlbum:
		return "Album"
	case FieldSong:
		return "Song"
	}

	return ""
}

// State is the window: the mode, what is written in each field, where the focus is and what is wrong
// with the last try. It is a value: a change gives a new state.
type State struct {
	Mode                Mode
	Focus               Field
	URLs                string
	Artist, Album, Song string
	// Problem is why the last try to add did not work, in words for the person.
	Problem string
}

// New is the window as it opens: the platform address is chosen and the focus is on its field.
func New() State { return State{Mode: ModeURL, Focus: FieldURL} }

// Fields are the places the focus goes through with the Tab key, in order, for the mode.
func (s State) Fields() []Field {
	if s.Mode == ModeSearch {
		return []Field{FieldMode, FieldArtist, FieldAlbum, FieldSong, FieldAdd, FieldCancel}
	}

	return []Field{FieldMode, FieldURL, FieldAdd, FieldCancel}
}

// WithMode chooses a mode. What was written in the other is kept, for the person to come back to.
func (s State) WithMode(mode Mode) State {
	if s.Mode == mode {
		return s
	}
	s.Mode, s.Problem = mode, ""
	if IsText(s.Focus) {
		s.Focus = s.Fields()[1]
	}

	return s
}

// Text is what is written in a field.
func (s State) Text(f Field) string {
	switch f {
	case FieldURL:
		return s.URLs
	case FieldArtist:
		return s.Artist
	case FieldAlbum:
		return s.Album
	case FieldSong:
		return s.Song
	}

	return ""
}

// SetText says what is written in a field.
func (s State) SetText(f Field, text string) State {
	switch f {
	case FieldURL:
		s.URLs = text
	case FieldArtist:
		s.Artist = text
	case FieldAlbum:
		s.Album = text
	case FieldSong:
		s.Song = text
	}
	s.Problem = ""

	return s
}

// IsText says whether the field is one that is written in.
func IsText(f Field) bool { return f >= FieldURL && f <= FieldSong }

// Next moves the focus on, round to the first.
func (s State) Next() State { return s.step(1) }

// Previous moves the focus back, round to the last.
func (s State) Previous() State { return s.step(-1) }

func (s State) step(by int) State {
	fields := s.Fields()
	at := 0
	for i, f := range fields {
		if f == s.Focus {
			at = i
		}
	}
	s.Focus = fields[(at+by+len(fields))%len(fields)]

	return s
}

// Focused is the state with the focus put on a field of its mode; any other is ignored.
func (s State) Focused(f Field) State {
	for _, own := range s.Fields() {
		if own == f {
			s.Focus = f
		}
	}

	return s
}
