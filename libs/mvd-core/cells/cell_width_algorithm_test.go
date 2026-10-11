package cells

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestTruncateEndsInAnEllipsis(t *testing.T) {
	if got := Truncate("Stardust", 5); got != "Star…" {
		t.Errorf("got %q", got)
	}
	if got := Truncate("Stardust", 20); got != "Stardust" {
		t.Errorf("a line that fits is kept, got %q", got)
	}
}

func TestFitPadsAndCuts(t *testing.T) {
	for _, w := range []int{3, 8, 12} {
		if got := ansi.StringWidth(Fit("Stardust", w)); got != w {
			t.Errorf("Fit to %d is %d cells wide", w, got)
		}
	}
}
