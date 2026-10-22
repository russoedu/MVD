package menubar

import (
	"slices"
	"testing"
)

func TestShortcutKeysUseBubbleTeaSpelling(t *testing.T) {
	cases := map[string][]string{
		"Ctrl+S":                 {"ctrl+s"},
		"Ctrl+Shift+S":           {"ctrl+shift+s"},
		"Ctrl+Y / Ctrl+Shift+Z":  {"ctrl+y", "ctrl+shift+z"},
		"Del":                    {"delete"},
		"F5":                     {"f5"},
		"":                       nil,
		"Ctrl + Alt + Backspace": {"ctrl+alt+backspace"},
	}
	for text, want := range cases {
		if got := shortcutKeys(text); !slices.Equal(got, want) {
			t.Errorf("shortcutKeys(%q) = %v, want %v", text, got, want)
		}
	}
}

func TestAMacShowsCmdWhereOthersShowCtrl(t *testing.T) {
	defer func(was string) { hostOS = was }(hostOS)

	hostOS = "windows"
	if got := ShortcutLabel("Ctrl+Shift+S"); got != "Ctrl+Shift+S" {
		t.Errorf("windows read %q", got)
	}
	hostOS = "darwin"
	if got := ShortcutLabel("Ctrl+Shift+S"); got != "Cmd+Shift+S" {
		t.Errorf("a mac read %q", got)
	}
	if got := ShortcutLabel("Del"); got != "Del" {
		t.Errorf("a mac read %q", got)
	}
}
