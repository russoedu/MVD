package runner

import (
	"context"

	"youtube-downloader/libs/mvd-core/applemusic"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/official"
	"youtube-downloader/libs/mvd-core/spotify"
)

// playlistProvider lists the public playlists of one music service.
type playlistProvider interface {
	Handles(link string) bool
	Tracks(ctx context.Context, link string) (string, []engine.Track, error)
}

// playlistTrackSource lists the public playlists of the music services the app
// knows (Spotify, Apple Music) and finds each song on YouTube, official video
// first and the best other upload when there is none.
type playlistTrackSource struct {
	providers []playlistProvider
	search    official.Searcher
	sources   official.Sources
}

func newPlaylistTrackSource(search official.Searcher, sources official.Sources) playlistTrackSource {
	return playlistTrackSource{
		providers: []playlistProvider{
			spotifyProvider{client: spotify.NewClient()},
			appleMusicProvider{client: applemusic.NewClient()},
		},
		search:  search,
		sources: sources,
	}
}

func (s playlistTrackSource) Handles(link string) bool {
	_, ok := s.providerFor(link)
	return ok
}

func (s playlistTrackSource) Tracks(ctx context.Context, link string) (string, []engine.Track, error) {
	provider, ok := s.providerFor(link)
	if !ok {
		return "", nil, nil
	}
	return provider.Tracks(ctx, link)
}

func (s playlistTrackSource) Find(_ context.Context, track engine.Track, logf func(format string, a ...interface{})) (string, bool, error) {
	found, err := official.FindTrack(s.search, s.sources, track.Title, track.Artist, track.DurationMs/1000, logf)
	return found.VideoID, found.Official, err
}

func (s playlistTrackSource) providerFor(link string) (playlistProvider, bool) {
	for _, provider := range s.providers {
		if provider.Handles(link) {
			return provider, true
		}
	}
	return nil, false
}

type spotifyProvider struct{ client *spotify.Client }

func (spotifyProvider) Handles(link string) bool {
	_, ok := spotify.PlaylistID(link)
	return ok
}

func (p spotifyProvider) Tracks(ctx context.Context, link string) (string, []engine.Track, error) {
	playlist, err := p.client.Fetch(ctx, link)
	if err != nil {
		return "", nil, err
	}
	tracks := make([]engine.Track, len(playlist.Tracks))
	for i, t := range playlist.Tracks {
		tracks[i] = engine.Track{Title: t.Title, Artist: t.Artist, DurationMs: t.DurationMs}
	}
	return playlist.Title, tracks, nil
}

type appleMusicProvider struct{ client *applemusic.Client }

func (appleMusicProvider) Handles(link string) bool {
	_, ok := applemusic.PlaylistURL(link)
	return ok
}

func (p appleMusicProvider) Tracks(ctx context.Context, link string) (string, []engine.Track, error) {
	playlist, err := p.client.Fetch(ctx, link)
	if err != nil {
		return "", nil, err
	}
	tracks := make([]engine.Track, len(playlist.Tracks))
	for i, t := range playlist.Tracks {
		tracks[i] = engine.Track{Title: t.Title, Artist: t.Artist, DurationMs: t.DurationMs}
	}
	return playlist.Title, tracks, nil
}
