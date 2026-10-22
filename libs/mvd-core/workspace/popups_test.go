package workspace

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/menubar"
)

func listOf(addresses ...string) Model {
	return sized(New(Host{Version: "1.2.3"}, config.Config{}, downloadlist.New(addresses)))
}

func click(m Model, x, y int) Model {
	next, _ := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	return next.(Model)
}

func TestTheCommandsOfTheEditOfTheListAreGreyedOutOutsideIt(t *testing.T) {
	undo := tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl}
	m := listOf("a", "b")
	if _, cmd := m.bar.Update(undo); cmd != nil {
		t.Error("Undo can be chosen with the edit of the list closed")
	}
	m = m.openEdit().refreshMenus()
	if _, cmd := m.bar.Update(undo); cmd == nil {
		t.Error("Undo cannot be chosen with the edit of the list open")
	}
}

func TestSaveAsHasItsOwnShortcutAndHelpHasNoOpenSource(t *testing.T) {
	var saveAs string
	var help []string
	for _, mn := range menus() {
		for _, e := range mn.Entries {
			if e.Name == entrySaveAs {
				saveAs = e.Shortcut
			}
			if mn.Name == "Help" {
				help = append(help, e.Name)
			}
		}
	}
	if saveAs != "Ctrl+Shift+S" {
		t.Errorf("Save as has %q", saveAs)
	}
	if len(help) != 1 || help[0] != entryAbout {
		t.Errorf("Help has %v", help)
	}
}

func TestTheAboutWindowTellsTheVersionAndTheLinkAndOpensItOnAClick(t *testing.T) {
	var opened string
	m := New(Host{Version: "1.2.3", Open: func(u string) error { opened = u; return nil }}, config.Config{}, downloadlist.New([]string{"a"}))
	m = sized(m)
	next, _ := m.Update(menubar.SelectedMsg{Name: entryAbout})
	m = next.(Model)
	screen := ansi.Strip(m.render())
	for _, want := range []string{"About MVD", "1.2.3", projectURL, "Remember the 90s?", "[ OK ]"} {
		if !strings.Contains(screen, want) {
			t.Errorf("the About window lacks %q", want)
		}
	}
	l := m.aboutLayout()
	x0, y0 := m.aboutOrigin()
	m = click(m, x0+1+l.link.X0+1, y0+1+l.linkY)
	if opened != projectURL {
		t.Errorf("the link opened %q", opened)
	}
	m = click(m, x0+1+l.button.X0+1, y0+1+l.buttonY)
	if m.about {
		t.Error("the OK button did not close the window")
	}
}

func TestThePencilOnTheListOpensItsEdit(t *testing.T) {
	m := listOf("https://youtu.be/a", "https://youtu.be/b")
	px, py, ok := m.pencil()
	if !ok {
		t.Fatal("no pencil")
	}
	m = click(m, px+1, py)
	if !m.editing {
		t.Error("a click on the pencil did not open the edit of the list")
	}
}

func TestTheEditOfTheListTakesTheMouse(t *testing.T) {
	m := listOf("https://youtu.be/a", "https://youtu.be/b", "https://youtu.be/c").openEdit()
	x0, y0 := m.editOrigin()
	w, _ := m.editSize()
	at := func(inside, row int) (int, int) { return x0 + 1 + inside, y0 + 1 + row }

	// Drag the first address down to the third place by its grip.
	x, y := at(1, 0)
	m = click(m, x, y)
	next, _ := m.Update(tea.MouseMotionMsg{X: x, Y: y + 2, Button: tea.MouseLeft})
	m = next.(Model)
	next, _ = m.Update(tea.MouseReleaseMsg{X: x, Y: y + 2, Button: tea.MouseLeft})
	m = next.(Model)
	if got := sources(m.edit); got[0] != "https://youtu.be/b" || got[2] != "https://youtu.be/a" {
		t.Errorf("after the drag %v", got)
	}

	// The cross on a row removes it.
	x, y = at(m.edit.removeX()+1, 0)
	m = click(m, x, y)
	if got := sources(m.edit); len(got) != 2 || got[0] != "https://youtu.be/c" {
		t.Errorf("after the cross %v", got)
	}

	// The pencil on a row asks for its address again.
	x, y = at(m.edit.editX()+1, 0)
	m = click(m, x, y)
	if m.edit.asking != "edit" {
		t.Errorf("asking %q", m.edit.asking)
	}
	m.edit.asking = ""

	// Undo is a button, and the close mark leaves without saving.
	_, spans := m.edit.buttonSpans()
	x, y = at(spans[buttonUndo].X0+1, m.edit.room())
	m = click(m, x, y)
	if got := sources(m.edit); len(got) != 3 {
		t.Errorf("after undo %v", got)
	}
	m = click(m, x0+w-4, y0)
	next, cmd := m.Update(tea.MouseClickMsg{X: x0 + w - 4, Y: y0, Button: tea.MouseLeft})
	_ = next
	if cmd == nil {
		t.Fatal("the close mark did nothing")
	}
	if _, ok := cmd().(editCancelledMsg); !ok {
		t.Errorf("the close mark sent %T", cmd())
	}
}
