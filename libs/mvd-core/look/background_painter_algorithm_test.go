package look

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/config"
)

func TestAScreenIsPaintedWithTheBackgroundOfTheTheme(t *testing.T) {
	defer Apply(config.DefaultThemeColors())

	if got := Paint("plain"); got != "plain" {
		t.Errorf("a theme with no background changed the screen: %q", got)
	}

	c := config.DefaultThemeColors()
	c.Background = "#1e1e1e"
	Apply(c)
	got := Paint("ab\n\x1b[31mred\x1b[0m end")
	if !strings.HasPrefix(got, "\x1b[48;2;30;30;30m") {
		t.Errorf("the screen does not begin with the background: %q", got)
	}
	if !strings.Contains(got, "\x1b[0m\x1b[48;2;30;30;30m") {
		t.Errorf("the background is not set again after a reset: %q", got)
	}
	for i, line := range strings.Split(got, "\n") {
		if w := ansi.StringWidth(line); w != 7 {
			t.Errorf("line %d is %d wide, want 7", i, w)
		}
	}
}
