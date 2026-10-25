package downloadlist

import (
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/runstate"
)

// Preview is the list as the download screen shows it when nothing is running: a
// playlist for each item, a song for each of its songs in the state the file has kept.
// An item nobody has looked at yet is a playlist with no songs.
func (s *Session) Preview() *runstate.State {
	s.mu.Lock()
	file := copyOf(s.file)
	s.mu.Unlock()
	return PreviewOf(file)
}

// PreviewOf is Preview of a list that is not a session.
func PreviewOf(file playlistfile.File) *runstate.State {
	sources := make([]engine.PlaylistSource, len(file.Items))
	for i, item := range file.Items {
		sources[i] = engine.PlaylistSource{Index: i, URL: item.Source}
	}
	state := runstate.New(sources)

	id := 0
	for i, item := range file.Items {
		if len(item.Songs) == 0 {
			// Nobody has looked at it yet: it is shown as it is, not as being listed.
			state.Apply(engine.EvPlaylistListed{Playlist: i, Title: item.Source})
			continue
		}
		infos := make([]engine.EntryInfo, len(item.Songs))
		for j, song := range item.Songs {
			index := song.Upload.PlaylistIndex
			if index == 0 {
				index = j + 1
			}
			infos[j] = engine.EntryInfo{ID: id + j, Playlist: i, Index: index, VideoID: song.Upload.ID, Title: song.Upload.Title, Channel: song.Upload.Channel}
		}
		state.Apply(engine.EvPlaylistListed{Playlist: i, Title: titleOf(item), Entries: infos})
		for j, song := range item.Songs {
			state.Apply(entryStateOf(infos[j].ID, song))
		}
		id += len(item.Songs)
	}
	return state
}

// titleOf is what an item is called on the screen: the playlist's own title once it has
// been looked at, else the address.
func titleOf(item playlistfile.Item) string {
	for _, song := range item.Songs {
		if song.Playlist != "" {
			return song.Playlist
		}
	}
	return item.Source
}

// entryStateOf is the event that puts a song of the file in the state it was left in.
func entryStateOf(id int, song playlistfile.Entry) engine.EvEntryState {
	ev := engine.EvEntryState{
		Entry:    id,
		State:    engine.StateQueued,
		TargetID: song.TargetID,
		Official: song.Kind == string(engine.KindOfficial),
		Better:   song.Kind == string(engine.KindBetter),
		Reason:   song.Reason,
		Artist:   song.Artist,
		Title:    song.Title,
		Err:      song.Error,
	}
	ev.OwnOfficial = song.Kind == string(engine.KindOwnOfficial)
	switch {
	case song.Downloaded || song.State == playlistfile.StateDone:
		ev.State = engine.StateDone
	case song.Decision == playlistfile.DecisionSkipped:
		ev.State = engine.StateSkipped
	case song.State == playlistfile.StateFailed:
		ev.State = engine.StateFailed
	case song.State == playlistfile.StateNotFound:
		ev.State = engine.StateNotFound
	case song.State == playlistfile.StatePlanned:
		ev.State = engine.StatePlanned
	}
	return ev
}
