package main

import (
	"fmt"
	"time"
)

const maxLogLines = 400

// playlistView is what renderers know about one playlist.
type playlistView struct {
	Index   int
	URL     string
	Title   string
	Err     string
	Listed  bool
	Entries []int // entry ids in playlist order
	Log     []string
}

// entryView is what renderers know about one entry.
type entryView struct {
	EntryInfo
	State      EntryState
	TargetID   string
	Official   bool
	Err        string
	Percent    float64
	Downloaded int64
	Total      int64
	Speed      float64
	ETA        int
	Log        []string
}

// Tally holds the global counters shown in the header.
type Tally struct {
	Total, Queued, Running, Done, Official, Duplicate, Failed int
}

// runState is the renderer side mirror of the engine, built purely from
// events. Both the TUI and the plain renderer use it.
type runState struct {
	Playlists []*playlistView
	Entries   []*entryView
	Idle      bool
	Started   time.Time
}

func newRunState(sources []PlaylistSource) *runState {
	s := &runState{Started: time.Now()}
	for _, src := range sources {
		title := src.URL
		s.Playlists = append(s.Playlists, &playlistView{Index: src.Index, URL: src.URL, Title: title})
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

// apply folds one engine event into the state. It returns the id of the
// entry the event concerned, or -1.
func (s *runState) apply(ev interface{}) int {
	switch e := ev.(type) {
	case EvPlaylistListed:
		pl := s.Playlists[e.Playlist]
		pl.Title = e.Title
		pl.Listed = true
		for _, info := range e.Entries {
			for len(s.Entries) <= info.ID {
				s.Entries = append(s.Entries, nil)
			}
			s.Entries[info.ID] = &entryView{EntryInfo: info, State: StateQueued, TargetID: info.VideoID, ETA: -1}
			pl.Entries = append(pl.Entries, info.ID)
		}
	case EvPlaylistFailed:
		pl := s.Playlists[e.Playlist]
		pl.Listed = true
		pl.Err = e.Err
	case EvEntryState:
		en := s.entry(e.Entry)
		if en == nil {
			return -1
		}
		en.State = e.State
		en.TargetID = e.TargetID
		en.Official = e.Official
		en.Err = e.Err
		if e.State == StateQueued {
			en.Percent, en.Downloaded, en.Total, en.Speed, en.ETA = 0, 0, 0, 0, -1
		}
		if e.State == StateDone {
			en.Percent = 100
		}
		return e.Entry
	case EvProgress:
		en := s.entry(e.Entry)
		if en == nil {
			return -1
		}
		en.Percent, en.Downloaded, en.Total, en.Speed, en.ETA = e.Percent, e.Downloaded, e.Total, e.Speed, e.ETA
		return e.Entry
	case EvLog:
		if e.Playlist >= 0 && e.Playlist < len(s.Playlists) {
			pl := s.Playlists[e.Playlist]
			line := e.Line
			if en := s.entry(e.Entry); en != nil {
				line = fmt.Sprintf("[%02d] %s", en.Index, e.Line)
				en.Log = appendBounded(en.Log, e.Line)
			}
			pl.Log = appendBounded(pl.Log, line)
		}
		return e.Entry
	case EvIdle:
		s.Idle = true
	}
	return -1
}

func (s *runState) entry(id int) *entryView {
	if id < 0 || id >= len(s.Entries) {
		return nil
	}
	return s.Entries[id]
}

func (s *runState) tally() Tally {
	var t Tally
	for _, en := range s.Entries {
		if en == nil {
			continue
		}
		t.Total++
		switch en.State {
		case StateQueued:
			t.Queued++
		case StateResolving, StateDownloading, StateMerging:
			t.Running++
		case StateDone:
			t.Done++
			if en.Official {
				t.Official++
			}
		case StateDuplicate:
			t.Duplicate++
		case StateFailed:
			t.Failed++
		}
	}
	return t
}

// playlistTally counts the entries of one playlist.
func (s *runState) playlistTally(pl *playlistView) (finished, failed, active int) {
	for _, id := range pl.Entries {
		en := s.Entries[id]
		switch {
		case en.State == StateFailed:
			failed++
			finished++
		case en.State.finished():
			finished++
		case en.State != StateQueued:
			active++
		}
	}
	return
}

// --- formatting helpers shared by renderers --------------------------------

func humanBytes(b int64) string {
	const unit = 1024.0
	f := float64(b)
	for _, suffix := range []string{"B", "KiB", "MiB", "GiB"} {
		if f < unit || suffix == "GiB" {
			if suffix == "B" {
				return fmt.Sprintf("%d%s", b, suffix)
			}
			return fmt.Sprintf("%.1f%s", f, suffix)
		}
		f /= unit
	}
	return fmt.Sprintf("%.1fGiB", f)
}

func humanSpeed(bps float64) string {
	if bps <= 0 {
		return "--"
	}
	return humanBytes(int64(bps)) + "/s"
}

func humanETA(sec int) string {
	if sec < 0 {
		return "--"
	}
	if sec >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", sec/3600, (sec%3600)/60, sec%60)
	}
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}

func humanDuration(d time.Duration) string {
	sec := int(d.Seconds())
	return fmt.Sprintf("%02d:%02d:%02d", sec/3600, (sec%3600)/60, sec%60)
}

// progressLine renders "34.2% of 112.4MiB at 8.1MiB/s ETA 0:09".
func progressLine(en *entryView) string {
	if en.Total > 0 {
		return fmt.Sprintf("%5.1f%% of %s at %s ETA %s", en.Percent, humanBytes(en.Total), humanSpeed(en.Speed), humanETA(en.ETA))
	}
	return fmt.Sprintf("%s downloaded at %s", humanBytes(en.Downloaded), humanSpeed(en.Speed))
}
