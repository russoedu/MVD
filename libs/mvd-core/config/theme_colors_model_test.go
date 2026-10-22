package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidColour(t *testing.T) {
	for _, v := range []string{"#fff", "#FF007f", " #00f0ff "} {
		if !ValidColour(v) {
			t.Errorf("%q should be valid", v)
		}
	}
	for _, v := range []string{"", "red", "ff007f", "#ff", "#ff007", "#gg0000", "#ff007f00"} {
		if ValidColour(v) {
			t.Errorf("%q should be invalid", v)
		}
	}
}

func TestColoursKeepDefaultsUnlessValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.conf")
	if err := os.WriteFile(path, []byte("color_accent=#123456\ncolor_focus=blue\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := parseInto(path, Default(dir, dir))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Colors.Accent != "#123456" {
		t.Errorf("accent = %q, want the configured colour", cfg.Colors.Accent)
	}
	if cfg.Colors.Focus != DefaultThemeColors().Focus {
		t.Errorf("focus = %q, an invalid value must keep the default", cfg.Colors.Focus)
	}
}

func TestColoursRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.conf")
	cfg := Default(dir, dir)
	cfg.Colors.Selected = "#010203"
	if err := Save(cfg, path); err != nil {
		t.Fatal(err)
	}
	got, err := parseInto(path, Default(dir, dir))
	if err != nil {
		t.Fatal(err)
	}
	if got.Colors != cfg.Colors {
		t.Errorf("colors = %+v, want %+v", got.Colors, cfg.Colors)
	}
}

func TestEveryThemeIsOfferedWithItsOwnNameAndColours(t *testing.T) {
	names := ThemeNames()
	if names[0] != DefaultThemeName || names[len(names)-1] != CustomTheme {
		t.Errorf("names %v", names)
	}
	seen := map[string]bool{}
	for _, p := range ThemePresets {
		if seen[p.Name] {
			t.Errorf("%q twice", p.Name)
		}
		seen[p.Name] = true
		for _, c := range []string{p.Colors.Accent, p.Colors.Focus, p.Colors.Highlight, p.Colors.Success, p.Colors.Error, p.Colors.Dim, p.Colors.Text, p.Colors.Selected} {
			if !ValidColour(c) {
				t.Errorf("%s has the colour %q", p.Name, c)
			}
		}
		if p.Colors.Background != "" && !ValidColour(p.Colors.Background) {
			t.Errorf("%s has the background %q", p.Name, p.Colors.Background)
		}
	}
}

func TestATitleChangedAfterTheThemeWasPickedIsTheirOwn(t *testing.T) {
	cfg := Default("/app", "/music")
	cfg.ThemeName = "Dracula"
	colours, _ := PresetColors("Dracula")
	cfg.Colors = colours
	settleTheme(&cfg)
	if cfg.ThemeName != "Dracula" {
		t.Errorf("an untouched theme became %q", cfg.ThemeName)
	}
	cfg.Colors.Accent = "#000"
	settleTheme(&cfg)
	if cfg.ThemeName != CustomTheme {
		t.Errorf("a changed theme is %q", cfg.ThemeName)
	}
}
