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
	if fields := themeFields(Capabilities{}); fields[0].Label == "Look" {
		t.Error("the terminal app offers a window style it cannot draw")
	}
	field := themeFields(Capabilities{CanStyleWindow: true})[0]
	if field.Label != "Look" || len(field.Options) != len(config.WindowStyles) {
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

func TestSignInIsOnlyOfferedWhereTheHostCanDoIt(t *testing.T) {
	has := func(caps Capabilities) (action, option bool) {
		for _, page := range Pages(caps) {
			for _, f := range page.Fields {
				if f.Kind == KindAction && f.Action == ActionSignIn {
					action = true
				}
				for _, o := range f.Options {
					if o == "mvd" {
						option = true
					}
				}
			}
		}

		return action, option
	}
	if a, o := has(Capabilities{}); a || o {
		t.Errorf("without the capability: action %v, option %v, want neither", a, o)
	}
	if a, o := has(Capabilities{CanSignIn: true}); !a || !o {
		t.Errorf("with the capability: action %v, option %v, want both", a, o)
	}
}
