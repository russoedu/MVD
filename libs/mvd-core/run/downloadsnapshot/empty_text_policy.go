package downloadsnapshot

import "youtube-downloader/libs/mvd-core/run/downloadsession"

// EmptyEntriesText is what the pane of the songs says while it has none to show: that there is
// nothing yet, or that a list has been made and not looked up, and how to do it.
func EmptyEntriesText(s downloadsession.Session) string {
	switch {
	case len(s.State.Playlists) == 0:
		return "The songs of each source show here, once you add one."
	case s.Editable:
		return "Nothing loaded yet. Press Load playlists (Ctrl+R) to see the songs of each playlist, then choose what to download."
	case s.SelPlaylist < len(s.State.Playlists) && !s.State.Playlists[s.SelPlaylist].Listed:
		return "Reading the songs of this link…"
	}

	return "No songs to show here."
}

// EmptyOutputText is what the pane of the details says when nothing was written for what is chosen:
// before the run it says what will show; for a link that could not be read, why there is none;
// otherwise that this one wrote nothing.
func EmptyOutputText(s downloadsession.Session) string {
	switch {
	case s.SelPlaylist < len(s.State.Playlists) && s.State.Playlists[s.SelPlaylist].Err != "":
		return "Nothing to show: this link could not be read."
	case s.Editable || len(s.State.Playlists) == 0:
		return "Nothing yet. What happens to each song shows here while it downloads."
	}

	return "Nothing to show for this one."
}
