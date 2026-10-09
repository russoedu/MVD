package appwindow

import "testing"

func TestKeepOpen(t *testing.T) {
	yes, no := func() bool { return true }, func() bool { return false }
	never := func() bool { t.Error("nobody should be asked"); return true }

	for name, c := range map[string]struct {
		busy, ask func() bool
		want      bool
	}{
		"nothing is running":        {no, never, false},
		"no way to tell":            {nil, never, false},
		"running, the person quits": {yes, yes, false},
		"running, the person stays": {yes, no, true},
	} {
		if got := keepOpen(c.busy, c.ask); got != c.want {
			t.Errorf("%s: got %v, want %v", name, got, c.want)
		}
	}
}
