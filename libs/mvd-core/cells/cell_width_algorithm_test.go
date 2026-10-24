package cells

import (
	"strings"
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

func TestWrapLineKeepsEveryRowWithinTheWidth(t *testing.T) {
	line := "ERROR: [youtube] dPAhFv57gpc: Video unavailable. This video is no longer available because the account was closed"
	rows := WrapLine(line, 30)
	if len(rows) < 4 {
		t.Fatalf("a long line should take several rows, got %d: %q", len(rows), rows)
	}
	for _, r := range rows {
		if w := ansi.StringWidth(r); w > 30 {
			t.Errorf("row %q is %d cells wide", r, w)
		}
	}
	if got := strings.Join(rows, " "); strings.Join(strings.Fields(got), " ") != line {
		t.Errorf("wrapping lost words: %q", rows)
	}
}

func TestWrapLineBreaksAWordThatIsLongerThanTheWidth(t *testing.T) {
	rows := WrapLine("https://www.youtube.com/watch?v=dPAhFv57gpc&list=PLabcdefghijklmnop", 20)
	for _, r := range rows {
		if w := ansi.StringWidth(r); w > 20 {
			t.Errorf("row %q is %d cells wide", r, w)
		}
	}
	if strings.Join(rows, "") != "https://www.youtube.com/watch?v=dPAhFv57gpc&list=PLabcdefghijklmnop" {
		t.Errorf("a long address lost characters: %q", rows)
	}
}

func TestWrapLineKeepsAShortLineAsItIs(t *testing.T) {
	if rows := WrapLine("  indented  text", 40); len(rows) != 1 || rows[0] != "  indented  text" {
		t.Errorf("got %q", rows)
	}
	if rows := WrapLine("", 40); len(rows) != 1 || rows[0] != "" {
		t.Errorf("a blank line is one row, got %q", rows)
	}
	if rows := WrapLine("text", 0); rows != nil {
		t.Errorf("no width gives no rows, got %q", rows)
	}
}

func TestWrapLineKeepsTheColours(t *testing.T) {
	rows := WrapLine("\x1b[31mred words that need more than one row\x1b[0m", 12)
	if len(rows) < 2 {
		t.Fatalf("expected several rows, got %q", rows)
	}
	if !strings.Contains(strings.Join(rows, ""), "\x1b[31m") {
		t.Errorf("the colour was lost: %q", rows)
	}
}

func TestTruncateMiddleKeepsBothEndsOfAPath(t *testing.T) {
	path := `C:\Users\russo\AppData\Local\Temp\claude\sandbox\out`
	for _, w := range []int{20, 30, 40} {
		got := TruncateMiddle(path, w)
		if ansi.StringWidth(got) > w {
			t.Errorf("width %d: %q is %d cells", w, got, ansi.StringWidth(got))
		}
		if !strings.Contains(got, "…") || !strings.HasPrefix(got, "C:") || !strings.HasSuffix(got, `\out`) {
			t.Errorf("width %d: both ends should stay: %q", w, got)
		}
	}
	if got := TruncateMiddle("short", 20); got != "short" {
		t.Errorf("a path that fits is kept, got %q", got)
	}
	if got := TruncateMiddle("abcdefgh", 3); ansi.StringWidth(got) > 3 {
		t.Errorf("a tiny width still fits, got %q", got)
	}
	if got := TruncateMiddle("abc", 0); got != "" {
		t.Errorf("no room gives nothing, got %q", got)
	}
}
