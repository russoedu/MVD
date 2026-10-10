package engine

import "youtube-downloader/libs/mvd-core/ytdlp"

// PlanKind says what a plan proposes to download for a song.
type PlanKind string

const (
	// KindOfficial is the official music video found for an art track.
	KindOfficial PlanKind = "official"
	// KindOwnOfficial is an upload that is the official video itself.
	KindOwnOfficial PlanKind = "own-official"
	// KindBetter is the upload of the song with the best picture and sound.
	KindBetter PlanKind = "better"
	// KindOriginal is the upload from the playlist, kept as it is.
	KindOriginal PlanKind = "original"
	// KindNone is a song nothing was found for.
	KindNone PlanKind = "none"
	// KindChosen is a video the person picked in place of the proposal.
	KindChosen PlanKind = "chosen"
)

// PlannedEntry is one song of a plan: the upload that is in the playlist and the video
// to download for it. A plan-only run produces them, and a run built with Options.Plan
// downloads them.
type PlannedEntry struct {
	// Playlist is the title of the playlist and PlaylistURL its address; entries of one
	// playlist share both.
	Playlist    string
	PlaylistURL string
	// Upload is the entry as the playlist lists it; the file name is made from it.
	Upload ytdlp.PlaylistEntry
	// TargetID is the video to download, "" when there is none.
	TargetID string
	Kind     PlanKind
	// Reason says why the version was chosen, as the lookup put it.
	Reason string
	// Artist and Title are the song as a music database named it, when it did.
	Artist, Title string
}

// planSources gives each playlist of Options.Plan its source, in the order they appear,
// and builds the entries. The playlists are not listed: Run announces them.
func (e *Engine) planSources() {
	if len(e.opts.Plan) == 0 {
		return
	}
	index := map[string]int{}
	for _, p := range e.opts.Plan {
		key := p.PlaylistURL
		if key == "" {
			key = p.Playlist
		}
		pl, ok := index[key]
		if !ok {
			pl = len(e.sources)
			index[key] = pl
			url := p.PlaylistURL
			if url == "" {
				url = p.Playlist
			}
			e.sources = append(e.sources, PlaylistSource{Index: pl, URL: url})
			e.playlistTitles[pl] = p.Playlist
		}
		idx := p.Upload.PlaylistIndex
		if idx == 0 {
			idx = len(e.entries) + 1
		}
		target := p.TargetID
		en := &engineEntry{
			info:        EntryInfo{ID: len(e.entries), Playlist: pl, Index: idx, VideoID: p.Upload.ID, Title: p.Upload.Title, Channel: p.Upload.Channel},
			raw:         p.Upload,
			state:       StateQueued,
			targetID:    target,
			official:    p.Kind == KindOfficial,
			better:      p.Kind == KindBetter,
			ownOfficial: p.Kind == KindOwnOfficial,
			reason:      p.Reason,
			fixed:       true,
		}
		if p.Artist != "" && p.Title != "" {
			en.named = &namedSong{artist: p.Artist, title: p.Title}
		}
		e.entries = append(e.entries, en)
	}
}

// seedPlan announces the playlists of a plan and hands its entries to the download stage.
func (e *Engine) seedPlan() {
	if len(e.opts.Plan) == 0 {
		return
	}
	e.mu.Lock()
	byPlaylist := map[int][]EntryInfo{}
	var ids []int
	for _, en := range e.entries {
		byPlaylist[en.info.Playlist] = append(byPlaylist[en.info.Playlist], en.info)
		ids = append(ids, en.info.ID)
	}
	e.pending += len(ids)
	sources := append([]PlaylistSource(nil), e.sources...)
	e.mu.Unlock()

	for _, src := range sources {
		e.emit(EvPlaylistListed{Playlist: src.Index, Title: e.playlistTitles[src.Index], Entries: byPlaylist[src.Index]})
	}
	for _, id := range ids {
		e.queue.Push(id)
	}
}

// Plan returns what a plan-only run proposes for each song it has dealt with, in the
// order the songs were listed. Songs still waiting for a stage are left out.
func (e *Engine) Plan() []PlannedEntry {
	e.mu.Lock()
	defer e.mu.Unlock()
	var plan []PlannedEntry
	for _, en := range e.entries {
		switch en.state {
		case StatePlanned, StateNotFound, StateFailed:
		default:
			continue
		}
		p := PlannedEntry{
			Playlist: e.playlistTitles[en.info.Playlist],
			Upload:   en.raw,
			Reason:   en.reason,
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
		case en.state == StatePlanned && en.official:
			p.Kind, p.TargetID = KindOfficial, en.targetID
		case en.state == StatePlanned && en.better:
			p.Kind, p.TargetID = KindBetter, en.targetID
		case en.state == StatePlanned && en.ownOfficial:
			p.Kind, p.TargetID = KindOwnOfficial, en.info.VideoID
		case en.state == StatePlanned:
			p.Kind, p.TargetID = KindOriginal, en.targetID
		default:
			// Not found, or a track no video was found for.
			p.Kind = KindNone
			if en.track == nil {
				p.TargetID = en.info.VideoID
			}
		}
		if en.state == StateFailed && en.err != "" {
			p.Reason = en.err
		}
		plan = append(plan, p)
	}
	return plan
}
