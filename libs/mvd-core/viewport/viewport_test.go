package viewport

import "testing"

func TestStartKeepsTheSelectedRowCentredAndInView(t *testing.T) {
	cases := []struct{ n, sel, h, want int }{
		{10, 0, 4, 0}, {10, 5, 4, 3}, {10, 9, 4, 6}, {3, 2, 4, 0}, {10, 5, 0, 0},
	}
	for _, c := range cases {
		if got := Start(c.n, c.sel, c.h); got != c.want {
			t.Errorf("Start(%d, %d, %d) = %d, want %d", c.n, c.sel, c.h, got, c.want)
		}
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
