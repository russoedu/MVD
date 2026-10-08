package applemusic

import "testing"

const pageFixture = `<html><head>
<script id=schema:music-playlist type="application/ld+json">{"@type":"MusicPlaylist","name":"Today&#x2019;s Hits"}</script>
<script type="application/json" id="serialized-server-data">{"data":[{"data":{"sections":[
 {"items":[{"id":"header","title":"Today's Hits","artistName":"Apple Music"}]},
 {"items":[
  {"id":"track-lockup - pl.x - 1","title":"Solar Eclipse","artistName":"Drake & Don Toliver","duration":218389},
  {"id":"track-lockup - pl.x - 2","title":"Blinding Lights","artistName":"The Weeknd","duration":200040},
  {"id":"track-lockup - pl.x - 1","title":"Solar Eclipse","artistName":"Drake & Don Toliver","duration":218389}
 ]},
 {"items":[{"id":"album-lockup - 9","title":"Some album","artistName":"Someone"}]}
]}}]}</script></head></html>`

func TestParsePlaylistPage(t *testing.T) {
	got, err := ParsePlaylistPage(pageFixture)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Today’s Hits" {
		t.Errorf("title = %q", got.Title)
	}
	want := []Track{
		{Title: "Solar Eclipse", Artist: "Drake & Don Toliver", DurationMs: 218389},
		{Title: "Blinding Lights", Artist: "The Weeknd", DurationMs: 200040},
	}
	if len(got.Tracks) != len(want) || got.Tracks[0] != want[0] || got.Tracks[1] != want[1] {
		t.Errorf("tracks = %+v, want the two songs once each, in order, and nothing else", got.Tracks)
	}
}

func TestParsePlaylistPageNeedsSongs(t *testing.T) {
	pages := map[string]string{
		"no data":  "<html></html>",
		"bad json": `<script type="application/json" id="serialized-server-data">{nope</script>`,
		"no songs": `<script type="application/json" id="serialized-server-data">{"data":[]}</script>`,
	}
	for name, page := range pages {
		if _, err := ParsePlaylistPage(page); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

func TestParsePlaylistPageWithoutAName(t *testing.T) {
	page := `<script type="application/json" id="serialized-server-data">{"x":{"id":"track-lockup - a","title":"Song","artistName":"Band"}}</script>`
	got, err := ParsePlaylistPage(page)
	if err != nil || got.Title == "" || len(got.Tracks) != 1 {
		t.Errorf("got %+v, %v; want a default name and the song", got, err)
	}
}
