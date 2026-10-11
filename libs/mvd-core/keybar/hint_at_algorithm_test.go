package keybar

import "testing"

func TestAtFindsTheHintUnderAColumn(t *testing.T) {
	hints := []Hint{{Key: "↑↓", Desc: "move"}, {Key: "q", Desc: "quit"}}
	// " ↑↓ move  q quit": columns 1-7 are the first hint, 10-15 the second.
	cases := []struct {
		x    int
		want string
		ok   bool
	}{{0, "", false}, {1, "↑↓", true}, {7, "↑↓", true}, {8, "", false}, {9, "", false}, {10, "q", true}, {15, "q", true}, {16, "", false}}
	for _, c := range cases {
		got, ok := At(hints, c.x)
		if ok != c.ok || got.Key != c.want {
			t.Errorf("At(%d) = %q, %v, want %q, %v", c.x, got.Key, ok, c.want, c.ok)
		}
	}
}

func TestOfKeepsKeyAndLabel(t *testing.T) {
	type typed struct{ Key, Label string }
	got := Of[typed]([]Hint{{Key: "q", Desc: "quit"}})
	if len(got) != 1 || got[0].Key != "q" || got[0].Label != "quit" {
		t.Errorf("got %v", got)
	}
}
