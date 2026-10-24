package fetch

import "testing"

func TestTheQualityYtDlpWritesIsRead(t *testing.T) {
	cases := []struct {
		text         string
		height, kbps int
		found        bool
	}{
		{"240|106.064\n", 240, 106, true},
		{"1080|NA\n", 1080, 0, true},
		{"NA|160.5\n", 0, 161, true},
		{"NA|NA\n", 0, 0, false},
		{"", 0, 0, false},
		{"garbage", 0, 0, false},
		{"2160|129.3\nsomething else", 2160, 129, true},
	}
	for _, c := range cases {
		height, kbps, found := ParseQualityLine(c.text)
		if height != c.height || kbps != c.kbps || found != c.found {
			t.Errorf("%q: got %d, %d, %v; want %d, %d, %v", c.text, height, kbps, found, c.height, c.kbps, c.found)
		}
	}
}
