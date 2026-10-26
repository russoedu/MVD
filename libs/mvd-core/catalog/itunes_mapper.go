package catalog

import "strconv"

// itunesAnswer is what iTunes answers to a search or a lookup: rows of several kinds in one list.
type itunesAnswer struct {
	Results []itunesRow `json:"results"`
}

type itunesRow struct {
	WrapperType    string `json:"wrapperType"`
	ArtistID       int64  `json:"artistId"`
	ArtistName     string `json:"artistName"`
	CollectionID   int64  `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	TrackName      string `json:"trackName"`
	ReleaseDate    string `json:"releaseDate"`
	TrackTimeMs    int    `json:"trackTimeMillis"`
	TrackNumber    int    `json:"trackNumber"`
	DiscNumber     int    `json:"discNumber"`
	Kind           string `json:"kind"`
}

func (a itunesAnswer) artists() []Artist {
	var out []Artist
	for _, r := range a.Results {
		if r.ArtistID != 0 && r.ArtistName != "" && (r.WrapperType == "artist" || r.WrapperType == "") {
			out = append(out, Artist{ID: strconv.FormatInt(r.ArtistID, 10), Name: r.ArtistName})
		}
	}

	return out
}

func (a itunesAnswer) albums() []Album {
	var out []Album
	for _, r := range a.Results {
		if r.WrapperType == "collection" && r.CollectionID != 0 && r.CollectionName != "" {
			out = append(out, Album{ID: strconv.FormatInt(r.CollectionID, 10), Artist: r.ArtistName, Title: r.CollectionName, Released: dateOf(r.ReleaseDate)})
		}
	}

	return out
}

func (a itunesAnswer) tracks() []Track {
	var out []Track
	for _, r := range a.Results {
		if r.WrapperType == "track" && r.TrackName != "" && (r.Kind == "song" || r.Kind == "") {
			out = append(out, Track{Artist: r.ArtistName, Album: r.CollectionName, Title: r.TrackName, DurationMs: r.TrackTimeMs, Disc: r.DiscNumber, Number: r.TrackNumber})
		}
	}

	return out
}

// dateOf is the day of a timestamp such as "2001-03-13T08:00:00Z".
func dateOf(stamp string) string {
	if len(stamp) >= 10 {
		return stamp[:10]
	}

	return stamp
}
