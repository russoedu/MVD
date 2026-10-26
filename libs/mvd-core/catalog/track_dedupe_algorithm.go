package catalog

import (
	"sort"

	"youtube-downloader/libs/mvd-core/official/titlematch"
)

// uniqueTracks keeps the first song of each name: the same song comes again on a compilation, a
// deluxe edition or as a single, spelt a little differently ("Intro (Remastered 2011)"), and is one
// song to download.
func uniqueTracks(tracks []Track) []Track {
	seen := map[string]bool{}
	out := make([]Track, 0, len(tracks))
	for _, t := range tracks {
		key := titlematch.Normalize(titlematch.SearchTitle(t.Title))
		if key == "" {
			key = titlematch.Normalize(t.Title)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, t)
	}

	return out
}

// inAlbumOrder is the songs of one album in the order of its discs and numbers; songs that say none stay
// as they came.
func inAlbumOrder(tracks []Track) []Track {
	out := append([]Track(nil), tracks...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Disc != out[j].Disc && out[i].Disc != 0 && out[j].Disc != 0 {
			return out[i].Disc < out[j].Disc
		}
		if out[i].Number == 0 || out[j].Number == 0 {
			return false
		}

		return out[i].Number < out[j].Number
	})

	return out
}

// inReleaseOrder is the albums with the music itself before its versions, each in the order it came
// out, so that the first appearance of a song, which is the one kept, is its own album.
func inReleaseOrder(albums []Album) []Album {
	out := append([]Album(nil), albums...)
	date := func(a Album) string {
		if a.Released == "" {
			return "9999"
		}

		return a.Released
	}
	sort.SliceStable(out, func(i, j int) bool {
		if ri, rj := versionRank(out[i].Title), versionRank(out[j].Title); ri != rj {
			return ri < rj
		}

		return date(out[i]) < date(out[j])
	})

	return out
}
