package prefspopup

import (
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/look"
)

func newPopup() Model {
	return New(config.Default("/app", "/music"), Host{}).SetSize(100, 30)
}

func press(m Model, key string) Model {
	code := rune(0)
	switch key {
	case "up":
		code = tea.KeyUp
	case "down":
		code = tea.KeyDown
	case "left":
		code = tea.KeyLeft
	case "right":
		code = tea.KeyRight
	case "tab":
		code = tea.KeyTab
	case "enter":
		code = tea.KeyEnter
	case "esc":
		code = tea.KeyEscape
	default:
		code = []rune(key)[0]
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: code})
	return m
}

func click(m Model, x, y int) Model {
	m, _ = m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	return m
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

func TestTabsShowTheirPagesAndKeepTheSettings(t *testing.T) {
	m := newPopup()
	if m = press(m, "tab"); m.Tab() != 1 {
		t.Fatalf("tab = %d", m.Tab())
	}
	m = press(m, "right") // the video quality, one option on
	changed := m.Config().VideoQuality
	if changed == newPopup().Config().VideoQuality {
		t.Fatal("the option did not change")
	}
	m = press(press(m, "tab"), "tab")
	m = press(press(m, "tab"), "tab") // all the way round
	if m.Tab() != 1 || m.Config().VideoQuality != changed {
		t.Errorf("back on tab %d with %q", m.Tab(), m.Config().VideoQuality)
	}
}

func TestAStepperStaysBetweenItsLimits(t *testing.T) {
	m := newPopup().show(1)
	for range 4 {
		m = press(m, "down") // to Max Concurrent Downloads
	}
	for range 50 {
		m = press(m, "right")
	}
	if got := m.Config().MaxConcurrentDownloads; got != 32 {
		t.Errorf("up to %d, want 32", got)
	}
	for range 50 {
		m = press(m, "left")
	}
	if got := m.Config().MaxConcurrentDownloads; got != 1 {
		t.Errorf("down to %d, want 1", got)
	}
}

func TestAColourIsOnlyTakenWhenItIsOne(t *testing.T) {
	m := newPopup().show(3)
	m = press(press(m, "down"), "down") // past the theme and the background: the accent
	m = press(m, "enter")
	m.input.SetValue("pink")
	if m = press(m, "enter"); !m.Editing() || m.status == "" {
		t.Fatal("a bad colour was accepted")
	}
	m.input.SetValue("#f07")
	m = press(m, "enter")
	if m.Editing() || m.Config().Colors.Accent != "#f07" {
		t.Errorf("accent %q, editing %v", m.Config().Colors.Accent, m.Editing())
	}
}

func TestClicksOnTabsOptionsStepsAndBlocks(t *testing.T) {
	m := newPopup()
	// The second tab, in the row of tabs.
	tabs := m.tabSpans()
	if m = click(m, tabs[1].first+1, 1); m.Tab() != 1 {
		t.Fatalf("a click on the Download tab shows tab %d", m.Tab())
	}
	l := m.layout()
	for _, h := range l.hits {
		if h.kind == hitOption && h.field == 0 && h.arg == 2 {
			m = click(m, h.x0+2, tabRows+h.y-m.scroll)
		}
	}
	if got := m.Config().VideoQuality; got != config.VideoPresets[2] {
		t.Errorf("a click on an option set %q", got)
	}
	for _, h := range l.hits {
		if h.kind == hitPlus && m.pages[m.tab].fields[h.field].label == "Max Concurrent Downloads" {
			before := m.Config().MaxConcurrentDownloads
			m = m.reveal()
			m.scroll = max(h.y-3, 0) // the page is longer than the room: bring the row into it
			m = click(m, h.x0+1, tabRows+h.y-m.scroll)
			if got := m.Config().MaxConcurrentDownloads; got != before+1 {
				t.Errorf("a click on + gave %d, was %d", got, before)
			}
		}
	}
	m = m.show(0)
	m.cursor = 1
	for _, h := range m.layout().hits {
		if h.kind == hitBlock && h.arg == 1 {
			m = click(m, h.x0+2, tabRows+h.y)
		}
	}
	if got := m.Config().OutputTemplate; !strings.HasSuffix(got, "%(title)s%(title)s.%(ext)s") {
		t.Errorf("a block was not added: %q", m.Config().OutputTemplate)
	}
}

func TestEscAsksToClose(t *testing.T) {
	_, cmd := newPopup().Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("no command")
	}
	if _, ok := cmd().(CloseMsg); !ok {
		t.Errorf("esc sent %T", cmd())
	}
}

func chipAt(m Model, kind hitKind, arg int) hit {
	for _, h := range m.layout().hits {
		if h.field == 1 && h.kind == kind && h.arg == arg {
			return h
		}
	}
	return hit{x1: -1}
}

func clickHit(m Model, h hit) Model {
	return click(m, 1+h.x0, tabRows+h.y-m.scroll)
}

func templateModel() Model {
	m := newPopup().show(0)
	m.cursor = 1
	m.cfg.OutputTemplate = "%(title)s - %(artist,uploader)s.%(ext)s"
	return m
}

func TestTheTemplateIsBlocksWithTextBetweenThem(t *testing.T) {
	units := parseTemplate("%(playlist_title,playlist)s/%(playlist_index)02d - %(title)s.%(ext)s")
	var blocks, text int
	for _, u := range units {
		if u.isBlock() {
			blocks++
		} else {
			text++
		}
	}
	if blocks != 3 || text != 4 || units[0].label != "PLAYLIST NAME" || units[len(units)-1].label != "TITLE" {
		t.Errorf("blocks %d, text %d, units %+v", blocks, text, units)
	}
	if got := formatTemplate(units); got != "%(playlist_title,playlist)s/%(playlist_index)02d - %(title)s.%(ext)s" {
		t.Errorf("round trip %q", got)
	}
}

func TestABlockOfThePaletteGoesWhereTheCursorIs(t *testing.T) {
	m := templateModel()
	// A click on the text puts the cursor there: before the "-", the third unit.
	if h := chipAt(m, hitCaret, 2); h.x1 > 0 {
		m = clickHit(m, h)
	} else {
		t.Fatal("no place for the cursor")
	}
	if !m.Editing() || m.caret != 2 {
		t.Fatalf("editing %v, cursor %d", m.Editing(), m.caret)
	}
	m = clickHit(m, chipAt(m, hitBlock, 0))
	if got, want := m.Config().OutputTemplate, "%(title)s %(artist,uploader)s- %(artist,uploader)s.%(ext)s"; got != want {
		t.Errorf("template %q, want %q", got, want)
	}
}

func TestTypedTextGoesBetweenBlocksAndBackspaceTakesAWholeBlock(t *testing.T) {
	m := templateModel()
	m = clickHit(m, chipAt(m, hitCaret, len(parseTemplate(m.cfg.OutputTemplate))))
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if got := m.Config().OutputTemplate; got != "%(title)s - .%(ext)s" {
		t.Errorf("after backspace %q", got)
	}
	m, _ = m.Update(tea.KeyPressMsg{Text: "!", Code: '!'})
	if got := m.Config().OutputTemplate; got != "%(title)s - !.%(ext)s" {
		t.Errorf("after typing %q", got)
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if got := m.Config().OutputTemplate; got != "%(title)s - %(artist,uploader)s.%(ext)s" || m.Editing() {
		t.Errorf("esc left %q, editing %v", got, m.Editing())
	}
}

func TestTheCrossOnABlockRemovesItAndEscPutsTheTemplateBack(t *testing.T) {
	m := templateModel()
	m = clickHit(m, chipAt(m, hitChipRemove, 0))
	if got := m.Config().OutputTemplate; got != " - %(artist,uploader)s.%(ext)s" {
		t.Errorf("after the cross %q", got)
	}
}

func TestABlockCanBeCarriedToAnotherPlace(t *testing.T) {
	m := templateModel()
	l := m.layout()
	from := chipAt(m, hitChipHandle, 0)
	var to boundary
	for _, b := range l.bounds {
		if b.index == len(parseTemplate(m.cfg.OutputTemplate)) {
			to = b
		}
	}
	m = clickHit(m, from)
	if !m.dragging {
		t.Fatal("the block was not taken hold of")
	}
	m, _ = m.Update(tea.MouseMotionMsg{X: 1 + to.x, Y: tabRows + to.y - m.scroll, Button: tea.MouseLeft})
	m, _ = m.Update(tea.MouseReleaseMsg{X: 1 + to.x, Y: tabRows + to.y - m.scroll, Button: tea.MouseLeft})
	if got := m.Config().OutputTemplate; got != " - %(artist,uploader)s%(title)s.%(ext)s" {
		t.Errorf("after the drag %q", got)
	}
	if m.dragging {
		t.Error("still carrying")
	}
}

func TestAThemeSetsAllTheColoursAndChangingOneMakesItTheirOwn(t *testing.T) {
	m := newPopup().show(3)
	if m.Config().ThemeName != config.DefaultThemeName {
		t.Fatalf("starts with %q", m.Config().ThemeName)
	}
	// Pick Dracula: the second option after Dark+ and Monokai.
	for i := 0; m.Config().ThemeName != "Dracula" && i < 20; i++ {
		m = press(m, "right")
	}
	want, _ := config.PresetColors("Dracula")
	if m.Config().Colors != want {
		t.Errorf("colours %+v, want %+v", m.Config().Colors, want)
	}
	if look.Colours.Background != "#282a36" {
		t.Errorf("the screen shows background %q", look.Colours.Background)
	}
	// Change the accent: the theme becomes the person's own, based on Dracula.
	m = press(press(m, "down"), "down")
	m = press(m, "enter")
	m.input.SetValue("#123456")
	m = press(m, "enter")
	if m.Config().ThemeName != config.CustomTheme || m.Config().Colors.Background != "#282a36" || m.Config().Colors.Accent != "#123456" {
		t.Errorf("theme %q, colours %+v", m.Config().ThemeName, m.Config().Colors)
	}
	// The background may be emptied, for the terminal's own.
	m = press(m, "up")
	m = press(m, "enter")
	m.input.SetValue("")
	m = press(m, "enter")
	if m.Editing() || m.Config().Colors.Background != "" {
		t.Errorf("background %q, editing %v", m.Config().Colors.Background, m.Editing())
	}
	look.Apply(config.DefaultThemeColors())
}

func TestALongDescriptionWrapsInsteadOfBeingCut(t *testing.T) {
	m := New(config.Default("/app", "/music"), Host{}).SetSize(60, 40).show(2)
	l := m.layout()
	for _, line := range l.lines {
		if strings.Contains(ansi.Strip(line), "…") {
			t.Errorf("a line was cut: %q", ansi.Strip(line))
		}
	}
	all := ansi.Strip(strings.Join(l.lines, " "))
	if !strings.Contains(all, "no official video was found for.") {
		t.Errorf("the end of the description is missing")
	}
}
