package engine

import (
	"context"
	"fmt"

	"youtube-downloader/libs/mvd-core/ytdlp"
)

// listEntries lists a source: through the track source when it handles the
// link, else through yt-dlp. For a track source the returned tracks run
// parallel to the entries; for yt-dlp they are nil.
func (e *Engine) listEntries(ctx context.Context, src PlaylistSource) ([]ytdlp.PlaylistEntry, []Track, error) {
	if e.opts.Tracks == nil || !e.opts.Tracks.Handles(src.URL) {
		entries, err := ytdlp.ListPlaylist(ctx, e.opts.YtDlp, src.URL, e.opts.ExtraArgs)
		return entries, nil, err
	}

	title, tracks, err := e.opts.Tracks.Tracks(ctx, src.URL)
	if err != nil {
		return nil, nil, err
	}

	entries := make([]ytdlp.PlaylistEntry, len(tracks))
	for i, t := range tracks {
		entries[i] = ytdlp.PlaylistEntry{
			Title:         t.Title,
			Channel:       t.Artist,
			Uploader:      t.Artist,
			Playlist:      title,
			PlaylistTitle: title,
			PlaylistIndex: i + 1,
			PlaylistCount: len(tracks),
		}
	}
	return entries, tracks, nil
}

// findTrack looks up the YouTube video of a track entry. It returns false,
// after failing the entry, when there is none.
func (e *Engine) findTrack(ctx context.Context, en *engineEntry) bool {
	pl, eid := en.info.Playlist, en.info.ID

	e.setState(en, StateResolving, "")
	id, official, err := e.opts.Tracks.Find(ctx, *en.track, func(format string, a ...interface{}) {
		e.log(pl, eid, "%s", fmt.Sprintf(format, a...))
	})
	switch {
	case ctx.Err() != nil:
		e.setState(en, StateFailed, "cancelled")
		return false
	case err != nil:
		e.log(pl, eid, "%v", err)
		e.mu.Lock()
		en.deferred = true // a search that could not run is worth retrying
		e.mu.Unlock()
		e.setState(en, StateFailed, err.Error())
		return false
	case id == "" && e.opts.OfficialOnly:
		// No video at all is no official video either.
		e.markNotFound(en)
		return false
	case id == "":
		e.log(pl, eid, "no video found for %s - %s", en.track.Artist, en.track.Title)
		e.setState(en, StateFailed, "no matching video found on YouTube")
		return false
	}

	e.mu.Lock()
	en.targetID = id
	en.official = official
	e.mu.Unlock()
	return true
}
