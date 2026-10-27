package prefsform

import (
	"testing"

	"youtube-downloader/libs/mvd-core/config"
)

// themeFields are the fields of the Theme page for a host that can do what caps say.
func themeFields(caps Capabilities) []Field {
	for _, page := range Pages(caps) {
		if page.Name == "Theme" {
			return page.Fields
		}
	}

	return nil
}

func TestTheWindowStyleIsOfferedFirstOnlyWhereTheHostDrawsIt(t *testing.T) {
	if fields := themeFields(Capabilities{}); fields[0].Label == "Window style" {
		t.Error("the terminal app offers a window style it cannot draw")
	}
	field := themeFields(Capabilities{CanStyleWindow: true})[0]
	if field.Label != "Window style" || len(field.Options) != len(config.WindowStyles) {
		t.Fatalf("first field %q with %d options", field.Label, len(field.Options))
	}
	cfg := config.Config{}
	if field.Get(cfg) != config.WindowStyleDOS {
		t.Errorf("an empty style reads %q", field.Get(cfg))
	}
	field.Set(&cfg, "windows-xp")
	if cfg.WindowStyle != "windows-xp" || field.OptionText("windows-xp") != "Windows XP" {
		t.Errorf("style %q reads %q", cfg.WindowStyle, field.OptionText("windows-xp"))
	}
}
