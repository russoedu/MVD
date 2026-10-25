package menumodel

import "testing"

func sample() Bar {
	return NewBar().WithMenus([]Menu{
		{Name: "File", Entries: []Entry{
			{Name: "New", Shortcut: "Ctrl+N"},
			{Name: "Open", Shortcut: "Ctrl+O"},
			{Separator: true},
			{Name: "Quit", Shortcut: "Ctrl+Q"},
		}},
		{Name: "Edit", Entries: []Entry{{Name: "Undo", Shortcut: "Ctrl+Z"}, {Name: "Redo"}}},
	}, None)
}

func TestClickingAHeadingOpensItAndAgainClosesIt(t *testing.T) {
	b := sample()
	b.ClickHeading(0)
	if b.Open() != 0 || !b.Active() || b.Cursor() != None {
		t.Fatalf("open %d active %v cursor %d", b.Open(), b.Active(), b.Cursor())
	}
	b.ClickHeading(0)
	if b.Open() != None || b.Active() {
		t.Error("a second click should close the box")
	}
}

func TestChoosingAnEntryClosesTheBarAndSaysWhatWasChosen(t *testing.T) {
	b := sample()
	b.ClickHeading(0)
	sel, ok := b.ClickEntry(3)
	if !ok || sel.Name != "Quit" || sel.MenuName != "File" || sel.Shortcut != "Ctrl+Q" || sel.Menu != 0 || sel.Entry != 3 {
		t.Fatalf("chose %+v (%v)", sel, ok)
	}
	if b.Open() != None || b.Active() {
		t.Error("choosing should close the bar")
	}
}

func TestTheKeyboardSkipsSeparatorsAndWraps(t *testing.T) {
	b := sample()
	b.Activate()
	if b.Open() != 0 || b.Cursor() != 0 {
		t.Fatalf("F10 opened %d on %d", b.Open(), b.Cursor())
	}
	b.MoveEntry(1)
	b.MoveEntry(1)
	if b.Cursor() != 3 {
		t.Errorf("cursor %d, want Quit (3) after skipping the separator", b.Cursor())
	}
	b.MoveEntry(1)
	if b.Cursor() != 0 {
		t.Errorf("cursor %d, want it to wrap to New", b.Cursor())
	}
	b.Last()
	if b.Cursor() != 3 {
		t.Errorf("End went to %d", b.Cursor())
	}
}

func TestTheArrowsAreLeftToTheProgramUntilTheBarIsTaken(t *testing.T) {
	b := sample()
	b.MoveHeading(1)
	b.MoveEntry(1)
	if b.Open() != None || b.MenuCursor() != 0 || b.Engaged() {
		t.Error("the bar moved without being taken")
	}
	b.OpenByLetter("e")
	if b.Open() != 1 || !b.Engaged() {
		t.Errorf("alt+e opened %d", b.Open())
	}
	b.Escape()
	if b.Open() != None || !b.Active() {
		t.Error("the first Escape closes the box and keeps the bar")
	}
	b.Escape()
	if b.Active() {
		t.Error("the second Escape gives the keyboard back")
	}
}

func TestAShortcutChoosesWithTheBarClosedAndAGreyedEntryChoosesNothing(t *testing.T) {
	b := sample()
	if sel, ok := b.ChooseShortcut("ctrl+o"); !ok || sel.Name != "Open" {
		t.Fatalf("ctrl+o chose %+v (%v)", sel, ok)
	}
	b.SetEnabled("Open", false)
	if _, ok := b.ChooseShortcut("ctrl+o"); ok {
		t.Error("a greyed entry was chosen by its shortcut")
	}
	b.SetOpen(0)
	if b.Cursor() != 0 {
		t.Fatalf("cursor %d", b.Cursor())
	}
	b.MoveEntry(1)
	if b.Cursor() != 3 {
		t.Errorf("the arrows stopped on a greyed entry: %d", b.Cursor())
	}
}

func TestThePointerOpensAnotherHeadingWhileABoxIsOpen(t *testing.T) {
	b := sample()
	b.PointerOver(1, None)
	if b.Hover() != 1 || b.Open() != None {
		t.Fatalf("hover %d open %d", b.Hover(), b.Open())
	}
	b.ClickHeading(0)
	b.PointerOver(1, None)
	if b.Open() != 1 || b.MenuCursor() != 1 || b.Hover() != None {
		t.Errorf("open %d menu cursor %d hover %d", b.Open(), b.MenuCursor(), b.Hover())
	}
}

func TestACopyOfTheBarKeepsItsOwnGreyedEntries(t *testing.T) {
	a := sample()
	b := a
	b.SetEnabled("Open", false)
	if !a.Enabled("Open") || b.Enabled("Open") {
		t.Error("greying an entry on a copy changed the original")
	}
}
