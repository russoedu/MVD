package prefspopup

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
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
			m.scroll = 0
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
	if !strings.HasSuffix(m.Config().OutputTemplate, "%(title)s") {
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

func TestABlockGoesWhereThePointerPutTheCursorInTheTemplate(t *testing.T) {
	m := newPopup().show(0)
	m.cursor = 1
	m.cfg.OutputTemplate = "%(title)s.%(ext)s"
	var value, block hit
	for _, h := range m.layout().hits {
		if h.field == 1 && h.kind == hitValue {
			value = h
		}
		if h.field == 1 && h.kind == hitBlock && h.arg == 0 {
			block = h
		}
	}
	// A click on the pattern starts typing in it, the cursor where the pointer was (after
	// the 9 characters of "%(title)s"; the page is in a box, and a prompt is before the text).
	m = click(m, 1+gutterWidth+promptWidth+9, tabRows+value.y-m.scroll)
	if !m.Editing() || m.input.Position() != 9 {
		t.Fatalf("editing %v, cursor %d", m.Editing(), m.input.Position())
	}
	m = click(m, 1+block.x0+1, tabRows+block.y-m.scroll)
	if got, want := m.Config().OutputTemplate, "%(title)s%(artist,uploader)s.%(ext)s"; got != want {
		t.Errorf("template %q, want %q", got, want)
	}
	if !m.Editing() {
		t.Error("typing ended with the block")
	}
	m = press(m, "esc")
	if got := m.Config().OutputTemplate; got != "%(title)s.%(ext)s" {
		t.Errorf("esc left %q", got)
	}
}
