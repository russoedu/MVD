package official

import "testing"

func TestArtistFromChannel(t *testing.T) {
	if got := ArtistFromChannel("Oasis - Topic"); got != "Oasis" {
		t.Errorf("got %q", got)
	}
	if got := ArtistFromChannel("  Oasis  "); got != "Oasis" {
		t.Errorf("got %q", got)
	}
}

func TestNormalizeFoldsAccentsAndPunctuation(t *testing.T) {
	cases := map[string]string{
		"ADÉLA":                  "adela",
		"Beyoncé - Halo":         "beyonce halo",
		"Don't Stop (Believin')": "don t stop believin",
		"  Mr.  Brightside ":     "mr brightside",
	}
	for in, want := range cases {
		if got := normalize(in); got != want {
			t.Errorf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}
