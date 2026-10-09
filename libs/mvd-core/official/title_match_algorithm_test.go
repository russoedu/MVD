package official

import "testing"

func TestTitleSimilarity(t *testing.T) {
	cases := []struct {
		name      string
		song      string
		candidate string
		artists   []string
		min, max  float64
	}{
		{"exact", "Uptown Girl", "Billy Joel - Uptown Girl (Official Video)", []string{"Billy Joel"}, 1, 1},
		{"a year the video does not carry", "Killer 2000", "ATB - Killer", []string{"ATB"}, 1, 1},
		{"a year the song does not carry", "Killer", "ATB - Killer 2000 (HD)", []string{"ATB"}, 1, 1},
		{"a split word", "Run Away", "Real McCoy - Runaway", []string{"Real McCoy"}, 0.9, 1},
		{"remaster in brackets", "Them Bones (2022 Remaster)", "Alice In Chains - Them Bones (Official HD Video)", []string{"Alice In Chains"}, 1, 1},
		{"remaster after a dash", "Highway Star - Remastered 2012", "Deep Purple - Highway Star", []string{"Deep Purple"}, 1, 1},
		{"accents", "Nicole Kidman", "ADÉLA - Nicole Kidman (Official Video)", []string{"Adela"}, 1, 1},
		{"a song named like its artist's word", "Fire", "Earth, Wind & Fire - Fire", []string{"Earth Wind & Fire"}, 1, 1},
		{"only part of the song's words", "Alive And Kicking", "Simple Minds - Alive", []string{"Simple Minds"}, 0, 0.59},
		{"another song of the artist", "Alive And Kicking", "Simple Minds - Don't You (Forget About Me)", []string{"Simple Minds"}, 0, 0.2},
		{"nothing in common", "Wonderwall", "Oasis - Champagne Supernova", []string{"Oasis"}, 0, 0.1},
		{"empty", "", "Anything", nil, 0, 0},
	}
	for _, c := range cases {
		got := TitleSimilarity(c.song, c.candidate, c.artists)
		if got < c.min || got > c.max {
			t.Errorf("%s: TitleSimilarity(%q, %q) = %.2f, want between %.2f and %.2f", c.name, c.song, c.candidate, got, c.min, c.max)
		}
	}
}
