package prefsform

import (
	"testing"

	"youtube-downloader/libs/mvd-core/config"
)

// formOn is the form on the tab named name, for a host that can do what caps say.
func formOn(t *testing.T, name string, caps Capabilities) Form {
	t.Helper()
	f := New(config.Default(t.TempDir(), t.TempDir()), caps)
	for i, page := range f.Pages {
		if page.Name == name {
			f.Apply(WindowIntent{Kind: IntentTab, Index: i})

			return f
		}
	}
	t.Fatalf("no tab %q", name)

	return f
}

// fieldIndex is where the setting labelled label is on the tab shown.
func fieldIndex(t *testing.T, f Form, label string) int {
	t.Helper()
	for i, field := range f.Window().Fields {
		if field.Label == label {
			return i
		}
	}
	t.Fatalf("no setting %q", label)

	return -1
}

func TestTheWindowShowsTheTabsAndTheSettingsOfTheOneShown(t *testing.T) {
	f := formOn(t, "Theme", Capabilities{CanStyleWindow: true})
	w := f.Window()
	if w.Title != "Preferences" || len(w.Tabs) != 4 || w.Tabs[w.Tab] != "Theme" {
		t.Fatalf("tabs %v, tab %d", w.Tabs, w.Tab)
	}
	style := w.Fields[0]
	if style.Kind != WindowChoice || style.Value != config.WindowStyleDOS || style.Options[1].Label != "Windows 3.1" {
		t.Fatalf("window style = %+v", style)
	}
	theme := w.Fields[1]
	if theme.Kind != WindowChoice || len(theme.Options[0].Swatches) == 0 {
		t.Errorf("the theme's options have no colours: %+v", theme.Options[0])
	}
	if w.Colours.Accent == "" {
		t.Error("the window does not carry the terminal look's colours")
	}
}

func TestAChoiceANumberAndATextAreSetFromTheWindow(t *testing.T) {
	f := formOn(t, "Theme", Capabilities{CanStyleWindow: true})
	f.Apply(WindowIntent{Kind: IntentSet, Field: 0, Text: "mac-os-9"})
	if f.Cfg.WindowStyle != "mac-os-9" {
		t.Errorf("style %q", f.Cfg.WindowStyle)
	}

	f = formOn(t, "Download", Capabilities{})
	at := fieldIndex(t, f, "Retry Attempts")
	f.Apply(WindowIntent{Kind: IntentSet, Field: at, Text: "9999"})
	if got := f.Window().Fields[at].Value; got != intText(f.Pages[f.Tab].Fields[at].Max) {
		t.Errorf("a number past its bound was kept as %s", got)
	}
}

func TestAWrongColourIsRefusedWithTheReasonAndADefaultOnePutBack(t *testing.T) {
	f := formOn(t, "Theme", Capabilities{})
	at := fieldIndex(t, f, "Accent (title)")
	f.Apply(WindowIntent{Kind: IntentSet, Field: at, Text: "pink"})
	if w := f.Window(); !w.Bad || w.Status == "" {
		t.Errorf("a wrong colour was not refused: %+v", w)
	}
	f.Apply(WindowIntent{Kind: IntentSet, Field: at, Text: "#123456"})
	if f.Cfg.Colors.Accent != "#123456" {
		t.Fatalf("accent %q", f.Cfg.Colors.Accent)
	}
	f.Apply(WindowIntent{Kind: IntentDefaultColour, Field: at})
	if f.Cfg.Colors.Accent != config.DefaultThemeColors().Accent {
		t.Errorf("accent %q after default", f.Cfg.Colors.Accent)
	}
}

func TestTheNamePatternIsShownAsBlocksAndChangedFromTheWindow(t *testing.T) {
	f := formOn(t, "Output", Capabilities{})
	at := fieldIndex(t, f, "Output Template")
	pattern := f.Window().Fields[at]
	if len(pattern.Blocks) != len(TemplateBlocks) || pattern.Example == "" || !pattern.Units[0].Block {
		t.Fatalf("pattern = %+v", pattern)
	}
	before := len(pattern.Units)
	f.Apply(WindowIntent{Kind: IntentAddText, Field: at, Text: " - "})
	f.Apply(WindowIntent{Kind: IntentAddBlock, Field: at, Index: 0})
	units := f.Window().Fields[at].Units
	if len(units) != before+4 || units[len(units)-1].Text != "ARTIST" || f.Editing {
		t.Fatalf("units after adding = %+v, editing %v", units, f.Editing)
	}
	f.Apply(WindowIntent{Kind: IntentRemoveUnit, Field: at, Index: len(units) - 1})
	if got := len(f.Window().Fields[at].Units); got != before+3 {
		t.Errorf("%d units after removing the last", got)
	}
}

func TestAFolderAsksTheHostForItsChooser(t *testing.T) {
	f := formOn(t, "Output", Capabilities{CanPickFolder: true})
	effect := f.Apply(WindowIntent{Kind: IntentActivate, Field: fieldIndex(t, f, "Videos Output Folder")})
	if _, ok := effect.(PickFolder); !ok || !f.Window().Busy {
		t.Fatalf("effect %T, busy %v", effect, f.Busy)
	}
	if f.Apply(WindowIntent{Kind: IntentTab, Index: 0}) != nil || f.Tab != 0 {
		t.Error("the form did something while the host's window was open")
	}
}
