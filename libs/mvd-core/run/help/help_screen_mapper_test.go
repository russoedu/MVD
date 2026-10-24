package help

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func plain(lines []string) string { return ansi.Strip(strings.Join(lines, "\n")) }

func TestTheDesktopHelpDoesNotTellAboutKeysItDoesNotHave(t *testing.T) {
	text := plain(helpLines(Options{Hosted: true}))
	for _, gone := range []string{"quit", "ctrl+c", "redraw", "add more links"} {
		if strings.Contains(text, gone) {
			t.Errorf("the desktop help mentions %q", gone)
		}
	}
	if !strings.Contains(text, "every keyboard shortcut of the app") {
		t.Error("the desktop help should point to the list of all shortcuts")
	}
}

func TestTheTerminalHelpStillTellsAboutQuitAndRedraw(t *testing.T) {
	text := plain(helpLines(Options{}))
	for _, want := range []string{"quit (asks for confirmation", "ctrl+c", "redraw the screen", "add more links"} {
		if !strings.Contains(text, want) {
			t.Errorf("the terminal help lacks %q", want)
		}
	}
}

func TestThereIsNoSentenceWithAMissingPath(t *testing.T) {
	if text := plain(helpLines(Options{})); strings.Contains(text, "written to") {
		t.Errorf("a log line with no path was shown: %q", text)
	}
	text := plain(helpLines(Options{LogPath: `C:\logs\mvd.log`}))
	if !strings.Contains(text, `full output is written to C:\logs\mvd.log`) {
		t.Errorf("the path was not shown: %q", text)
	}
}

func TestHelpSaysHowToLeaveIt(t *testing.T) {
	if text := plain(helpLines(Options{})); !strings.Contains(text, "press esc to close") {
		t.Errorf("the way out is missing: %q", text)
	}
}
