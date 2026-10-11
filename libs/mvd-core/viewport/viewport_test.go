package viewport

import "testing"

func TestRowsKeepTheSelectedRowInView(t *testing.T) {
	all := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for sel := range all {
		got := Rows(all, sel, 4)
		if len(got) != 4 {
			t.Fatalf("sel %d: %d rows, want 4", sel, len(got))
		}
		if got[0] != Start(len(all), sel, 4) {
			t.Errorf("sel %d: window starts at %d, Start says %d", sel, got[0], Start(len(all), sel, 4))
		}
		found := false
		for _, v := range got {
			found = found || v == sel
		}
		if !found {
			t.Errorf("sel %d is not in %v", sel, got)
		}
	}
}

func TestRowsKeepsAShortListWhole(t *testing.T) {
	if got := Rows([]int{1, 2}, 1, 5); len(got) != 2 {
		t.Errorf("got %v", got)
	}
}

func TestClamp(t *testing.T) {
	cases := [][4]int{{5, 0, 3, 3}, {-1, 0, 3, 0}, {2, 0, 3, 2}, {4, 2, 1, 2}}
	for _, c := range cases {
		if got := Clamp(c[0], c[1], c[2]); got != c[3] {
			t.Errorf("Clamp(%d, %d, %d) = %d, want %d", c[0], c[1], c[2], got, c[3])
		}
	}
}
