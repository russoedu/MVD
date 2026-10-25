package prefsform

import "testing"

func patternForm(t *testing.T) Form {
	t.Helper()
	f := newForm(Capabilities{})
	onField(t, &f, "Output Template")
	f.Current().Set(&f.Cfg, FormatTemplate(nil)) // an empty pattern to build on
	f.Edit()

	return f
}

func TestTheExampleFollowsTheTemplate(t *testing.T) {
	got := ExampleName("%(playlist_title,playlist)s/%(playlist_index)02d - %(title)s.%(ext)s", "mp4")
	if want := "Dance 90's/10 - Part Time Lover.mp4"; got != want {
		t.Errorf("example %q, want %q", got, want)
	}
	if got := ExampleName("%(nothing)s", "mp4"); got != "NA" {
		t.Errorf("an unknown field read %q", got)
	}
}

func TestEditingAPatternPutsTheCursorAtTheEnd(t *testing.T) {
	f := patternForm(t)
	if !f.Editing || f.Caret != f.TemplateLen() || f.Before != f.Cfg.OutputTemplate {
		t.Errorf("editing %v caret %d of %d", f.Editing, f.Caret, f.TemplateLen())
	}
}

func TestTypedTextGoesBetweenBlocksAndBackspaceTakesAWholeBlock(t *testing.T) {
	f := patternForm(t)
	f.TemplateCaretTo(0)
	f.AddBlock(1) // TITLE
	f.TemplateInsert(" - ")
	f.AddBlock(0) // ARTIST
	if got := f.Cfg.OutputTemplate; got != "%(title)s - %(artist,uploader)s.%(ext)s" {
		t.Fatalf("pattern is %q", got)
	}
	f.TemplateBackspace()
	if got := f.Cfg.OutputTemplate; got != "%(title)s - .%(ext)s" {
		t.Errorf("backspace left %q", got)
	}
}

func TestCancelPutsThePatternBackAndDoneKeepsIt(t *testing.T) {
	f := patternForm(t)
	before := f.Cfg.OutputTemplate
	f.TemplateInsert("x")
	f.TemplateCancel()
	if f.Editing || f.Cfg.OutputTemplate != before {
		t.Errorf("cancel left %q, editing %v", f.Cfg.OutputTemplate, f.Editing)
	}
	f.Edit()
	f.TemplateInsert("x")
	f.TemplateDone()
	if f.Editing || f.Cfg.OutputTemplate == before {
		t.Error("done should keep what was typed")
	}
}

func TestACrossRemovesABlockAndTheCursorStaysAfterTheOnesBeforeIt(t *testing.T) {
	f := patternForm(t)
	f.TemplateCaretTo(0)
	f.AddBlock(1)
	f.AddBlock(0)
	if f.Caret != 2 {
		t.Fatalf("caret %d", f.Caret)
	}
	f.RemoveBlock(0)
	if f.Caret != 1 || f.Cfg.OutputTemplate != "%(artist,uploader)s.%(ext)s" {
		t.Errorf("caret %d pattern %q", f.Caret, f.Cfg.OutputTemplate)
	}
}

func TestABlockCanBeCarriedToAnotherPlace(t *testing.T) {
	f := patternForm(t)
	f.TemplateCaretTo(0)
	f.AddBlock(1)
	f.AddBlock(0)
	f.MoveBlock(0, 2)
	if got := f.Cfg.OutputTemplate; got != "%(artist,uploader)s%(title)s.%(ext)s" {
		t.Errorf("pattern is %q", got)
	}
}

func TestAPaletteBlockStartsTypingWhenItWasNotAndPlaceCaretDoesToo(t *testing.T) {
	f := newForm(Capabilities{})
	onField(t, &f, "Output Template")
	f.AddBlock(1)
	if !f.Editing {
		t.Error("adding a block should start typing")
	}
	f.CancelEdit()
	f.PlaceCaret(0)
	if !f.Editing || f.Caret != 0 {
		t.Errorf("editing %v caret %d", f.Editing, f.Caret)
	}
	f.AddBlock(99)
	f.AddBlock(-1)
}
