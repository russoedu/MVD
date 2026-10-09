package tui

import (
	"testing"

	"charm.land/lipgloss/v2"

	"youtube-downloader/libs/mvd-core/config"
)

func TestApplyThemeSetsTheColours(t *testing.T) {
	t.Cleanup(func() { ApplyTheme(config.DefaultThemeColors()) })

	c := config.DefaultThemeColors()
	c.Accent = "#112233"
	ApplyTheme(c)

	if colMagenta != lipgloss.Color("#112233") {
		t.Errorf("accent = %q, want #112233", colMagenta)
	}
	if got := styTitle.GetForeground(); got != colMagenta {
		t.Errorf("title style = %v, want the accent colour", got)
	}
}
