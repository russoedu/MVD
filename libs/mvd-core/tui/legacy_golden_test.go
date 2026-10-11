package tui

import (
	"charm.land/bubbles/v2/textarea"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"youtube-downloader/libs/mvd-core/testing/drawing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
)

// legacyCase is one screen in one state, drawn by the hand written code that the Cuppa
// screens replace. The drawing is kept in testdata/legacy as the reference: a screen
// made from a design must draw the same cells.
type legacyCase struct {
	name string
	w, h int
	draw func(t *testing.T) string
}

func legacyCases() []legacyCase {
	cases := []legacyCase{}

	// The download screen, in the states the screens test walks through.
	download := func(setup func(m *model)) func(t *testing.T) string {
		return func(t *testing.T) string {
			ApplyTheme(config.DefaultThemeColors())
			m := sampleModel(t)
			m.width, m.height = 120, 36
			setup(&m)
			return m.View().Content
		}
	}
	cases = append(cases,
		legacyCase{"download-entries", 120, 36, download(func(m *model) { m.focus = paneEntries; m.selEntry = 4 })},
		legacyCase{"download-lists", 120, 36, download(func(m *model) { m.focus = paneLists })},
		legacyCase{"download-emoji-playlist", 120, 36, download(func(m *model) { m.focus = paneLists; m.selPlaylist = 1 })},
		legacyCase{"download-failed-playlist", 120, 36, download(func(m *model) { m.focus = paneLists; m.selPlaylist = 2 })},
		legacyCase{"download-full-log", 120, 36, download(func(m *model) { m.fullLog = true; m.focus = paneOutput; m.outScroll = 5 })},
		legacyCase{"download-filter-failed", 120, 36, download(func(m *model) { m.filterFailed = true; m.focus = paneEntries })},
		legacyCase{"download-confirm-quit", 120, 36, download(func(m *model) { m.confirmQuit = true })},
		legacyCase{"download-help", 120, 36, download(func(m *model) { m.showHelp = true })},
		legacyCase{"download-add-links", 120, 36, download(func(m *model) {
			ta := textarea.New()
			ta.Placeholder = "Paste or type playlist / video URLs, one per line"
			ta.ShowLineNumbers = false
			ta.SetWidth(m.addBoxWidth())
			ta.SetHeight(8)
			ta.SetValue("https://youtube.com/playlist?list=A\nhttps://youtube.com/watch?v=B")
			ta.Focus()
			m.adding, m.links = true, ta
		})},
		legacyCase{"download-add-links-narrow", 70, 20, download(func(m *model) {
			m.width, m.height = 70, 20
			ta := textarea.New()
			ta.Placeholder = "Paste or type playlist / video URLs, one per line"
			ta.ShowLineNumbers = false
			ta.SetWidth(m.addBoxWidth())
			ta.SetHeight(8)
			ta.Focus()
			m.adding, m.links = true, ta
		})},
		legacyCase{"download-help-narrow", 70, 20, download(func(m *model) { m.width, m.height = 70, 20; m.showHelp = true })},
		legacyCase{"download-narrow", 80, 24, download(func(m *model) { m.width, m.height = 80, 24; m.focus = paneEntries; m.selEntry = 2 })},
	)

	// The playlist editor, in the states of its tests (120 by 30 cells).
	editor := func(steps func(f *editorFixture)) func(t *testing.T) string {
		return func(t *testing.T) string {
			ApplyTheme(config.DefaultThemeColors())
			f := newEditorFixture(t, []string{"https://a"}, samePlan())
			steps(f)
			return f.screen()
		}
	}
	cases = append(cases,
		legacyCase{"editor-planning", 120, 30, editor(func(f *editorFixture) { f.press("ctrl+e") })},
		legacyCase{"editor-first", 120, 30, editor(func(f *editorFixture) { f.press("ctrl+e"); f.planDone() })},
		legacyCase{"editor-decided", 120, 30, editor(func(f *editorFixture) {
			f.press("ctrl+e")
			f.planDone()
			f.press("c")
			f.m = drive(f.t, f.m, key("r"))
			f.m = drive(f.t, f.m, tea.PasteMsg{Content: "https://www.youtube.com/watch?v=CHOSEN00001"})
			f.m = drive(f.t, f.m, key("enter"))
			f.skipNote()
			f.press("up", "up")
		})},
		legacyCase{"editor-replace-prompt", 120, 30, editor(func(f *editorFixture) { f.press("ctrl+e"); f.planDone(); f.press("r") })},
		legacyCase{"editor-ask-download", 120, 30, editor(func(f *editorFixture) { f.press("ctrl+e"); f.planDone(); f.press("c", "d") })},
	)
	return cases
}

// knownGaps are screens that do not yet draw exactly what they drew, each waiting for
// Cuppa: name -> what is missing. They are reported, not failed; a gap that is fixed
// must be removed from here, or the test says so.
var knownGaps = map[string]string{}

// TestLegacyScreensKeepTheirLook compares what the screens draw with the reference
// drawings in testdata/legacy, cell by cell (character, colours, bold and so on).
// UPDATE_LEGACY=name,name rewrites those references from the current code: only for a
// screen that was never migrated, or the reference is lost (git has the old ones).
func TestLegacyScreensKeepTheirLook(t *testing.T) {
	dir := filepath.Join("testdata", "legacy")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, c := range legacyCases() {
		names = append(names, c.name)
		path := filepath.Join(dir, c.name+".ansi")
		got := c.draw(t)
		if wanted := os.Getenv("UPDATE_LEGACY"); wanted != "" && strings.Contains(","+wanted+",", ","+c.name+",") {
			if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("%s: no reference drawing (%v); run with UPDATE_LEGACY=1", c.name, err)
			continue
		}
		diffs := drawing.SameCells(string(want), got, c.w, 8)
		if reason, known := knownGaps[c.name]; known {
			if len(diffs) == 0 {
				t.Errorf("%s now matches its reference: remove it from knownGaps", c.name)
			} else {
				t.Logf("%s: known gap: %s", c.name, reason)
			}
			continue
		}
		if len(diffs) > 0 {
			t.Errorf("%s differs from its reference:\n  %s", c.name, strings.Join(diffs, "\n  "))
		}
	}
	sort.Strings(names)
}
