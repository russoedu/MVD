package downloadsession

import (
	"strings"

	"youtube-downloader/libs/mvd-core/songlist"
)

// The columns of the table of songs the person can arrange.
const (
	ColumnArtist   = songlist.ColumnArtist
	ColumnVideo    = songlist.ColumnVideo
	ColumnAlbum    = songlist.ColumnAlbum
	ColumnQuality  = songlist.ColumnQuality
	ColumnProgress = songlist.ColumnProgress
	ColumnStatus   = songlist.ColumnStatus
)

// Column is a column the person can hide, and what the list of columns calls it.
type Column struct{ Name, Label string }

// Configurable are the columns the person can hide, in the order the list of columns has them to
// begin with.
var Configurable = []Column{
	{ColumnArtist, "Artist"}, {ColumnVideo, "Video name"}, {ColumnAlbum, "Album"},
	{ColumnQuality, "Quality"}, {ColumnProgress, "Progress"}, {ColumnStatus, "Status"},
}

// LabelOf is how the list of columns calls a column.
func LabelOf(name string) string {
	for _, c := range Configurable {
		if c.Name == name {
			return c.Label
		}
	}

	return name
}

// Resizable says whether a column can be given another width by the person.
func Resizable(name string) bool {
	return strings.Contains(" artist video album quality progress status ", " "+name+" ")
}

// Sortable says whether a click on the heading of a column sorts the songs by it.
func Sortable(name string) bool { return name != "mark" && name != "check" }

// ArrangedColumns are the columns the person can arrange, in the order they are shown: the one they
// gave, else the order of Configurable.
func (s Session) ArrangedColumns() []string {
	order := make([]string, 0, len(Configurable))
	seen := map[string]bool{}
	for _, name := range s.ColumnOrder {
		if Resizable(name) && !seen[name] {
			order, seen[name] = append(order, name), true
		}
	}
	for _, c := range Configurable {
		if !seen[c.Name] {
			order = append(order, c.Name)
		}
	}

	return order
}

// MoveColumn puts the column that is at row from of the list at row to.
func (s *Session) MoveColumn(from, to int) {
	order := s.ArrangedColumns()
	if from < 0 || from >= len(order) || to < 0 || to >= len(order) || from == to {
		return
	}
	moved := order[from]
	rest := append(append([]string(nil), order[:from]...), order[from+1:]...)
	s.ColumnOrder = append(append(append([]string(nil), rest[:to]...), moved), rest[to:]...)
}

// ToggleColumn shows or hides a column.
func (s *Session) ToggleColumn(name string) {
	hidden := make(map[string]bool, len(s.HiddenCols)+1)
	for k, v := range s.HiddenCols {
		hidden[k] = v
	}
	hidden[name] = !hidden[name]
	s.HiddenCols = hidden
}

// ControlsShown says whether the buttons of the table are there: the playlist has songs, also when
// the filter hides them all.
func (s Session) ControlsShown() bool {
	if s.SelPlaylist >= len(s.State.Playlists) {
		return false
	}
	pl := s.State.Playlists[s.SelPlaylist]

	return pl.Err == "" && pl.Listed && len(pl.Entries) > 0
}
