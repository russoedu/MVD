package engine

// PlaylistSource is one URL from downloads.conf.
type PlaylistSource struct {
	Index int
	URL   string
}

// EntryInfo is the immutable description of an entry handed to renderers.
type EntryInfo struct {
	ID       int // global, sequential
	Playlist int // index into the playlist list
	Index    int // position inside the playlist (1 based)
	VideoID  string
	Title    string
	Channel  string
}

// Events emitted by the Engine. Renderers (TUI, plain log) consume them.
type (
	// EvPlaylistAdded is sent when AddSource queues a playlist after the engine
	// was built, before its listing starts. Playlists the engine was built with
	// are known from Sources() and send no such event.
	EvPlaylistAdded struct {
		Source PlaylistSource
	}
	// EvPlaylistListing is sent when the listing of a playlist starts.
	EvPlaylistListing struct {
		Playlist int
		URL      string
	}
	// EvPlaylistListed is sent once the flat listing of a playlist is known.
	EvPlaylistListed struct {
		Playlist int
		Title    string
		Entries  []EntryInfo
	}
	// EvPlaylistFailed is sent when a playlist could not be listed.
	EvPlaylistFailed struct {
		Playlist int
		Err      string
	}
	// EvEntryState is sent on every state transition of an entry.
	EvEntryState struct {
		Entry    int
		State    EntryState
		TargetID string // video actually downloaded (official or original)
		Official bool   // true when TargetID is the official video
		Err      string // set for StateFailed
	}
	// EvProgress carries a download progress tick.
	EvProgress struct {
		Entry      int
		Percent    float64
		Downloaded int64
		Total      int64
		Speed      float64 // bytes per second, 0 when unknown
		ETA        int     // seconds, -1 when unknown
	}
	// EvLog is one line of output. Entry is -1 for playlist level lines.
	EvLog struct {
		Playlist int
		Entry    int
		Line     string
	}
	// EvIdle is sent when the queue is empty and no worker is busy.
	EvIdle struct{}
)
