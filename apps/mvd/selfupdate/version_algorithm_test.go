package selfupdate

import "testing"

func TestIsNewer(t *testing.T) {
	cases := []struct {
		have, latest string
		want         bool
	}{
		{"0.6.1", "0.6.2", true},
		{"0.6.1", "0.10.0", true},
		{"0.6.1", "1.0.0", true},
		{"0.6.1", "0.6.1", false},
		{"0.6.1", "0.6.0", false},
		{"dev", "9.9.9", false},
		{"0.6.1", "v0.7.0", false},
		{"0.6.1", "", false},
	}
	for _, c := range cases {
		if got := isNewer(c.have, c.latest); got != c.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", c.have, c.latest, got, c.want)
		}
	}
}
