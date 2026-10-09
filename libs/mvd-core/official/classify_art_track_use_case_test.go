package official

import (
	"errors"
	"strings"
	"testing"
)

const artTrackID = "rnlp_avexYQ"

// artTrackPage is a watch page whose description is an auto-generated one.
func artTrackPage(musicID string) string {
	return strings.Replace(watchPage(artTrackID, musicID, ""), "Tricky", "Provided to YouTube by Label Tricky", 1)
}

func tagged(tag string) VideoTyper {
	return func(string) (string, error) { return tag, nil }
}

func quietLog(string, ...interface{}) {}

func TestAnArtTrackUnderTheArtistsNameIsResolved(t *testing.T) {
	f := &fakeYouTube{
		pages:   map[string]string{artTrackID: artTrackPage("-bsONE-kZwI")},
		authors: map[string]string{"-bsONE-kZwI": "London Records"},
	}
	_, res := f.server(t)
	res.VideoTypes = tagged("ATV")

	got, reason := res.ResolveLog(artTrackID, "Killer", "ATB", quietLog)
	if got != "-bsONE-kZwI" {
		t.Fatalf("want the official video, got %q (%s)", got, reason)
	}
}

func TestARealVideoIsLeftAloneWithoutFetchingAnything(t *testing.T) {
	f := &fakeYouTube{
		pages: map[string]string{artTrackID: artTrackPage("-bsONE-kZwI")},
	}
	_, res := f.server(t)
	for _, tag := range []string{"OMV", "UGC"} {
		res.VideoTypes = tagged(tag)
		got, reason := res.ResolveLog(artTrackID, "Song", "Artist", quietLog)
		if got != "" || !strings.Contains(reason, "already a video") || !strings.Contains(reason, tag) {
			t.Errorf("%s: want it left alone with the reason, got %q (%s)", tag, got, reason)
		}
	}
	if f.count("watch") != 0 || f.count("oembed") != 0 {
		t.Errorf("a real video needs no lookup, but the page was fetched %d times", f.count("watch"))
	}
}

func TestWithoutATagTheChannelAndTheDescriptionDecide(t *testing.T) {
	cases := []struct {
		name    string
		typer   VideoTyper
		channel string
		page    string
		want    string
	}{
		{"a Topic channel", tagged(""), "Artist - Topic", "", "-bsONE-kZwI"},
		{"YouTube Music is down, the description says art track", func(string) (string, error) { return "", errors.New("down") }, "ATB", artTrackPage("-bsONE-kZwI"), "-bsONE-kZwI"},
		{"the description is a normal one", tagged(""), "ATB", `<script>var ytInitialPlayerResponse = {"videoDetails":{"shortDescription":"Official video for the song. Subscribe!"}};</script>`, ""},
	}
	for _, c := range cases {
		f := &fakeYouTube{
			pages:   map[string]string{artTrackID: c.page},
			authors: map[string]string{"-bsONE-kZwI": "London Records"},
		}
		if c.page == "" {
			f.pages[artTrackID] = watchPage(artTrackID, "-bsONE-kZwI", "")
		}
		_, res := f.server(t)
		res.VideoTypes = c.typer

		got, reason := res.ResolveLog(artTrackID, "Song", c.channel, quietLog)
		if got != c.want {
			t.Errorf("%s: got %q (%s), want %q", c.name, got, reason, c.want)
		}
	}
}

func TestWithoutAVideoTyperEveryUploadIsLookedUp(t *testing.T) {
	f := &fakeYouTube{
		pages:   map[string]string{artTrackID: watchPage(artTrackID, "-bsONE-kZwI", "")},
		authors: map[string]string{"-bsONE-kZwI": "London Records"},
	}
	_, res := f.server(t)

	if got, _ := res.ResolveLog(artTrackID, "Song", "Band", quietLog); got != "-bsONE-kZwI" {
		t.Errorf("without a VideoTyper the behaviour is the one from before, got %q", got)
	}
}
