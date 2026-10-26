package musicsearch

import "testing"

func TestOnlyAnArtistAloneOrWithAnAlbumOrASongCanBeSearched(t *testing.T) {
	cases := []struct {
		q       Query
		kind    Kind
		problem string
	}{
		{Query{Artist: "Daft Punk"}, Artist, ""},
		{Query{Artist: "Daft Punk", Album: "Discovery"}, Album, ""},
		{Query{Artist: "Daft Punk", Song: "One More Time"}, Song, ""},
		{Query{Artist: "Daft Punk", Album: "Discovery", Song: "One More Time"}, Song, ""},
		{Query{Album: "Discovery"}, Invalid, "An album needs its artist"},
		{Query{Song: "One More Time"}, Invalid, "A song needs its artist"},
		{Query{Album: "Discovery", Song: "One More Time"}, Invalid, "A song needs its artist"},
		{Query{Artist: "  "}, Invalid, "Write an artist"},
	}
	for _, c := range cases {
		if got := c.q.Kind(); got != c.kind {
			t.Errorf("%+v: kind %v, want %v", c.q, got, c.kind)
		}
		if got := c.q.Problem(); got != c.problem {
			t.Errorf("%+v: problem %q, want %q", c.q, got, c.problem)
		}
	}
}

func TestAQueryKeepsItsPartsInItsAddress(t *testing.T) {
	q := Query{Artist: "Florence + The Machine", Album: "Lungs", Song: "Dog Days Are Over & more"}
	back, ok := Parse(q.Address())
	if !ok || back != q {
		t.Errorf("got %+v, %v from %q", back, ok, q.Address())
	}
	if !Is(q.Address()) || Is("https://youtu.be/x") {
		t.Error("only the address of a query is one")
	}
	if _, ok := Parse(Scheme + "album=Lungs"); ok {
		t.Error("an address with no artist was taken")
	}
}

func TestAQueryIsReadAsAnArtistThenTheAlbumThenTheSong(t *testing.T) {
	if got := (Query{Artist: " Daft Punk ", Album: "Discovery"}).Label(); got != "Daft Punk – Discovery" {
		t.Errorf("label %q", got)
	}
}
