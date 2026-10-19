package workspace

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
)

func TestAnOpenMenuCoversOnlyItsOwnBox(t *testing.T) {
	m := sized(New(Host{}, config.Config{}, downloadlist.New([]string{"https://youtu.be/x"})))
	closed := strings.Split(ansi.Strip(m.render()), "\n")
	m.bar.SetOpen(0)
	open := strings.Split(ansi.Strip(m.render()), "\n")
	if len(open) != len(closed) {
		t.Fatalf("%d rows, was %d", len(open), len(closed))
	}
	if !strings.Contains(strings.Join(open, "\n"), entryQuit) {
		t.Error("the menu is not shown")
	}
	row := 3 // inside the box
	if !strings.HasSuffix(open[row], closed[row][len(closed[row])-30:]) {
		t.Errorf("the screen right of the menu was covered: %q", open[row])
	}
}
