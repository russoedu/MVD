package engine

// EntrySnapshot is one song of the run as it stands: what the plan says about it and how
// far it has got.
type EntrySnapshot struct {
	PlannedEntry
	State EntryState
	// Err is why the song failed, when it did.
	Err string
}

// Snapshot returns every song the run has listed, in the order they were listed, with the
// version each is to download and its state. Songs still waiting for the name or pick
// stage have no TargetID yet. The list file is written from it.
func (e *Engine) Snapshot() []EntrySnapshot {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]EntrySnapshot, 0, len(e.entries))
	for _, en := range e.entries {
		if e.removed[en.info.Playlist] {
			continue
		}
		p := PlannedEntry{
			Playlist: e.playlistTitles[en.info.Playlist],
			Upload:   en.raw,
			Reason:   en.reason,
			TargetID: en.targetID,
		}
		if en.info.Playlist < len(e.sources) {
			p.PlaylistURL = e.sources[en.info.Playlist].URL
		}
		if p.Upload.ID == "" {
			p.Upload.ID = en.info.VideoID
		}
		if en.named != nil {
			p.Artist, p.Title = en.named.artist, en.named.title
		}
		switch {
		case en.userChose:
			p.Kind = KindChosen
		case en.official:
			p.Kind = KindOfficial
		case en.better:
			p.Kind = KindBetter
		case en.ownOfficial:
			p.Kind, p.TargetID = KindOwnOfficial, en.info.VideoID
		case en.targetID == "":
			p.Kind = KindNone
		default:
			p.Kind = KindOriginal
		}
		out = append(out, EntrySnapshot{PlannedEntry: p, State: en.state, Err: en.err})
	}
	return out
}
