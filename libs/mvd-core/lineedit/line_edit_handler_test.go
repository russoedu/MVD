package lineedit

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func typed(m Model, s string) Model {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Text: string(r), Code: r})
	}
	return m
}

func key(m Model, k tea.KeyPressMsg) Model {
	m, _ = m.Update(k)
	return m
}

func TestTypingReplacesWhatIsSelectedAndBackspaceTakesOneCharacter(t *testing.T) {
	m := New("old", 20)
	if m.Selection() != "old" {
		t.Fatalf("selection %q", m.Selection())
	}
	m = typed(m, "new!")
	if m.Value() != "new!" {
		t.Errorf("value %q", m.Value())
	}
	m = key(m, tea.KeyPressMsg{Code: tea.KeyBackspace})
	if m.Value() != "new" {
		t.Errorf("after backspace %q", m.Value())
	}
}

func TestShiftArrowsAndTheMouseSelect(t *testing.T) {
	m := New("hello world", 20)
	m = key(m, tea.KeyPressMsg{Code: tea.KeyHome})
	for i := 0; i < 5; i++ {
		m = key(m, tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModShift})
	}
	if m.Selection() != "hello" {
		t.Errorf("selection %q", m.Selection())
	}
	// Press at 6 and drag to 11 with the mouse.
	m, _ = m.Update(tea.MouseClickMsg{X: 6, Button: tea.MouseLeft})
	m, _ = m.Update(tea.MouseMotionMsg{X: 11, Button: tea.MouseLeft})
	m, _ = m.Update(tea.MouseReleaseMsg{X: 11, Button: tea.MouseLeft})
	if m.Selection() != "world" {
		t.Errorf("dragged selection %q", m.Selection())
	}
	if m.selecting {
		t.Error("still selecting after the release")
	}
}

func TestCutCopyAndPaste(t *testing.T) {
	m := New("abc def", 20)
	m, cmd := m.Cut()
	if m.Value() != "" || cmd == nil {
		t.Errorf("after the cut %q, cmd %v", m.Value(), cmd != nil)
	}
	m = m.Paste("one\ntwo")
	if m.Value() != "one two" {
		t.Errorf("after the paste %q", m.Value())
	}
	if m.Copy() != nil {
		t.Error("nothing is selected, so there is nothing to copy")
	}
}

func TestTheViewIsAsWideAsTheEditorAndFollowsTheCursor(t *testing.T) {
	m := New("a long address that does not fit in the room", 12)
	if w := ansi.StringWidth(m.View()); w != 12 {
		t.Errorf("view is %d wide", w)
	}
	m = key(m, tea.KeyPressMsg{Code: tea.KeyHome})
	if got := ansi.Strip(m.View()); got != "a long addre" {
		t.Errorf("at the start %q", got)
	}
	m = key(m, tea.KeyPressMsg{Code: tea.KeyEnd})
	if got := ansi.Strip(m.View()); len(got) != 12 || got[len(got)-1] != ' ' {
		t.Errorf("at the end %q", got)
	}
}
