package official

import "testing"

func TestMentions(t *testing.T) {
	cases := []struct {
		text, name string
		want       bool
	}{
		{"BEYONCÉ - Halo", "Beyonce", true},
		{"Pump Up The Volume MARRS", "M/A/R/R/S", true},
		{"Pump Up The Volume M A R R S", "M/A/R/R/S", true},
		{"Courage - Song", "Rage", false},
		{"R E M - Losing My Religion", "R.E.M.", true},
		{"Some Title", "", false},
		{"A B - Song", "A B", true},
		{"AB - Song", "A B", false},
	}
	for _, c := range cases {
		if got := Mentions(c.text, c.name); got != c.want {
			t.Errorf("Mentions(%q, %q) = %v, want %v", c.text, c.name, got, c.want)
		}
	}
}
