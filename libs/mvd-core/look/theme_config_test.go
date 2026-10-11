package look

import (
	"testing"

	"charm.land/lipgloss/v2"

	"youtube-downloader/libs/mvd-core/config"
)

func TestApplyBuildsTheStylesAndThePaletteFromTheColours(t *testing.T) {
	t.Cleanup(func() { Apply(config.DefaultThemeColors()) })

	c := config.DefaultThemeColors()
	c.Accent = "#112233"
	Apply(c)

	if got := Title.GetForeground(); got != lipgloss.Color("#112233") {
		t.Errorf("title colour = %v, want #112233", got)
	}
	if got := CurrentPalette().Accent; got != "#112233" {
		t.Errorf("palette accent = %q, want #112233", got)
	}
}
