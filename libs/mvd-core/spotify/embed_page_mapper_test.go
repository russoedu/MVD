package spotify

import "testing"

const embedFixture = `<html><body><script id="__NEXT_DATA__" type="application/json">{"props":{"pageProps":{"state":{"data":{"entity":{"type":"playlist","title":"Road trip","trackList":[
{"title":"Blinding Lights","subtitle":"The Weeknd","duration":200040},
{"title":"","subtitle":"nobody","duration":1},
{"title":"Under Pressure","subtitle":"Queen, David Bowie","duration":248000}
]}}}}}}</script></body></html>`

func TestParseEmbedPage(t *testing.T) {
	got, err := ParseEmbedPage(embedFixture)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Road trip" || len(got.Tracks) != 2 {
		t.Fatalf("got %+v, want the playlist and its 2 titled tracks", got)
	}
	if got.Tracks[1] != (Track{Title: "Under Pressure", Artist: "Queen, David Bowie", DurationMs: 248000}) {
		t.Errorf("track = %+v", got.Tracks[1])
	}
}

func TestParseEmbedPageRefusesWhatIsNotAPlaylist(t *testing.T) {
	pages := map[string]string{
		"no data":  "<html></html>",
		"bad json": `<script id="__NEXT_DATA__" type="application/json">{nope</script>`,
		"an album": `<script id="__NEXT_DATA__" type="application/json">{"props":{"pageProps":{"state":{"data":{"entity":{"type":"album"}}}}}}</script>`,
	}
	for name, html := range pages {
		if _, err := ParseEmbedPage(html); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}
