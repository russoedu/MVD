package dialog

import (
	"testing"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/look"
)

func TestEveryKindOfButtonCanBeReadInEveryTheme(t *testing.T) {
	defer look.Apply(config.DefaultThemeColors())
	for _, p := range config.ThemePresets {
		look.Apply(p.Colors)
		for name, kind := range map[string]PillKind{"normal": PillNormal, "primary": PillPrimary, "disabled": PillDisabled} {
			for _, hover := range []bool{false, true} {
				ground, text := pillColours(kind, hover)
				if got := config.Contrast(text, ground); got < config.MinTextContrast {
					t.Errorf("%s, %s button (hover %v): %s on %s is %.2f:1", p.Name, name, hover, text, ground, got)
				}
			}
		}
	}
}

func TestADisabledButtonIsDarkerThanOneThatWorksOnADarkTheme(t *testing.T) {
	defer look.Apply(config.DefaultThemeColors())
	look.Apply(config.DefaultThemeColors())
	normal, _ := pillColours(PillNormal, false)
	disabled, _ := pillColours(PillDisabled, false)
	if disabled >= normal {
		t.Errorf("the disabled ground %s should be darker than %s", disabled, normal)
	}
}

func TestMixingTwoColoursGoesPartOfTheWay(t *testing.T) {
	if got := mixColours("#000000", "#ffffff", 0.5); got != "#808080" {
		t.Errorf("half way is %s", got)
	}
	if got := mixColours("#102030", "#102030", 0.7); got != "#102030" {
		t.Errorf("the same colour mixed with itself is %s", got)
	}
}
