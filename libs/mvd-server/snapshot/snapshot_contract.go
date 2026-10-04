// Package snapshot is the JSON picture of a run that the web front end draws.
//
// It is a contract, not a model: the field names and shapes here are what the
// browser depends on, so they are written out with explicit tags instead of
// leaking the Go names of the engine's own types.
package snapshot

import "time"

// Snapshot is the whole state of the run at one moment. Version only ever
// increases, so a client that holds version N can ask for anything newer.
type Snapshot struct {
	Version   int64          `json:"version"`
	Idle      bool           `json:"idle"`
	Started   time.Time      `json:"started"`
	Tally     Tally          `json:"tally"`
	Playlists []PlaylistView `json:"playlists"`
	Entries   []EntryView    `json:"entries"`
}

// Tally holds the global counters.
type Tally struct {
	Total     int `json:"total"`
	Queued    int `json:"queued"`
	Running   int `json:"running"`
	Done      int `json:"done"`
	Official  int `json:"official"`
	Duplicate int `json:"duplicate"`
	Failed    int `json:"failed"`
	Retried   int `json:"retried"`
}

// PlaylistView is one playlist (or single video) and how far it has got.
type PlaylistView struct {
	Index    int    `json:"index"`
	URL      string `json:"url"`
	Title    string `json:"title"`
	Err      string `json:"err"`
	Listed   bool   `json:"listed"`
	Total    int    `json:"total"`
	Finished int    `json:"finished"`
	Failed   int    `json:"failed"`
	Active   int    `json:"active"`
	// Entries are the ids of its entries, in playlist order.
	Entries []int `json:"entries"`
}

// EntryView is one video. Logs are left out on purpose: they are large and
// only wanted for the entry being looked at.
type EntryView struct {
	ID       int    `json:"id"`
	Playlist int    `json:"playlist"`
	Index    int    `json:"index"`
	VideoID  string `json:"videoId"`
	// TargetID is the video actually downloaded: the official one when it replaced the art track.
	TargetID   string  `json:"targetId"`
	Title      string  `json:"title"`
	Channel    string  `json:"channel"`
	State      string  `json:"state"`
	Official   bool    `json:"official"`
	Err        string  `json:"err"`
	Percent    float64 `json:"percent"`
	Downloaded int64   `json:"downloaded"`
	TotalBytes int64   `json:"totalBytes"`
	// Speed is bytes per second, 0 when unknown.
	Speed float64 `json:"speed"`
	// ETA is seconds, -1 when unknown.
	ETA int `json:"eta"`
}
