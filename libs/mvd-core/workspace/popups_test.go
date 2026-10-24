package workspace

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/menubar"
	downloadpage "youtube-downloader/libs/mvd-core/run/download"
)

func listOf(addresses ...string) Model {
	return sized(New(Host{Version: "1.2.3"}, config.Config{}, downloadlist.New(addresses)))
}

func click(m Model, x, y int) Model {
	next, _ := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	return next.(Model)
}

func addressesOf(m Model) []string { return m.session.Sources() }

func TestSaveAsHasItsOwnShortcutAndHelpHasNoOpenSource(t *testing.T) {
	var saveAs string
	var help []string
	for _, mn := range menus() {
		for _, e := range mn.Entries {
			if e.Name == entrySaveAs {
				saveAs = e.Shortcut
			}
			if mn.Name == "Help" && !e.Separator {
				help = append(help, e.Name)
			}
		}
	}
	if saveAs != "Ctrl+Shift+S" {
		t.Errorf("Save as has %q", saveAs)
	}
	if want := []string{entryShortcuts, entryDocs, entryReport, entryAbout}; strings.Join(help, "|") != strings.Join(want, "|") {
		t.Errorf("Help has %v, want %v", help, want)
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

func TestTheListIsChangedInPlaceAndTheChangesCanBeUndone(t *testing.T) {
	m := listOf("https://a", "https://b", "https://c")
	next, _ := m.Update(downloadpage.ListEditMsg{Op: "move", Index: 0, To: 2})
	m = next.(Model)
	if got := addressesOf(m); got[0] != "https://b" || got[2] != "https://a" {
		t.Errorf("after the move %v", got)
	}
	next, _ = m.Update(downloadpage.ListEditMsg{Op: "remove", Index: 1})
	m = next.(Model)
	if got := addressesOf(m); len(got) != 2 || got[1] != "https://a" {
		t.Errorf("after the remove %v", got)
	}
	m, _ = choose(m, entryUndo)
	m, _ = choose(m, entryUndo)
	if got := addressesOf(m); len(got) != 3 || got[0] != "https://a" {
		t.Errorf("after two undos %v", got)
	}
	m, _ = choose(m, entryRedo)
	if got := addressesOf(m); got[0] != "https://b" {
		t.Errorf("after a redo %v", got)
	}
}

func typeInto(m Model, text string) Model {
	for _, r := range text {
		next, _ := m.Update(tea.KeyPressMsg{Text: string(r), Code: r})
		m = next.(Model)
	}
	return m
}

func TestChangingAnAddressIsDoneInItsRow(t *testing.T) {
	m := listOf("https://a", "https://b")
	next, _ := m.Update(downloadpage.ListEditMsg{Op: "edit", Index: 1})
	m = next.(Model)
	if !m.inline.on || m.inline.adding || m.inline.index != 1 {
		t.Fatalf("inline %+v", m.inline)
	}
	if !strings.Contains(ansi.Strip(m.render()), "https://b") {
		t.Error("the address is not shown in its row")
	}
	m = typeInto(m, "https://z") // the address is selected, so this replaces it
	m, _ = press(m, "enter")
	if got := addressesOf(m); got[1] != "https://z" || got[0] != "https://a" || m.inline.on {
		t.Errorf("after the change %v, inline %v", got, m.inline.on)
	}
}

func TestAddingALinkOpensARowAfterTheSelectedOneAndEscDropsIt(t *testing.T) {
	m := listOf("https://a", "https://b")
	m, _ = choose(m, entryAddLink)
	if !m.inline.on || !m.inline.adding || m.inline.index != 1 {
		t.Fatalf("inline %+v", m.inline)
	}
	m = typeInto(m, "https://new")
	m, _ = press(m, "enter")
	if got := addressesOf(m); len(got) != 3 || got[1] != "https://new" {
		t.Errorf("after the add %v", got)
	}
	m, _ = choose(m, entryAddLink)
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = next.(Model)
	if got := addressesOf(m); len(got) != 3 || m.inline.on {
		t.Errorf("after esc %v, inline %v", got, m.inline.on)
	}
}

func TestADoubleClickOnARowChangesItAndOnTheEmptyRoomAddsOne(t *testing.T) {
	m := listOf("https://a", "https://b")
	x, y, _, _, _, _ := m.screen.LeftColumn()
	row := y + 1 + topRows // the first playlist
	m = click(m, x+4, row)
	m = click(m, x+4, row)
	if !m.inline.on || m.inline.adding || m.inline.index != 0 {
		t.Fatalf("after a double click on a row: %+v", m.inline)
	}
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = next.(Model)
	empty := y + 8 + topRows // below the last playlist
	m = click(m, x+4, empty)
	m = click(m, x+4, empty)
	if !m.inline.on || !m.inline.adding {
		t.Errorf("after a double click on the empty room: %+v", m.inline)
	}
}

func TestCopyAndPasteOfTheEditMenuAreForTheAddressBeingWritten(t *testing.T) {
	copyKey := tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	_ = copyKey
	m := listOf("https://a")
	m = m.refreshMenus()
	if m.bar.Enabled(entryCopy) || m.bar.Enabled(entryCut) {
		t.Error("Copy and Cut are on with no address being written")
	}
	if !m.bar.Enabled(entryDelete) || m.bar.Enabled(entryMoveUp) || m.bar.Enabled(entryMoveDown) {
		t.Error("Delete works on the only playlist; there is nowhere to move it")
	}
	m, _ = m.startChange(0)
	m = m.refreshMenus()
	if !m.bar.Enabled(entryPaste) || !m.bar.Enabled(entryCopy) {
		t.Error("Paste and Copy (the address starts selected) are off while writing")
	}
	if m.bar.Enabled(entryDelete) {
		t.Error("Delete is on while writing")
	}
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	m = next.(Model).refreshMenus()
	if m.bar.Enabled(entryCopy) {
		t.Error("Copy is on with nothing selected")
	}
}

func TestDraggingOverTheAddressSelectsIt(t *testing.T) {
	m := listOf("https://abcdef")
	m, _ = m.startChange(0)
	x, y, _, ok := m.inlineRow()
	if !ok {
		t.Fatal("no row")
	}
	left := x + gripCells
	next, _ := m.Update(tea.MouseClickMsg{X: left + 8, Y: y, Button: tea.MouseLeft})
	next, _ = next.Update(tea.MouseMotionMsg{X: left + 11, Y: y, Button: tea.MouseLeft})
	next, _ = next.Update(tea.MouseReleaseMsg{X: left + 11, Y: y, Button: tea.MouseLeft})
	m = next.(Model)
	if got := m.inline.line.Selection(); got != "abc" {
		t.Errorf("selected %q", got)
	}
}

func TestPastedLinesBecomePlaylistsAfterTheSelectedOne(t *testing.T) {
	m := listOf("https://a", "https://z")
	next, _ := m.Update(tea.PasteMsg{Content: "https://b\n\n  https://c "})
	m = next.(Model)
	if got := addressesOf(m); len(got) != 4 || got[1] != "https://b" || got[2] != "https://c" || got[3] != "https://z" {
		t.Errorf("after the paste %v", got)
	}
}

func TestTheIconsOnARowChangeAndRemoveIt(t *testing.T) {
	m := listOf("https://a", "https://b")
	x, y, w, _, _, ok := m.screen.LeftColumn()
	if !ok {
		t.Fatal("no column of playlists")
	}
	// The second row: its cross is three cells from the right edge of the box, its pencil five.
	row := y + 2 + topRows
	next, cmd := m.Update(tea.MouseClickMsg{X: x + w - 5, Y: row, Button: tea.MouseLeft})
	m = next.(Model)
	if cmd == nil {
		t.Fatal("the pencil did nothing")
	}
	next, _ = m.Update(cmd())
	m = next.(Model)
	if !m.inline.on || m.inline.index != 1 {
		t.Errorf("inline %+v", m.inline)
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = next.(Model)
	next, cmd = m.Update(tea.MouseClickMsg{X: x + w - 3, Y: row, Button: tea.MouseLeft})
	m = next.(Model)
	next, _ = m.Update(cmd())
	m = next.(Model)
	if got := addressesOf(m); len(got) != 1 || got[0] != "https://a" {
		t.Errorf("after the cross %v", got)
	}
}

func TestTheButtonsStandUnderThePlaylistsAndLoadIsTheFirst(t *testing.T) {
	m := listOf("https://a")
	_, y, _, ok := m.buttonRows()
	if !ok {
		t.Fatal("no buttons")
	}
	screen := strings.Split(ansi.Strip(m.render()), "\n")
	if !strings.Contains(screen[y], entryStart) {
		t.Errorf("the first button row is %q", screen[y])
	}
}

type fakeClipboard struct{ text string }

func (c *fakeClipboard) Read() (string, error) { return c.text, nil }
func (c *fakeClipboard) Write(t string) error  { c.text = t; return nil }

func TestCopyCutAndPasteUseTheMachineClipboard(t *testing.T) {
	clip := &fakeClipboard{text: "https://from-the-clipboard"}
	m := New(Host{Clipboard: clip}, config.Config{}, downloadlist.New([]string{"https://a"}))
	m = sized(m)
	m, _ = m.startChange(0)
	// The address starts selected: copy keeps it, cut takes it out.
	next, _ := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	m = next.(Model)
	if clip.text != "https://a" {
		t.Errorf("copied %q", clip.text)
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: 'x', Mod: tea.ModCtrl})
	m = next.(Model)
	if m.inline.line.Value() != "" {
		t.Errorf("after the cut %q", m.inline.line.Value())
	}
	clip.text = "https://pasted"
	next, _ = m.Update(tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
	m = next.(Model)
	if m.inline.line.Value() != "https://pasted" {
		t.Errorf("after the paste %q", m.inline.line.Value())
	}
	// With nothing being written, Paste from the menu adds the links on the clipboard.
	m, _ = press(m, "enter")
	clip.text = "https://one" + string(rune(10)) + "https://two"
	m, _ = choose(m, entryPaste)
	if got := addressesOf(m); len(got) != 3 || got[1] != "https://one" {
		t.Errorf("after pasting links %v", got)
	}
}
