package snapshot

import "youtube-downloader/libs/mvd-core/runstate"

// From converts a run state into a snapshot at the given version.
//
// The caller must hold whatever lock guards the state: this only reads it. The
// slices are never nil, so they encode as [] and a client can iterate them
// without a nil check.
func From(state *runstate.State, version int64) Snapshot {
	tally := state.Tally()
	out := Snapshot{
		Version: version,
		Idle:    state.Idle,
		Started: state.Started,
		Tally: Tally{
			Total: tally.Total, Queued: tally.Queued, Running: tally.Running, Done: tally.Done,
			Official: tally.Official, Duplicate: tally.Duplicate, Failed: tally.Failed, Retried: tally.Retried,
		},
		Playlists: make([]PlaylistView, 0, len(state.Playlists)),
		Entries:   make([]EntryView, 0, len(state.Entries)),
	}

	for _, pl := range state.Playlists {
		finished, failed, active := state.PlaylistTally(pl)
		ids := append([]int{}, pl.Entries...)
		out.Playlists = append(out.Playlists, PlaylistView{
			Index: pl.Index, URL: pl.URL, Title: pl.Title, Err: pl.Err, Listed: pl.Listed,
			Total: len(pl.Entries), Finished: finished, Failed: failed, Active: active, Entries: ids,
		})
	}

	for _, en := range state.Entries {
		// An id is reserved before its entry arrives, so there can be a gap.
		if en == nil {
			continue
		}
		out.Entries = append(out.Entries, EntryView{
			ID: en.ID, Playlist: en.Playlist, Index: en.Index, VideoID: en.VideoID, TargetID: en.TargetID, Title: en.Title,
			Channel: en.Channel, State: en.State.String(), Official: en.Official, Err: en.Err,
			Percent: en.Percent, Downloaded: en.Downloaded, TotalBytes: en.Total, Speed: en.Speed, ETA: en.ETA,
		})
	}

	return out
}
