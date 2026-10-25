package songlist

import (
	"sort"
	"strings"

	"youtube-downloader/libs/mvd-core/runstate"
)

// The columns the songs can be sorted by.
const (
	ColumnIdx      = "idx"
	ColumnSong     = "song"
	ColumnArtist   = "artist"
	ColumnVideo    = "video"
	ColumnAlbum    = "album"
	ColumnQuality  = "quality"
	ColumnProgress = "progress"
	ColumnStatus   = "status"
)

// Order is the column the songs are sorted by, "" for the order they came in.
type Order struct {
	By   string
	Desc bool
}

// current is the column that sorts: the songs are in the order of the # column to begin with.
func (o Order) current() string {
	if o.By == "" {
		return ColumnIdx
	}

	return o.By
}

// Clicked is the order after a click on the heading of a column: the column that sorts already
// turns its direction round, any other begins ascending.
func (o Order) Clicked(column string) Order {
	if o.current() == column {
		return Order{By: column, Desc: !o.Desc}
	}

	return Order{By: column}
}

// Direction says how a column is sorted: 0 when it does not sort, 1 ascending, -1 descending.
func (o Order) Direction(column string) int {
	switch {
	case o.current() != column:
		return 0
	case o.Desc:
		return -1
	}

	return 1
}

// Sorted puts the songs in the order, else as they are. ids index entries.
func (o Order) Sorted(ids []int, entries []*runstate.Entry) []int {
	if o.By == "" {
		return ids
	}
	out := append([]int(nil), ids...)
	sort.SliceStable(out, func(i, j int) bool {
		c := Compare(o.By, entries[out[i]], entries[out[j]])
		if o.Desc {
			return c > 0
		}

		return c < 0
	})

	return out
}

// Compare orders two songs by a column: negative when a comes first.
func Compare(column string, a, b *runstate.Entry) int {
	text := func(x, y string) int { return strings.Compare(strings.ToLower(x), strings.ToLower(y)) }
	number := func(x, y float64) int {
		switch {
		case x < y:
			return -1
		case x > y:
			return 1
		}

		return 0
	}
	switch column {
	case ColumnIdx:
		return number(float64(a.Index), float64(b.Index))
	case ColumnSong:
		return text(Name(a), Name(b))
	case ColumnArtist:
		return text(Artist(a), Artist(b))
	case ColumnVideo:
		return text(VideoTitle(a), VideoTitle(b))
	case ColumnAlbum:
		return text(a.NamedAlbum, b.NamedAlbum)
	case ColumnQuality:
		return number(float64(a.Height*1000+a.AudioKbps), float64(b.Height*1000+b.AudioKbps))
	case ColumnProgress:
		return number(a.Percent, b.Percent)
	case ColumnStatus:
		x, _ := StatusWord(a)
		y, _ := StatusWord(b)

		return text(x, y)
	}

	return 0
}
