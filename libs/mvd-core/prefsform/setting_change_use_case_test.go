package prefsform

import (
	"errors"
	"testing"

	"youtube-downloader/libs/mvd-core/config"
)

func newForm(caps Capabilities) Form {
	return New(config.Default("/app", "/music"), caps)
}

// onField puts the cursor on the setting with this label, whatever its tab.
func onField(t *testing.T, f *Form, label string) {
	t.Helper()
	for tab, p := range f.Pages {
		for i, field := range p.Fields {
			if field.Label == label {
				f.ShowTab(tab)
				f.SelectSetting(i)

				return
			}
		}
	}
	t.Fatalf("no setting called %q", label)
}

func TestTabsWrapAndKeepTheSettings(t *testing.T) {
	f := newForm(Capabilities{})
	f.ShowTab(-1)
	if f.Tab != len(f.Pages)-1 {
		t.Fatalf("tab -1 shows %d", f.Tab)
	}
	f.ShowTab(len(f.Pages))
	if f.Tab != 0 || f.Cursor != 0 {
		t.Errorf("tab past the end shows %d at %d", f.Tab, f.Cursor)
	}
}

func TestAStepperStaysBetweenItsLimits(t *testing.T) {
	f := newForm(Capabilities{})
	onField(t, &f, "Max Concurrent Downloads")
	for range 100 {
		f.Change(1)
	}
	if f.Cfg.MaxConcurrentDownloads != 32 {
		t.Errorf("went up to %d, want the limit 32", f.Cfg.MaxConcurrentDownloads)
	}
	for range 100 {
		f.Change(-1)
	}
	if f.Cfg.MaxConcurrentDownloads != 1 {
		t.Errorf("went down to %d, want the limit 1", f.Cfg.MaxConcurrentDownloads)
	}
}

func TestActivateOnAChoiceGoesToTheNextOptionAndWraps(t *testing.T) {
	f := newForm(Capabilities{})
	onField(t, &f, "Merge Format")
	seen := []string{f.Cfg.MergeOutputFormat}
	for range 3 {
		if effect := f.Activate(); effect != nil {
			t.Fatalf("a choice asked the host for %v", effect)
		}
		seen = append(seen, f.Cfg.MergeOutputFormat)
	}
	if seen[0] != seen[3] || seen[0] == seen[1] {
		t.Errorf("three options did not wrap: %v", seen)
	}
}

func TestAFolderAsksTheHostForItsChooserOrIsTypedWithoutOne(t *testing.T) {
	f := newForm(Capabilities{CanPickFolder: true})
	onField(t, &f, "Videos Output Folder")
	pick, ok := f.Activate().(PickFolder)
	if !ok || pick.Field != f.Cursor || pick.Start != f.Cfg.OutputDir || !f.Busy {
		t.Fatalf("asked for %+v, busy %v", pick, f.Busy)
	}
	f.Picked(pick.Field, "/new", true, nil)
	if f.Busy || f.Cfg.OutputDir != "/new" {
		t.Errorf("busy %v, folder %q", f.Busy, f.Cfg.OutputDir)
	}

	typed := newForm(Capabilities{})
	onField(t, &typed, "Videos Output Folder")
	if effect := typed.Activate(); effect != nil || !typed.Editing {
		t.Errorf("without a chooser: effect %v, editing %v", effect, typed.Editing)
	}
}

func TestAChooserThatCouldNotBeShownLeavesThePathToBeTyped(t *testing.T) {
	f := newForm(Capabilities{CanPickFolder: true})
	onField(t, &f, "Videos Output Folder")
	pick := f.Activate().(PickFolder)
	f.Picked(pick.Field, "", false, errors.New("no window"))
	if f.Busy || !f.Editing || f.Before != f.Cfg.OutputDir {
		t.Errorf("busy %v editing %v before %q", f.Busy, f.Editing, f.Before)
	}
}

func TestAnActionIsOfferedOnlyWhenTheHostCanDoItAndSaysWhatCameOfIt(t *testing.T) {
	hasAction := func(f Form, kind ActionKind) bool {
		for _, field := range f.Pages[2].Fields {
			if field.Kind == KindAction && field.Action == kind {
				return true
			}
		}

		return false
	}
	if f := newForm(Capabilities{}); hasAction(f, ActionMove) || hasAction(f, ActionUninstall) {
		t.Error("actions offered without a host that can do them")
	}
	f := newForm(Capabilities{CanMove: true, CanUninstall: true})
	onField(t, &f, "Uninstall MVD")
	if run, ok := f.Activate().(RunAction); !ok || run.Kind != ActionUninstall || !f.Busy {
		t.Fatalf("asked for %+v, busy %v", run, f.Busy)
	}
	f.ActionDone("", errors.New("could not"))
	if f.Busy || !f.StatusBad || f.Status != "could not" {
		t.Errorf("busy %v bad %v status %q", f.Busy, f.StatusBad, f.Status)
	}
}

func TestAColourIsOnlyTakenWhenItIsOneAndAnOptionalOneMayBeEmpty(t *testing.T) {
	f := newForm(Capabilities{})
	onField(t, &f, "Accent (title)")
	f.Edit()
	f.Accept("nonsense")
	if !f.Editing || !f.StatusBad {
		t.Fatalf("a bad colour: editing %v bad %v", f.Editing, f.StatusBad)
	}
	f.Accept("#f07")
	if f.Editing || f.Cfg.Colors.Accent != "#f07" || f.Cfg.ThemeName != config.CustomTheme {
		t.Errorf("editing %v accent %q theme %q", f.Editing, f.Cfg.Colors.Accent, f.Cfg.ThemeName)
	}

	onField(t, &f, "Background")
	f.Edit()
	f.Accept("")
	if f.Editing || f.Cfg.Colors.Background != "" {
		t.Errorf("an optional colour could not be emptied: %q", f.Cfg.Colors.Background)
	}
}

func TestUseDefaultColourPutsTheColourBack(t *testing.T) {
	f := newForm(Capabilities{})
	onField(t, &f, "Success")
	f.Edit()
	f.Accept("#123456")
	f.UseDefaultColour()
	if want := config.DefaultThemeColors().Success; f.Cfg.Colors.Success != want {
		t.Errorf("success is %q, want %q", f.Cfg.Colors.Success, want)
	}
}

func TestACopyOfTheFormKeepsItsOwnSettings(t *testing.T) {
	a := newForm(Capabilities{})
	onField(t, &a, "Max Concurrent Downloads")
	b := a
	b.Change(1)
	if a.Cfg.MaxConcurrentDownloads == b.Cfg.MaxConcurrentDownloads {
		t.Error("changing the copy changed the original")
	}
}
