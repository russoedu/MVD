// Package runstate mirrors engine events into a state that renderers can
// draw: playlists, entries, their progress and bounded logs.
package runstate

import (
	"fmt"
	"time"

	"youtube-downloader/libs/mvd-core/engine"
)

const maxLogLines = 400

// Playlist is what renderers know about one playlist.
type Playlist struct {
	Index   int
	URL     string
	Title   string
	Err     string
	Listed  bool
	Entries []int // entry ids in playlist order
	Log     []string
}

// Entry is what renderers know about one entry.
type Entry struct {
	engine.EntryInfo
	State      engine.EntryState
	TargetID   string
	Official   bool
	Better     bool
	Err        string
	Percent    float64
	Downloaded int64
	Total      int64
	Speed      float64
	ETA        int
	// Parts is how many parts of the file were being fetched at once, at the last tick.
	Parts int
	Log   []string
}

// Tally holds the global counters shown in the header.
type Tally struct {
	Total, Queued, Running, Done, Official, Better, Duplicate, NotFound, Failed, Retried int
	// Downloading is how many files are coming in now, Parts how many parts of them
	// at once, and Speed how fast, all of them together, in bytes per second.
	Downloading, Parts int
	Speed              float64
}

// State is the renderer side mirror of the engine, built purely from
// events. Both the TUI and the plain renderer use it.
type State struct {
	Playlists []*Playlist
	Entries   []*Entry
	Idle      bool
	Started   time.Time
	retried   int // cumulative failed -> queued transitions
}

// New prepares a state with one playlist per source.
func New(sources []engine.PlaylistSource) *State {
	s := &State{Started: time.Now()}
	for _, src := range sources {
		s.Playlists = append(s.Playlists, &Playlist{Index: src.Index, URL: src.URL, Title: src.URL})
	}
	return s
}

func appendBounded(lines []string, line string) []string {
	lines = append(lines, line)
	if len(lines) > maxLogLines {
		lines = lines[len(lines)-maxLogLines:]
	}
	return lines
}

// Apply folds one engine event into the state. It returns the id of the
// entry the event concerned, or -1.
func (s *State) Apply(ev interface{}) int {
	switch e := ev.(type) {
	case engine.EvPlaylistAdded:
		// New knows the sources the engine was built with; later ones arrive here.
		// The engine gives each the next free index, so anything else is an event
		// already folded in, and is ignored rather than duplicated.
		if e.Source.Index == len(s.Playlists) {
			s.Playlists = append(s.Playlists, &Playlist{Index: e.Source.Index, URL: e.Source.URL, Title: e.Source.URL})
		}
		// There is work again, so a finished run is not finished any more.
		s.Idle = false
	case engine.EvPlaylistListed:
		pl := s.Playlists[e.Playlist]
		pl.Title = e.Title
		pl.Listed = true
		for _, info := range e.Entries {
			for len(s.Entries) <= info.ID {
				s.Entries = append(s.Entries, nil)
			}
			s.Entries[info.ID] = &Entry{EntryInfo: info, State: engine.StateQueued, TargetID: info.VideoID, ETA: -1}
			pl.Entries = append(pl.Entries, info.ID)
		}
	case engine.EvPlaylistFailed:
		pl := s.Playlists[e.Playlist]
		pl.Listed = true
		pl.Err = e.Err
	case engine.EvEntryState:
		en := s.Entry(e.Entry)
		if en == nil {
			return -1
		}
		if en.State == engine.StateFailed && e.State == engine.StateQueued {
			s.retried++
		}
		en.State = e.State
		en.TargetID = e.TargetID
		en.Official = e.Official
		en.Better = e.Better
		en.Err = e.Err
		if e.State == engine.StateQueued {
			en.Percent, en.Downloaded, en.Total, en.Speed, en.ETA = 0, 0, 0, 0, -1
		}
		if e.State == engine.StateDone {
			en.Percent = 100
		}
		return e.Entry
	case engine.EvProgress:
		en := s.Entry(e.Entry)
		if en == nil {
			return -1
		}
		en.Percent, en.Downloaded, en.Total, en.Speed, en.ETA = e.Percent, e.Downloaded, e.Total, e.Speed, e.ETA
		en.Parts = e.Parts
		return e.Entry
	case engine.EvLog:
		if e.Playlist >= 0 && e.Playlist < len(s.Playlists) {
			pl := s.Playlists[e.Playlist]
			line := e.Line
			if en := s.Entry(e.Entry); en != nil {
				line = fmt.Sprintf("[%02d] %s", en.Index, e.Line)
				en.Log = appendBounded(en.Log, e.Line)
			}
			pl.Log = appendBounded(pl.Log, line)
		}
		return e.Entry
	case engine.EvIdle:
		s.Idle = true
	}
	return -1
}

// Entry returns the entry with the given id, or nil.
func (s *State) Entry(id int) *Entry {
	if id < 0 || id >= len(s.Entries) {
		return nil
	}
	return s.Entries[id]
}

// Tally counts every entry by state.
func (s *State) Tally() Tally {
	var t Tally
	for _, en := range s.Entries {
		if en == nil {
			continue
		}
		t.Total++
		switch en.State {
		case engine.StateQueued:
			t.Queued++
		case engine.StateResolving, engine.StateDownloading, engine.StateMerging:
			t.Running++
			if en.State == engine.StateDownloading {
				t.Downloading++
				t.Parts += max(1, en.Parts)
				t.Speed += en.Speed
			}
		case engine.StateDone:
			t.Done++
			if en.Official {
				t.Official++
			}
			if en.Better {
				t.Better++
			}
		case engine.StateDuplicate:
			t.Duplicate++
		case engine.StateNotFound:
			t.NotFound++
		case engine.StateFailed:
			t.Failed++
		}
	}
	t.Retried = s.retried
	return t
}

// PlaylistTally counts the entries of one playlist.
func (s *State) PlaylistTally(pl *Playlist) (finished, failed, active int) {
	for _, id := range pl.Entries {
		en := s.Entries[id]
		switch {
		case en.State == engine.StateFailed:
			failed++
			finished++
		case en.State.Finished():
			finished++
		case en.State != engine.StateQueued:
			active++
		}
	}
	return
}
