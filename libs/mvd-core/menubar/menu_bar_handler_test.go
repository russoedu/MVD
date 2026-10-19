package menubar

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func sample() Model {
	return New(
		WithMenus(
			Menu{Name: "File", Entries: []Entry{
				{Name: "New", Shortcut: "Ctrl+N"},
				{Name: "Open…", Shortcut: "Ctrl+O"},
				{Separator: true},
				{Name: "Quit", Shortcut: "Ctrl+Q"},
			}},
			Menu{Name: "Edit", Entries: []Entry{{Name: "Undo", Shortcut: "Ctrl+Z"}, {Name: "Redo"}}},
			Menu{Name: "Help", Entries: []Entry{{Name: "About"}}},
		),
		WithSize(40, 8),
	)
}

func click(x, y int) tea.MouseClickMsg {
	return tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft}
}

func press(name string) tea.KeyPressMsg {
	switch name {
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "f10":
		return tea.KeyPressMsg{Code: tea.KeyF10}
	case "alt+e":
		return tea.KeyPressMsg{Code: 'e', Mod: tea.ModAlt}
	case "ctrl+q":
		return tea.KeyPressMsg{Code: 'q', Mod: tea.ModCtrl}
	}
	panic("unknown key " + name)
}

func selected(t *testing.T, cmd tea.Cmd) SelectedMsg {
	t.Helper()
	if cmd == nil {
		t.Fatal("no message")
	}
	msg, ok := cmd().(SelectedMsg)
	if !ok {
		t.Fatalf("message is %T", cmd())
	}
	return msg
}

func TestTheBarShowsTheHeadingsOnTheFirstRow(t *testing.T) {
	rows := strings.Split(ansi.Strip(sample().View()), "\n")
	if len(rows) != 8 {
		t.Fatalf("%d rows", len(rows))
	}
	if !strings.Contains(rows[0], "File") || !strings.Contains(rows[0], "Edit") || !strings.Contains(rows[0], "Help") {
		t.Errorf("bar = %q", rows[0])
	}
	for i, r := range rows {
		if len([]rune(r)) != 40 {
			t.Errorf("row %d is %d cells", i, len([]rune(r)))
		}
	}
}

func TestClickingAHeadingOpensItsBoxAndAnEntryChoosesIt(t *testing.T) {
	m := sample()
	m, _ = m.Update(click(m.headingX()[0]+1, 0))
	if m.Open() != 0 {
		t.Fatalf("open = %d", m.Open())
	}
	view := ansi.Strip(m.View())
	for _, want := range []string{"New", "Ctrl+N", "Open…", "Quit", "╭", "╰"} {
		if !strings.Contains(view, want) {
			t.Errorf("box lacks %q:\n%s", want, view)
		}
	}
	// Row 0 is the bar, row 1 the top edge, so Quit (entry 3) is on row 5.
	x := m.boxX() + 3
	m, cmd := m.Update(click(x, 5))
	got := selected(t, cmd)
	if got.Menu != 0 || got.Entry != 3 || got.Name != "Quit" || got.Shortcut != "Ctrl+Q" || got.MenuName != "File" {
		t.Errorf("chose %+v", got)
	}
	if m.Open() != -1 {
		t.Error("the box stays open after a choice")
	}
}

func TestASeparatorCannotBeChosen(t *testing.T) {
	m := sample()
	m, _ = m.Update(click(m.headingX()[0]+1, 0))
	// The separator is entry 2, on row 4.
	m, cmd := m.Update(click(m.boxX()+3, 4))
	if cmd != nil {
		t.Error("a separator was chosen")
	}
	if m.Open() != -1 {
		t.Error("a click that chose nothing leaves the box open")
	}
}

func TestClickingOutsideClosesWithoutChoosing(t *testing.T) {
	m := sample()
	m, _ = m.Update(click(m.headingX()[1]+1, 0))
	m, cmd := m.Update(click(35, 7))
	if cmd != nil || m.Open() != -1 {
		t.Errorf("open %d cmd %v", m.Open(), cmd != nil)
	}
}

func TestMovingOverAnotherHeadingSwitchesMenu(t *testing.T) {
	m := sample()
	m, _ = m.Update(click(m.headingX()[0]+1, 0))
	m, _ = m.Update(tea.MouseMotionMsg{X: m.headingX()[2] + 1, Y: 0})
	if m.Open() != 2 {
		t.Errorf("open = %d", m.Open())
	}
}

func TestKeysOpenMoveAndChoose(t *testing.T) {
	m := sample()
	// Arrow keys belong to the program until the bar is taken.
	m, _ = m.Update(press("down"))
	if m.Open() != -1 {
		t.Fatal("down opened a menu that was never asked for")
	}
	m, _ = m.Update(press("f10"))
	if m.Open() != 0 {
		t.Fatalf("F10 opened %d", m.Open())
	}
	m, _ = m.Update(press("right"))
	m, _ = m.Update(press("down"))
	m, cmd := m.Update(press("enter"))
	got := selected(t, cmd)
	if got.MenuName != "Edit" || got.Name != "Redo" {
		t.Errorf("chose %+v (the first entry is highlighted on opening, down moves to Redo)", got)
	}
	if m.Open() != -1 {
		t.Error("still open")
	}
}

func TestTheArrowsSkipSeparators(t *testing.T) {
	m := sample()
	m.SetOpen(0)
	m, _ = m.Update(press("f10"))
	for range 2 {
		m, _ = m.Update(press("down"))
	}
	if m.cursor != 3 {
		t.Errorf("cursor = %d, want Quit (3) after skipping the separator", m.cursor)
	}
	m, _ = m.Update(press("down"))
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want it to wrap to New", m.cursor)
	}
}

func TestAltAndTheFirstLetterOpensAMenuAndEscapeCloses(t *testing.T) {
	m := sample()
	m, _ = m.Update(press("alt+e"))
	if m.Open() != 1 {
		t.Fatalf("open = %d", m.Open())
	}
	m, _ = m.Update(press("esc"))
	if m.Open() != -1 {
		t.Error("escape did not close the box")
	}
}

func TestAShortcutChoosesItsEntryWithTheBarClosed(t *testing.T) {
	m := sample()
	_, cmd := m.Update(press("ctrl+q"))
	got := selected(t, cmd)
	if got.Name != "Quit" || got.Menu != 0 || got.Entry != 3 {
		t.Errorf("chose %+v", got)
	}
}

func TestADesignBuildsItFromScalarsOnly(t *testing.T) {
	text := `[{"name":"File","entries":[{"name":"New","shortcut":"Ctrl+N"},{"separator":true}]},{"name":"Help","entries":[]}]`
	m := New(WithMenusJSON(text), WithTitle("App"), WithOpen(0), WithSize(30, 6))
	if len(m.Menus()) != 2 || m.Open() != 0 {
		t.Fatalf("menus %d open %d", len(m.Menus()), m.Open())
	}
	if !strings.Contains(ansi.Strip(m.View()), "App") {
		t.Error("title not drawn")
	}
	if got := New(WithMenusJSON("not json")); len(got.Menus()) != 0 {
		t.Error("text that does not parse leaves menus")
	}
	if New(WithMenusJSON(text), WithOpen(5)).Open() != -1 {
		t.Error("a menu that does not exist cannot be open")
	}
	if round := m.MenusJSON(); round != text {
		t.Errorf("round trip:\n%s\n%s", round, text)
	}
}

func TestTheBoxStaysInsideTheComponent(t *testing.T) {
	m := New(WithMenusJSON(`[{"name":"Aaaaaaaaaaaaaaaaaaaa","entries":[{"name":"x"}]},{"name":"Last","entries":[{"name":"Quit","shortcut":"Ctrl+Q"}]}]`), WithOpen(1), WithSize(30, 6))
	for i, r := range strings.Split(ansi.Strip(m.View()), "\n") {
		if len([]rune(r)) != 30 {
			t.Errorf("row %d is %d cells wide", i, len([]rune(r)))
		}
	}
	if m.boxX()+m.boxWidth()+2 > 30 {
		t.Errorf("the box reaches column %d", m.boxX()+m.boxWidth()+2)
	}
}

func TestContentHeightIsTheBarAndTheTallestBox(t *testing.T) {
	if got := sample().ContentHeight(); got != 1+4+2 {
		t.Errorf("height = %d", got)
	}
	if got := New().ContentHeight(); got != 1 {
		t.Errorf("empty bar height = %d", got)
	}
}

func TestAGreyedEntryChoosesNothingAndIsSkipped(t *testing.T) {
	m := sample()
	m.SetEnabled("Open…", false)

	// Its shortcut does nothing.
	if _, cmd := m.Update(tea.KeyPressMsg{Code: 'o', Mod: tea.ModCtrl}); cmd != nil {
		t.Error("the shortcut of a greyed entry must not choose it")
	}
	// Clicking it does nothing, and the box stays as it was.
	m, _ = m.Update(click(2, 0))
	if m.Open() != 0 {
		t.Fatalf("menu %d open, want the first", m.Open())
	}
	if _, cmd := m.Update(click(4, 3)); cmd != nil {
		t.Error("a click on a greyed entry must not choose it")
	}
	// The arrow keys step over it: New, then Quit.
	m, _ = m.Update(press("down"))
	m, _ = m.Update(press("down"))
	m, cmd := m.Update(press("enter"))
	if cmd == nil {
		t.Fatal("enter chose nothing")
	}
	if got := cmd().(SelectedMsg).Name; got != "Quit" {
		t.Errorf("chose %q, want Quit (Open… is greyed and the separator is skipped)", got)
	}
}
