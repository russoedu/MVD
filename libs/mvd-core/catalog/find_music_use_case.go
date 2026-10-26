package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"youtube-downloader/libs/mvd-core/musicsearch"
	"youtube-downloader/libs/mvd-core/official/titlematch"
)

// maxAlbums is the most albums of one artist that are read, so a long career is asked about politely.
const maxAlbums = 60

// albumsAtOnce is how many albums are read at the same time.
const albumsAtOnce = 3

// Result is the music a query names.
type Result struct {
	// Title is the query as a person reads it, for the name of the list of songs.
	Title  string
	Tracks []Track
}

// Finder names the songs of a query with the sources it has, one after the other: the first that knows
// the music answers.
type Finder struct {
	Sources []Source
}

// NewFinder is a finder with the sources, best first.
func NewFinder(sources ...Source) Finder { return Finder{Sources: sources} }

// DefaultFinder asks iTunes, and Deezer when iTunes does not know or cannot answer.
func DefaultFinder() Finder { return NewFinder(NewITunes(), NewDeezer()) }

// Find names the songs: every song of the artist, the songs of the album, or the one song. A song no
// catalogue knows is still the song that was typed, for the search of its video to try; an artist or an
// album no catalogue knows is an error.
func (f Finder) Find(ctx context.Context, q musicsearch.Query) (Result, error) {
	q = q.Tidy()
	if problem := q.Problem(); problem != "" {
		return Result{}, errors.New(problem)
	}
	var ask func(context.Context, Source) ([]Track, error)
	switch q.Kind() {
	case musicsearch.Artist:
		ask = func(ctx context.Context, s Source) ([]Track, error) { return artistTracks(ctx, s, q) }
	case musicsearch.Album:
		ask = func(ctx context.Context, s Source) ([]Track, error) { return albumTracks(ctx, s, q) }
	default:
		ask = func(ctx context.Context, s Source) ([]Track, error) { return songTrack(ctx, s, q) }
	}

	var failures []string
	for _, source := range f.Sources {
		tracks, err := ask(ctx, source)
		if ctx.Err() != nil {
			return Result{}, ctx.Err()
		}
		if len(tracks) > 0 {
			return Result{Title: q.Label(), Tracks: tracks}, nil
		}
		if err != nil {
			failures = append(failures, source.Name()+": "+err.Error())
		}
	}

	if q.Kind() == musicsearch.Song {
		return Result{Title: q.Label(), Tracks: []Track{{Artist: q.Artist, Album: q.Album, Title: q.Song}}}, nil
	}
	if len(failures) > 0 && len(failures) == len(f.Sources) {
		return Result{}, fmt.Errorf("cannot look up %s: %s", q.Label(), strings.Join(failures, "; "))
	}

	return Result{}, fmt.Errorf("no songs found for %s", q.Label())
}

// artistTracks is every song of the artist the source knows, the music itself before its versions, each
// song once.
func artistTracks(ctx context.Context, source Source, q musicsearch.Query) ([]Track, error) {
	artists, err := source.Artists(ctx, q.Artist)
	artist, found := bestArtist(q.Artist, artists)
	if !found {
		return nil, err
	}
	albums, err := source.Albums(ctx, artist)
	if err != nil {
		return nil, err
	}
	var own []Album
	for _, a := range albums {
		if (a.Artist == "" || artistIsCredited(q.Artist, a.Artist)) && !isUnwanted(a.Title) {
			own = append(own, a)
		}
	}
	own = inReleaseOrder(own)
	if len(own) > maxAlbums {
		own = own[:maxAlbums]
	}

	lists := make([][]Track, len(own))
	failures := make([]error, len(own))
	slots := make(chan struct{}, albumsAtOnce)
	var wg sync.WaitGroup
	for i, album := range own {
		wg.Add(1)
		slots <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-slots }()
			lists[i], failures[i] = source.AlbumTracks(ctx, album)
		}()
	}
	wg.Wait()

	var all []Track
	var lastFailure error
	for i, list := range lists {
		if failures[i] != nil {
			lastFailure = failures[i]
		}
		for _, t := range inAlbumOrder(list) {
			if artistIsCredited(q.Artist, t.Artist) && !isUnwanted(t.Title) && !isUnwanted(t.Album) {
				all = append(all, t)
			}
		}
	}
	all = uniqueTracks(all)
	if len(all) == 0 {
		return nil, lastFailure
	}

	return all, nil
}

// bestArtist is the artist of the catalogue that is the one asked for: the same name first, else the
// first that is close to it.
func bestArtist(asked string, artists []Artist) (Artist, bool) {
	for _, a := range artists {
		if titlematch.CompactName(a.Name) == titlematch.CompactName(asked) {
			return a, true
		}
	}
	for _, a := range artists {
		if artistIsNamed(asked, a.Name) {
			return a, true
		}
	}

	return Artist{}, false
}

// albumTracks is the songs of the album of the artist, in the order of the album.
func albumTracks(ctx context.Context, source Source, q musicsearch.Query) ([]Track, error) {
	albums, err := source.SearchAlbums(ctx, q.Artist, q.Album)
	best, bestScore := -1, 0.0
	for i, a := range albums {
		if !artistIsCredited(q.Artist, a.Artist) || isUnwanted(a.Title) {
			continue
		}
		score := songScore(q.Album, a.Title, q.Artist)
		if score < minAlbumScore {
			continue
		}
		// A version of the album loses to the album itself that scores as well.
		score -= 0.01 * float64(versionRank(a.Title))
		if best < 0 || score > bestScore {
			best, bestScore = i, score
		}
	}
	if best < 0 {
		return nil, err
	}
	tracks, err := source.AlbumTracks(ctx, albums[best])
	if err != nil {
		return nil, err
	}
	out := inAlbumOrder(tracks)
	for i := range out {
		if out[i].Artist == "" {
			out[i].Artist = albums[best].Artist
		}
		if out[i].Album == "" {
			out[i].Album = albums[best].Title
		}
	}

	return out, nil
}

// songTrack is the one song of the artist (on the album, when one was named) that the source knows.
func songTrack(ctx context.Context, source Source, q musicsearch.Query) ([]Track, error) {
	songs, err := source.SearchSongs(ctx, q.Artist, q.Song)
	best, bestScore := -1, 0.0
	for i, t := range songs {
		if !artistIsCredited(q.Artist, t.Artist) || isUnwanted(t.Title) || isUnwanted(t.Album) {
			continue
		}
		score := songScore(q.Song, t.Title, q.Artist)
		if score < minSongScore {
			continue
		}
		if q.Album != "" && songScore(q.Album, t.Album, q.Artist) >= minAlbumScore {
			score += 0.05
		}
		score -= 0.02 * float64(versionRank(t.Title)+versionRank(t.Album))
		if best < 0 || score > bestScore {
			best, bestScore = i, score
		}
	}
	if best < 0 {
		return nil, err
	}

	return []Track{songs[best]}, nil
}
