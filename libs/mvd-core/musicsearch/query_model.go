package musicsearch

import (
	"net/url"
	"strings"
)

// Scheme begins the address a query is kept under in a download list, which holds addresses.
const Scheme = "mvdsearch:"

// Query is the music asked for by name.
type Query struct {
	Artist, Album, Song string
}

// Kind says what a query asks for.
type Kind int

// The kinds of query.
const (
	// Invalid is a query that asks for nothing, or for an album or a song with no artist.
	Invalid Kind = iota
	// Artist asks for every song of the artist.
	Artist
	// Album asks for the songs of one album of the artist.
	Album
	// Song asks for one song of the artist, on an album or not.
	Song
)

// Tidy is the query with the spaces around each part taken away.
func (q Query) Tidy() Query {
	return Query{Artist: strings.TrimSpace(q.Artist), Album: strings.TrimSpace(q.Album), Song: strings.TrimSpace(q.Song)}
}

// Kind is what the query asks for. A song is the most particular, then an album, then the artist.
func (q Query) Kind() Kind {
	q = q.Tidy()
	switch {
	case q.Artist == "":
		return Invalid
	case q.Song != "":
		return Song
	case q.Album != "":
		return Album
	}

	return Artist
}

// Problem is why the query cannot be searched, in words for the person, or "" when it can.
func (q Query) Problem() string {
	q = q.Tidy()
	switch {
	case q.Artist == "" && q.Album == "" && q.Song == "":
		return "Write an artist"
	case q.Artist == "" && q.Song != "":
		return "A song needs its artist"
	case q.Artist == "":
		return "An album needs its artist"
	}

	return ""
}

// Label is the query as a person reads it: "Daft Punk", "Daft Punk – Discovery", "Daft Punk – One
// More Time" or "Daft Punk – Discovery – One More Time".
func (q Query) Label() string {
	q = q.Tidy()
	parts := []string{q.Artist}
	if q.Album != "" {
		parts = append(parts, q.Album)
	}
	if q.Song != "" {
		parts = append(parts, q.Song)
	}

	return strings.Join(parts, " – ")
}

// Address is the text a download list keeps the query under.
func (q Query) Address() string {
	q = q.Tidy()
	v := url.Values{}
	for key, value := range map[string]string{"artist": q.Artist, "album": q.Album, "song": q.Song} {
		if value != "" {
			v.Set(key, value)
		}
	}

	return Scheme + v.Encode()
}

// Parse reads an address made by Address. ok is false for any other text, and for a query that cannot
// be searched.
func Parse(address string) (q Query, ok bool) {
	rest, found := strings.CutPrefix(strings.TrimSpace(address), Scheme)
	if !found {
		return Query{}, false
	}
	v, err := url.ParseQuery(rest)
	if err != nil {
		return Query{}, false
	}
	q = Query{Artist: v.Get("artist"), Album: v.Get("album"), Song: v.Get("song")}.Tidy()

	return q, q.Problem() == ""
}

// Is says whether address is made by Address, whatever it holds.
func Is(address string) bool { return strings.HasPrefix(strings.TrimSpace(address), Scheme) }
