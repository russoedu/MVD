package addsource

import (
	"strings"
	"testing"

	"youtube-downloader/libs/mvd-core/musicsearch"
)

func links(text string) ([]string, string) {
	if strings.HasPrefix(text, "http") {
		return strings.Fields(text), ""
	}

	return nil, "That does not look like a link"
}

func TestTheWindowOpensOnThePlatformAddress(t *testing.T) {
	s := New()
	if s.Mode != ModeURL || s.Focus != FieldURL {
		t.Errorf("opened on %+v", s)
	}
	help := strings.Join(Help(s.Mode), "\n")
	if !strings.Contains(help, "Spotify") || !strings.Contains(help, "YouTube playlist") || !strings.Contains(help, "Apple Music") {
		t.Errorf("the help does not say what links are taken: %v", help)
	}
}

func TestTabGoesThroughTheFieldsOfTheModeAndRoundAgain(t *testing.T) {
	s := New().WithMode(ModeSearch)
	if s.Focus != FieldArtist {
		t.Fatalf("a search begins on the artist, not %v", s.Focus)
	}
	var seen []Field
	for i := 0; i < len(s.Fields()); i++ {
		seen = append(seen, s.Focus)
		s = s.Next()
	}
	want := []Field{FieldArtist, FieldAlbum, FieldSong, FieldAdd, FieldCancel, FieldMode}
	for i := range want {
		if seen[i] != want[i] {
			t.Fatalf("order %v, want %v", seen, want)
		}
	}
	if got := s.Focused(FieldMode).Previous().Focus; got != FieldCancel {
		t.Errorf("back from the mode goes to the last, not %v", got)
	}
}

func TestChangingTheModeKeepsWhatWasWrittenInTheOther(t *testing.T) {
	s := New().SetText(FieldURL, "https://youtu.be/x").WithMode(ModeSearch).SetText(FieldArtist, "Daft Punk").WithMode(ModeURL)
	if s.URLs != "https://youtu.be/x" || s.Artist != "Daft Punk" {
		t.Errorf("lost: %+v", s)
	}
}

func TestASearchNeedsAnArtistAndGivesItsAddress(t *testing.T) {
	s := New().WithMode(ModeSearch).SetText(FieldSong, "One More Time")
	r := s.Submit(links)
	if r.Done || r.State.Problem != "A song needs its artist" || r.State.Focus != FieldArtist {
		t.Errorf("a song alone: %+v", r)
	}
	r = s.SetText(FieldArtist, "Daft Punk").Submit(links)
	want := musicsearch.Query{Artist: "Daft Punk", Song: "One More Time"}.Address()
	if !r.Done || len(r.Addresses) != 1 || r.Addresses[0] != want {
		t.Errorf("artist and song: %+v", r)
	}
	r = New().WithMode(ModeSearch).SetText(FieldAlbum, "Discovery").Submit(links)
	if r.Done || r.State.Problem != "An album needs its artist" {
		t.Errorf("an album alone: %+v", r)
	}
}

func TestALinkIsReadWithTheRulesOfTheRenderer(t *testing.T) {
	if r := New().Submit(links); r.Done || r.State.Problem != "Paste a link" {
		t.Errorf("nothing written: %+v", r)
	}
	if r := New().SetText(FieldURL, "hello").Submit(links); r.Done || r.State.Problem == "" {
		t.Errorf("not a link: %+v", r)
	}
	r := New().SetText(FieldURL, "https://a https://b").Submit(links)
	if !r.Done || len(r.Addresses) != 2 {
		t.Errorf("two links: %+v", r)
	}
}
