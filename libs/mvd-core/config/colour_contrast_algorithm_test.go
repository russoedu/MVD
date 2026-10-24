package config

import (
	"math"
	"testing"
)

func TestContrastIsTheWCAGRatio(t *testing.T) {
	if got := Contrast("#000", "#fff"); math.Abs(got-21) > 0.01 {
		t.Errorf("black on white is %.2f, want 21", got)
	}
	if got := Contrast("#777777", "#777777"); math.Abs(got-1) > 0.001 {
		t.Errorf("a colour on itself is %.2f, want 1", got)
	}
	if got := Contrast("#6b7280", "#000000"); got > 4.4 || got < 4.3 {
		t.Errorf("the old default dim colour was %.2f, want about 4.34", got)
	}
	if Contrast("nope", "#000") != 0 {
		t.Error("something that is not a colour has no contrast")
	}
}

func TestEveryThemeKeepsItsSmallTextReadable(t *testing.T) {
	for _, p := range ThemePresets {
		background := p.Colors.Background
		if background == "" {
			background = "#000000" // the terminal's own is dark
		}
		for name, colour := range map[string]string{"Dim": p.Colors.Dim, "Text": p.Colors.Text} {
			if got := Contrast(colour, background); got < MinTextContrast {
				t.Errorf("%s: %s %s on %s is %.2f:1, below %.1f:1", p.Name, name, colour, background, got, MinTextContrast)
			}
		}
	}
}

func TestATreatedThemeSavedWithTheOlderDimColourIsUpgraded(t *testing.T) {
	cfg := Default("/app", "/music")
	cfg.ThemeName = "Dracula"
	cfg.Colors, _ = PresetColors("Dracula")
	cfg.Colors.Dim = "#6272a4" // what older versions saved
	settleTheme(&cfg)
	if cfg.ThemeName != "Dracula" {
		t.Errorf("an untouched theme became %q", cfg.ThemeName)
	}
	if want, _ := PresetColors("Dracula"); cfg.Colors.Dim != want.Dim {
		t.Errorf("the dim colour is %s, want %s", cfg.Colors.Dim, want.Dim)
	}
}

func TestATreatedThemeTheUserChangedKeepsTheirColours(t *testing.T) {
	cfg := Default("/app", "/music")
	cfg.ThemeName = "Dracula"
	cfg.Colors, _ = PresetColors("Dracula")
	cfg.Colors.Dim = "#6272a4"
	cfg.Colors.Accent = "#000000" // they changed something else, so it is theirs
	settleTheme(&cfg)
	if cfg.ThemeName != CustomTheme || cfg.Colors.Dim != "#6272a4" {
		t.Errorf("a theme the person changed was touched: %s %s", cfg.ThemeName, cfg.Colors.Dim)
	}
}
