package runner

import (
	"context"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/official"
	"youtube-downloader/libs/mvd-core/spotify"
)

// spotifyTrackSource lists public Spotify playlists and finds each song on
// YouTube, official video first and the best other upload when there is none.
type spotifyTrackSource struct {
	client *spotify.Client
	search official.Searcher
}

func (spotifyTrackSource) Handles(link string) bool {
	_, ok := spotify.PlaylistID(link)
	return ok
}

func (s spotifyTrackSource) Tracks(ctx context.Context, link string) (string, []engine.Track, error) {
	playlist, err := s.client.Fetch(ctx, link)
	if err != nil {
		return "", nil, err
	}
	tracks := make([]engine.Track, len(playlist.Tracks))
	for i, t := range playlist.Tracks {
		tracks[i] = engine.Track{Title: t.Title, Artist: t.Artist, DurationMs: t.DurationMs}
	}
	return playlist.Title, tracks, nil
}

func (s spotifyTrackSource) Find(_ context.Context, track engine.Track, logf func(format string, a ...interface{})) (string, bool, error) {
	found, err := official.FindTrack(s.search, track.Title, track.Artist, logf)
	return found.VideoID, found.Official, err
}
