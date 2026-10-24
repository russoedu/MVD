package download

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestALongLogLineWrapsInsteadOfBeingCut(t *testing.T) {
	m := sampleModel(t)
	m.focus = paneEntries
	m.selEntry = 4
	last := "[download] line " + strings.Repeat("x", 49) // 65 cells, wider than the pane below

	rows, _ := m.outputWindow(30, 10)
	if len(rows) != 10 {
		t.Fatalf("want a full pane of 10 rows, got %d", len(rows))
	}
	for _, r := range rows {
		if w := ansi.StringWidth(r); w > 30 {
			t.Errorf("row %q is %d cells wide, wider than the pane", r, w)
		}
		if strings.Contains(r, "…") {
			t.Errorf("row %q was cut", r)
		}
	}
	if tail := strings.ReplaceAll(strings.Join(rows[len(rows)-3:], ""), " ", ""); !strings.HasSuffix(tail, strings.ReplaceAll(last, " ", "")) {
		t.Errorf("the last log line is not whole at the bottom: %q", rows)
	}
}

func TestAWideEnoughPaneShowsOneRowPerLogLine(t *testing.T) {
	m := sampleModel(t)
	m.focus = paneEntries
	m.selEntry = 4

	rows, first := m.outputWindow(200, 10)
	if len(rows) != 10 {
		t.Fatalf("want 10 rows, got %d", len(rows))
	}
	if first != len(m.outputLines())-10 {
		t.Errorf("the first row belongs to line %d, want %d", first, len(m.outputLines())-10)
	}
}

func TestScrollingStillMovesByLogLines(t *testing.T) {
	m := sampleModel(t)
	m.focus = paneEntries
	m.selEntry = 4
	atBottom, _ := m.outputWindow(30, 6)

	m.outScroll = 2
	scrolled, _ := m.outputWindow(30, 6)
	if strings.Join(scrolled, "|") == strings.Join(atBottom, "|") {
		t.Error("scrolling up two lines changed nothing")
	}
	if got := scrolled[len(scrolled)-1]; !strings.HasPrefix(got, "e") && !strings.Contains(got, "x") && !strings.Contains(got, "download") {
		t.Errorf("the bottom row after scrolling is not part of an earlier log line: %q", got)
	}
}

func TestAShortLogShowsEveryRowOnceWithRoomToSpare(t *testing.T) {
	m := sampleModel(t)
	m.focus = paneLists
	m.selPlaylist = 2 // a playlist with no log lines

	rows, first := m.outputWindow(40, 8)
	if len(rows) != 0 || first != 0 {
		t.Errorf("a playlist with nothing logged has no rows, got %q from %d", rows, first)
	}
}
