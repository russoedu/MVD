package catalog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func serving(t *testing.T, routes map[string]string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := routes[r.URL.Path+"?"+r.URL.Query().Get("entity")]
		if !ok {
			body, ok = routes[r.URL.Path]
		}
		if !ok {
			http.NotFound(w, r)

			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return server
}

func TestITunesNamesArtistsAlbumsAndSongs(t *testing.T) {
	server := serving(t, map[string]string{
		"/search?musicArtist": `{"results":[{"wrapperType":"artist","artistId":5468295,"artistName":"Daft Punk"}]}`,
		"/lookup?album": `{"results":[{"wrapperType":"artist","artistId":5468295,"artistName":"Daft Punk"},
			{"wrapperType":"collection","collectionId":697194953,"collectionName":"Discovery","artistName":"Daft Punk","releaseDate":"2001-03-13T08:00:00Z"}]}`,
		"/lookup?song": `{"results":[{"wrapperType":"collection","collectionId":697194953,"collectionName":"Discovery"},
			{"wrapperType":"track","kind":"song","trackName":"One More Time","artistName":"Daft Punk","collectionName":"Discovery","trackTimeMillis":320357,"trackNumber":2,"discNumber":1}]}`,
		"/search?song": `{"results":[{"wrapperType":"track","kind":"song","trackName":"Around the World","artistName":"Daft Punk","collectionName":"Homework","trackTimeMillis":429533}]}`,
	})
	source := &ITunes{HTTP: server.Client(), BaseURL: server.URL}
	ctx := context.Background()

	artists, err := source.Artists(ctx, "daft punk")
	if err != nil || len(artists) != 1 || artists[0].ID != "5468295" || artists[0].Name != "Daft Punk" {
		t.Fatalf("artists %+v, %v", artists, err)
	}
	albums, err := source.Albums(ctx, artists[0])
	if err != nil || len(albums) != 1 || albums[0].ID != "697194953" || albums[0].Released != "2001-03-13" {
		t.Fatalf("albums %+v, %v (the artist row is not an album)", albums, err)
	}
	tracks, err := source.AlbumTracks(ctx, albums[0])
	if err != nil || len(tracks) != 1 || tracks[0].Title != "One More Time" || tracks[0].DurationMs != 320357 || tracks[0].Number != 2 {
		t.Fatalf("tracks %+v, %v (the album row is not a song)", tracks, err)
	}
	songs, err := source.SearchSongs(ctx, "Daft Punk", "Around the World")
	if err != nil || len(songs) != 1 || songs[0].Album != "Homework" {
		t.Fatalf("songs %+v, %v", songs, err)
	}
}

func TestACatalogueThatAsksToWaitIsAskedAgain(t *testing.T) {
	retryDelay = time.Millisecond
	t.Cleanup(func() { retryDelay = 2 * time.Second })
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusForbidden)

			return
		}
		_, _ = w.Write([]byte(`{"results":[{"wrapperType":"artist","artistId":1,"artistName":"ATB"}]}`))
	}))
	defer server.Close()

	artists, err := (&ITunes{HTTP: server.Client(), BaseURL: server.URL}).Artists(context.Background(), "ATB")
	if err != nil || len(artists) != 1 || calls.Load() != 2 {
		t.Errorf("artists %+v, %v after %d calls", artists, err, calls.Load())
	}
}

func TestACatalogueThatFailsForGoodIsAnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
	defer server.Close()
	if _, err := (&ITunes{HTTP: server.Client(), BaseURL: server.URL}).Artists(context.Background(), "x"); err == nil {
		t.Error("a 500 was taken for an answer")
	}
}

func TestAnEndedContextStopsTheWait(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTooManyRequests) }))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (&ITunes{HTTP: server.Client(), BaseURL: server.URL}).Artists(ctx, "x"); err == nil {
		t.Error("an ended context was not noticed")
	}
}
