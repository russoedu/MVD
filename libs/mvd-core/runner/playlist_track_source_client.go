package runner

import (
	"context"
	"errors"
	"youtube-downloader/libs/mvd-core/official/trackfind"
	"youtube-downloader/libs/mvd-core/official/trackpick"

	"youtube-downloader/libs/mvd-core/applemusic"
	"youtube-downloader/libs/mvd-core/catalog"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/musicsearch"
	"youtube-downloader/libs/mvd-core/songfile"
	"youtube-downloader/libs/mvd-core/spotify"
)

// playlistProvider lists the public playlists of one music service.
type playlistProvider interface {
	Handles(link string) bool
	Tracks(ctx context.Context, link string) (string, []engine.Track, error)
}

// playlistTrackSource lists the public playlists of the music services the app
// knows (Spotify, Apple Music) and the song files of the user, and finds each
// song on YouTube, official video first and the best other upload when there
// is none.
type playlistTrackSource struct {
	providers []playlistProvider
	search    trackpick.Searcher
	sources   trackfind.Sources
}

func newPlaylistTrackSource(search trackpick.Searcher, sources trackfind.Sources) playlistTrackSource {
	return playlistTrackSource{
		providers: []playlistProvider{
			spotifyProvider{client: spotify.NewClient()},
			appleMusicProvider{client: applemusic.NewClient()},
			songFileProvider{},
			searchProvider{finder: catalog.DefaultFinder()},
		},
		search:  search,
		sources: sources,
	}
}

// Handles reports whether one of the providers can list the link.
func (s playlistTrackSource) Handles(link string) bool {
	_, ok := s.providerFor(link)
	return ok
}

// Tracks lists the title and the songs of the playlist at link, through the provider that handles it.
// It returns nothing when no provider handles the link, and the provider's error when listing fails.
func (s playlistTrackSource) Tracks(ctx context.Context, link string) (string, []engine.Track, error) {
	provider, ok := s.providerFor(link)
	if !ok {
		return "", nil, nil
	}
	return provider.Tracks(ctx, link)
}

// Find looks for the YouTube video of track and receives the progress lines in logf.
// It returns the video id and whether that video is official.
// It returns an error when the search fails.
func (s playlistTrackSource) Find(_ context.Context, track engine.Track, logf func(format string, a ...interface{})) (string, bool, error) {
	found, err := trackfind.FindTrack(s.search, s.sources, track.Title, track.Artist, track.DurationMs/1000, logf)
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

// Handles reports whether link is a Spotify playlist.
func (spotifyProvider) Handles(link string) bool {
	_, ok := spotify.PlaylistID(link)
	return ok
}

// Tracks fetches the Spotify playlist at link and returns its title and songs.
// It returns an error when the fetch fails.
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

// Handles reports whether link is an Apple Music playlist.
func (appleMusicProvider) Handles(link string) bool {
	_, ok := applemusic.PlaylistURL(link)
	return ok
}

// Tracks fetches the Apple Music playlist at link and returns its title and songs.
// It returns an error when the fetch fails.
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

// songFileProvider lists a file of songs on the user's computer: "Artist -
// Title" lines or a CSV, whatever service the list came from.
type songFileProvider struct{}

// Handles reports whether link points to a song file.
func (songFileProvider) Handles(link string) bool {
	_, ok := songfile.FilePath(link)
	return ok
}

// Tracks loads the song file at link and returns its title and songs.
// It returns an error when the file cannot be loaded.
func (songFileProvider) Tracks(_ context.Context, link string) (string, []engine.Track, error) {
	path, _ := songfile.FilePath(link)
	list, err := songfile.Load(path)
	if err != nil {
		return "", nil, err
	}
	tracks := make([]engine.Track, len(list.Songs))
	for i, s := range list.Songs {
		tracks[i] = engine.Track{Title: s.Title, Artist: s.Artist, DurationMs: s.DurationMs}
	}
	return list.Title, tracks, nil
}

// searchProvider lists the music a person asked for by name (see package musicsearch): the songs of an
// artist, of an album, or the one song, as the catalogues name them. The video of each is found like
// that of any other song.
type searchProvider struct{ finder catalog.Finder }

// Handles reports whether link is a search by name.
func (searchProvider) Handles(link string) bool { return musicsearch.Is(link) }

// Tracks finds the songs the search at link asks for and returns the title and the songs.
// It returns an error when the search is not for an artist, an album or a song, or when the catalogues fail.
func (p searchProvider) Tracks(ctx context.Context, link string) (string, []engine.Track, error) {
	query, ok := musicsearch.Parse(link)
	if !ok {
		return "", nil, errors.New("this search is not an artist, an artist and an album, or an artist and a song")
	}
	found, err := p.finder.Find(ctx, query)
	if err != nil {
		return "", nil, err
	}
	tracks := make([]engine.Track, len(found.Tracks))
	for i, t := range found.Tracks {
		tracks[i] = engine.Track{Title: t.Title, Artist: t.Artist, DurationMs: t.DurationMs}
	}

	return found.Title, tracks, nil
}
