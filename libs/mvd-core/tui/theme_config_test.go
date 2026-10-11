package tui

import (
	"testing"
	"youtube-downloader/libs/mvd-core/look"

	"charm.land/lipgloss/v2"

	"youtube-downloader/libs/mvd-core/config"
)

func TestApplyThemeSetsTheColours(t *testing.T) {
	t.Cleanup(func() { ApplyTheme(config.DefaultThemeColors()) })

	c := config.DefaultThemeColors()
	c.Accent = "#112233"
	ApplyTheme(c)

	if got := look.Title.GetForeground(); got != lipgloss.Color("#112233") {
		t.Errorf("title style = %v, want the accent colour #112233", got)
	}
}
