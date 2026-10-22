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

func TestChangingAnAddressAsksForTheNewOne(t *testing.T) {
	m := listOf("https://a", "https://b")
	next, _ := m.Update(downloadpage.ListEditMsg{Op: "edit", Index: 1})
	m = next.(Model)
	if m.asking != askChangeLink {
		t.Fatalf("asking %v", m.asking)
	}
	m.input.SetValue("https://z")
	m, _ = press(m, "enter")
	if got := addressesOf(m); got[1] != "https://z" || got[0] != "https://a" {
		t.Errorf("after the change %v", got)
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
	if m.asking != askChangeLink || m.changing != 1 {
		t.Errorf("asking %v, changing %d", m.asking, m.changing)
	}
	m.asking = askNone
	next, cmd = m.Update(tea.MouseClickMsg{X: x + w - 3, Y: row, Button: tea.MouseLeft})
	m = next.(Model)
	next, _ = m.Update(cmd())
	m = next.(Model)
	if got := addressesOf(m); len(got) != 1 || got[0] != "https://a" {
		t.Errorf("after the cross %v", got)
	}
}

func TestTheButtonsStandUnderThePlaylistsAndStartIsTheFirst(t *testing.T) {
	m := listOf("https://a")
	x, y, w, ok := m.buttonRows()
	if !ok {
		t.Fatal("no buttons")
	}
	screen := strings.Split(ansi.Strip(m.render()), "\n")
	if !strings.Contains(screen[y], entryStart) {
		t.Errorf("the first button row is %q", screen[y])
	}
	m = click(m, x+w/2, y+1) // Get data for review
	_ = m
}
