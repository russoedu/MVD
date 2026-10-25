package linetext

import "testing"

func TestTypingReplacesWhatIsSelectedAndDeleteTakesOneCharacter(t *testing.T) {
	x := New("hello")
	if !x.HasSelection() || x.Selection() != "hello" {
		t.Fatal("a new text should have all of it selected")
	}
	x = x.Insert("a")
	if x.Value() != "a" || x.HasSelection() {
		t.Fatalf("typing over the selection gave %q", x.Value())
	}
	x = x.Insert("bc").Delete(true)
	if x.Value() != "ab" {
		t.Errorf("backspace gave %q", x.Value())
	}
	x = x.MoveTo(0, false).Delete(false)
	if x.Value() != "b" {
		t.Errorf("delete gave %q", x.Value())
	}
}

func TestStepEndsASelectionAtItsNearEdgeUnlessExtending(t *testing.T) {
	x := New("abcd").MoveTo(1, false).MoveTo(3, true)
	if x.Selection() != "bc" {
		t.Fatalf("selection is %q", x.Selection())
	}
	if got := x.Step(-1, false); got.Caret() != 1 || got.HasSelection() {
		t.Errorf("left ended the selection at %d", got.Caret())
	}
	if got := x.Step(1, false); got.Caret() != 3 || got.HasSelection() {
		t.Errorf("right ended the selection at %d", got.Caret())
	}
	if got := x.Step(1, true); got.Selection() != "bcd" {
		t.Errorf("shift+right gave %q", got.Selection())
	}
}

func TestPasteMakesOneLineAndCopiesAreIndependent(t *testing.T) {
	x := New("").Paste("a\nb\tc")
	if x.Value() != "a b c" {
		t.Errorf("paste gave %q", x.Value())
	}
	y := x.Insert("z")
	if x.Value() != "a b c" || y.Value() == x.Value() {
		t.Error("a change altered the text it came from")
	}
}
