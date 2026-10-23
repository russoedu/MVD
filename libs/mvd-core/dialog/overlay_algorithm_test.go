package dialog

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestAPopupCoversOnlyItsOwnCells(t *testing.T) {
	base := strings.Join([]string{"aaaaaaaaaa", "bbbbbbbbbb", "cccccccccc"}, "\n")
	got := ansi.Strip(Overlay(base, "XX\nYY", 3, 1))
	want := "aaaaaaaaaa\nbbbXXbbbbb\ncccYYccccc"
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestAPopupKeepsTheColoursAroundIt(t *testing.T) {
	base := "\x1b[31mredredred\x1b[0m"
	got := Overlay(base, "..", 3, 0)
	if ansi.Strip(got) != "red..dred" {
		t.Errorf("text %q", ansi.Strip(got))
	}
	if !strings.Contains(got, "\x1b[31m") {
		t.Error("the colour of the screen behind was lost")
	}
}

func TestAPopupOffTheScreenIsClippedNotAPanic(t *testing.T) {
	got := Overlay("ab\ncd", "XXXX\nYYYY\nZZZZ", 0, 1)
	if ansi.Strip(got) != "ab\nXXXX" {
		t.Errorf("got %q", ansi.Strip(got))
	}
}

func TestFrameHasTheSizeAskedFor(t *testing.T) {
	box := ansi.Strip(Frame("Preferences", []string{"one", "two"}, 20, 6))
	lines := strings.Split(box, "\n")
	if len(lines) != 6 {
		t.Fatalf("%d rows", len(lines))
	}
	for i, l := range lines {
		if ansi.StringWidth(l) != 20 {
			t.Errorf("row %d is %d wide: %q", i, ansi.StringWidth(l), l)
		}
	}
	if !strings.Contains(lines[0], "Preferences") || !strings.Contains(lines[1], "one") {
		t.Errorf("box %q", box)
	}
}

func TestTabsAndButtonsSayWhereTheyAre(t *testing.T) {
	_, tabs := Tabs([]string{"Output", "Download"}, 0)
	if At(tabs, 1) != 0 || At(tabs, tabs[1].X0+1) != 1 || At(tabs, 200) != -1 {
		t.Errorf("tabs %+v", tabs)
	}
	line, buttons := Buttons([]string{"Save", "Cancel"}, 0, map[string]bool{"Cancel": true})
	if got := ansi.Strip(line); got != "[ Save ]  [ Cancel ]" {
		t.Errorf("buttons %q", got)
	}
	if At(buttons, buttons[1].X0) != 1 {
		t.Errorf("buttons %+v", buttons)
	}
}

func TestAClosableFrameKeepsItsWidthAndTellsWhereItsCloseMarkIs(t *testing.T) {
	for _, title := range []string{"", "About MVD"} {
		box := ansi.Strip(ClosableFrame(title, []string{"x"}, 30, 4))
		top := strings.Split(box, "\n")[0]
		if ansi.StringWidth(top) != 30 {
			t.Errorf("%q: top is %d wide", title, ansi.StringWidth(top))
		}
		runes := []rune(top)
		if got := string(runes[CloseSpan(30).X0 : CloseSpan(30).X1]); got != " × " {
			t.Errorf("%q: the close span holds %q", title, got)
		}
	}
}

func TestAPopupTakesAShareOfTheScreenWithinLimits(t *testing.T) {
	if w, h := Sized(200, 50, 85, 40, 12); w != 170 || h != 42 {
		t.Errorf("big screen: %d x %d", w, h)
	}
	if w, h := Sized(30, 10, 85, 40, 12); w != 28 || h != 8 {
		t.Errorf("tiny screen: %d x %d", w, h)
	}
}
