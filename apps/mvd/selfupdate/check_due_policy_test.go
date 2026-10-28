package selfupdate

import (
	"testing"
	"time"
)

func TestDue(t *testing.T) {
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	cases := map[string]struct {
		last time.Time
		want bool
	}{
		"never":      {time.Time{}, true},
		"an hour":    {now.Add(-time.Hour), false},
		"a day":      {now.Add(-24 * time.Hour), true},
		"future":     {now.Add(time.Hour), true},
		"just under": {now.Add(-23*time.Hour - 59*time.Minute), false},
	}
	for name, c := range cases {
		if got := due(c.last, now); got != c.want {
			t.Errorf("%s: due = %v, want %v", name, got, c.want)
		}
	}
}
