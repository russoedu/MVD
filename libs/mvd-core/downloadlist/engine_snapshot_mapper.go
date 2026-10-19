package downloadlist

import (
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlistfile"
)

// Apply brings the list up to date with what a run reports: new songs are added to the
// item they came from, and each song takes the proposal and the state the run has for it.
// What the person decided is never touched, and a song that was downloaded stays so.
func (s *Session) Apply(snapshot []engine.EntrySnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := make(map[string]int, len(s.file.Items))
	for i, it := range s.file.Items {
		item[it.Source] = i
	}
	for _, snap := range snapshot {
		fresh := playlistfile.Entry{
			Playlist:    snap.Playlist,
			PlaylistURL: snap.PlaylistURL,
			Upload:      snap.Upload,
			TargetID:    snap.TargetID,
			Kind:        string(snap.Kind),
			Reason:      snap.Reason,
			Artist:      snap.Artist,
			Title:       snap.Title,
		}
		fresh.State, fresh.Error = stateOf(snap)
		fresh.Downloaded = snap.State == engine.StateDone || snap.State == engine.StateDuplicate

		at, ok := item[snap.PlaylistURL]
		if !ok {
			// The run knows a playlist the list does not: a link added while it ran.
			at = len(s.file.Items)
			item[snap.PlaylistURL] = at
			s.file.Items = append(s.file.Items, playlistfile.Item{Source: snap.PlaylistURL})
		}
		songs := s.file.Items[at].Songs
		key := fresh.Key()
		found := false
		for i := range songs {
			if songs[i].Key() != key {
				continue
			}
			found = true
			keep := songs[i]
			keep.Playlist, keep.Upload, keep.Artist, keep.Title = fresh.Playlist, fresh.Upload, fresh.Artist, fresh.Title
			// A song the run has not got to yet knows less than the file does.
			if snap.TargetID != "" || keep.TargetID == "" {
				keep.TargetID, keep.Kind, keep.Reason = fresh.TargetID, fresh.Kind, fresh.Reason
			}
			keep.State, keep.Error = fresh.State, fresh.Error
			keep.Downloaded = keep.Downloaded || fresh.Downloaded
			songs[i] = keep
			break
		}
		if !found {
			songs = append(songs, fresh)
		}
		s.file.Items[at].Songs = songs
	}
}

// stateOf is the state and error to write in the file for a song of the run. A song that
// was cancelled when the run was closed has not failed: it simply goes on next time.
func stateOf(snap engine.EntrySnapshot) (state, err string) {
	switch snap.State {
	case engine.StatePlanned:
		return playlistfile.StatePlanned, ""
	case engine.StateDownloading, engine.StateMerging:
		return playlistfile.StateDownloading, ""
	case engine.StateDone, engine.StateDuplicate:
		return playlistfile.StateDone, ""
	case engine.StateNotFound:
		return playlistfile.StateNotFound, ""
	case engine.StateFailed:
		if snap.Err == "cancelled" {
			return playlistfile.StateWaiting, ""
		}
		return playlistfile.StateFailed, snap.Err
	}
	if snap.TargetID != "" {
		return playlistfile.StateQueued, ""
	}
	return playlistfile.StateWaiting, ""
}
