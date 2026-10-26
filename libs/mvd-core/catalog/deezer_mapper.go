package catalog

import "strconv"

type deezerArtist struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type deezerAlbum struct {
	ID          int64        `json:"id"`
	Title       string       `json:"title"`
	ReleaseDate string       `json:"release_date"`
	Artist      deezerArtist `json:"artist"`
}

type deezerTrack struct {
	Title    string       `json:"title"`
	Duration int          `json:"duration"` // seconds
	Position int          `json:"track_position"`
	Disc     int          `json:"disk_number"`
	Artist   deezerArtist `json:"artist"`
	Album    deezerAlbum  `json:"album"`
}

// deezerList is the shape of every answer of Deezer: a list under "data".
type deezerList[T any] struct {
	Data  []T `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (l deezerList[T]) problem() string {
	if l.Error != nil {
		return l.Error.Message
	}

	return ""
}

func artistsOf(rows []deezerArtist) []Artist {
	var out []Artist
	for _, r := range rows {
		if r.ID != 0 && r.Name != "" {
			out = append(out, Artist{ID: strconv.FormatInt(r.ID, 10), Name: r.Name})
		}
	}

	return out
}

// albumsOf maps albums; artist is the name to give those that carry none of their own (the albums of an
// artist do not say whose they are).
func albumsOf(rows []deezerAlbum, artist string) []Album {
	var out []Album
	for _, r := range rows {
		if r.ID == 0 || r.Title == "" {
			continue
		}
		owner := r.Artist.Name
		if owner == "" {
			owner = artist
		}
		out = append(out, Album{ID: strconv.FormatInt(r.ID, 10), Artist: owner, Title: r.Title, Released: r.ReleaseDate})
	}

	return out
}

// tracksOf maps songs; album names the album of those that carry none (the songs of an album do not).
func tracksOf(rows []deezerTrack, album string) []Track {
	var out []Track
	for _, r := range rows {
		if r.Title == "" {
			continue
		}
		title := r.Album.Title
		if title == "" {
			title = album
		}
		out = append(out, Track{Artist: r.Artist.Name, Album: title, Title: r.Title, DurationMs: r.Duration * 1000, Disc: r.Disc, Number: r.Position})
	}

	return out
}
