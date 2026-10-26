package catalog

import "testing"

func TestTheSameSongOnAnotherAlbumIsOneSong(t *testing.T) {
	got := uniqueTracks([]Track{
		{Title: "Intro", Album: "A"},
		{Title: "Intro (Remastered 2011)", Album: "Best of"},
		{Title: "Outro", Album: "A"},
	})
	if len(got) != 2 || got[0].Album != "A" || got[1].Title != "Outro" {
		t.Errorf("got %+v", got)
	}
}

func TestSongsAreOrderedByDiscThenNumberAndThoseWithNoneStay(t *testing.T) {
	got := inAlbumOrder([]Track{{Title: "c", Disc: 1, Number: 3}, {Title: "a", Disc: 1, Number: 1}, {Title: "d", Disc: 2, Number: 1}, {Title: "b", Disc: 1, Number: 2}})
	want := []string{"a", "b", "c", "d"}
	for i, w := range want {
		if got[i].Title != w {
			t.Fatalf("got %+v", got)
		}
	}
	same := inAlbumOrder([]Track{{Title: "x"}, {Title: "y"}})
	if same[0].Title != "x" {
		t.Errorf("songs with no numbers moved: %+v", same)
	}
}

func TestAlbumsComeInReleaseOrderWithTheirVersionsLast(t *testing.T) {
	got := inReleaseOrder([]Album{
		{Title: "Alive 2007 (Live)", Released: "2007-01-01"},
		{Title: "Discovery", Released: "2001-03-13"},
		{Title: "Homework", Released: "1997-01-20"},
		{Title: "Unknown date"},
	})
	want := []string{"Homework", "Discovery", "Unknown date", "Alive 2007 (Live)"}
	for i, w := range want {
		if got[i].Title != w {
			t.Fatalf("got %+v", got)
		}
	}
}
