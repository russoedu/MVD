package catalog

import "testing"

func TestAnArtistIsNamedDespiteTheSpacingAndTheSpelling(t *testing.T) {
	yes := [][2]string{{"daft punk", "Daft Punk"}, {"AC/DC", "ACDC"}, {"guns n roses", "Guns N' Roses"}, {"Beyonce", "Beyoncé"}}
	for _, c := range yes {
		if !artistIsNamed(c[0], c[1]) {
			t.Errorf("%q should be %q", c[0], c[1])
		}
	}
	no := [][2]string{{"ATB", "ABBA"}, {"Madonna", "Madness"}}
	for _, c := range no {
		if artistIsNamed(c[0], c[1]) {
			t.Errorf("%q should not be %q", c[0], c[1])
		}
	}
}

func TestACreditThatMentionsTheArtistCounts(t *testing.T) {
	if !artistIsCredited("Daft Punk", "Daft Punk & Pharrell Williams") {
		t.Error("a song with a guest is the artist's")
	}
	if artistIsCredited("Daft Punk", "Daft Punk Tribute Band") {
		t.Error("a tribute band is not the artist")
	}
	if artistIsCredited("Daft Punk", "Justice") {
		t.Error("another artist is not the artist")
	}
}

func TestKaraokeAndVersionsAreRecognisedByWholeWords(t *testing.T) {
	if !isUnwanted("Discovery (Karaoke Version)") || isUnwanted("Discovery") {
		t.Error("karaoke")
	}
	if versionRank("Alive 2007") != 0 || versionRank("Alive 2007 (Live)") != 1 || versionRank("Around the World (Remix)") != 1 {
		t.Error("a word inside another word is not a version")
	}
}
