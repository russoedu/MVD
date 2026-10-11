package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

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

func keysOf(keys ...string) []tea.Msg {
	var out []tea.Msg
	for _, k := range keys {
		if k == "backspace" {
			out = append(out, tea.KeyPressMsg{Code: tea.KeyBackspace})
			continue
		}
		out = append(out, key(k))
	}
	return out
}

func legacyCases() []legacyCase {
	const w, h = 100, 30
	cfg := func() config.Config {
		dir := "C:\\Music"
		c := config.Default(dir, "C:\\Music\\logs")
		return c
	}

	colours := func(keys ...string) func(t *testing.T) string {
		return func(t *testing.T) string {
			ApplyTheme(config.DefaultThemeColors())
			m := newColoursModel(cfg()).setSize(w, h)
			for _, msg := range keysOf(keys...) {
				m, _, _, _ = m.update(msg)
			}
			return m.view(w, h)
		}
	}
	advanced := func(extra []string, keys ...string) func(t *testing.T) string {
		return func(t *testing.T) string {
			ApplyTheme(config.DefaultThemeColors())
			c := cfg()
			c.ExtraArgs = extra
			m := newAdvancedModel(c).setSize(w, h)
			for _, msg := range keysOf(keys...) {
				m, _, _, _ = m.update(msg)
			}
			return m.view(w, h)
		}
	}
	list := func(urls []string, keys ...string) func(t *testing.T) string {
		return func(t *testing.T) string {
			ApplyTheme(config.DefaultThemeColors())
			m := newListModel(urls).setSize(w, h)
			for _, msg := range keysOf(keys...) {
				m, _, _ = m.update(msg)
			}
			return m.view(w, h)
		}
	}
	songs := func(text string) func(t *testing.T) string {
		return func(t *testing.T) string {
			ApplyTheme(config.DefaultThemeColors())
			m := newSongsModel("C:\\Music\\songs").setSize(w, h)
			m.ta.SetValue(text)
			return m.view(w, h)
		}
	}
	folder := func(creating bool, err string) func(t *testing.T) string {
		return func(t *testing.T) string {
			ApplyTheme(config.DefaultThemeColors())
			m := folderModel{dir: "C:\\Users\\me\\Music", entries: []string{"90s", "Dance", "Rock", "Soundtracks"}, cursor: 1, err: err}
			m.input = newFolderModel(".").input
			m.creating = creating
			return m.setSize(w, h).view(w, h)
		}
	}
	prefs := func(keys ...string) func(t *testing.T) string {
		return func(t *testing.T) string {
			ApplyTheme(config.DefaultThemeColors())
			m := newConfigModel(cfg()).setSize(w, h)
			for _, msg := range keysOf(keys...) {
				m, _, _, _ = m.update(msg)
			}
			return m.view(w, h)
		}
	}

	cases := []legacyCase{
		{"colours-first", w, h, colours()},
		{"colours-fourth", w, h, colours("down", "down", "down")},
		{"colours-editing", w, h, colours("down", "enter")},
		{"colours-error", w, h, colours("enter", "backspace", "z", "enter")},
		{"advanced-first", w, h, advanced(nil)},
		{"advanced-args", w, h, advanced([]string{"--proxy", "http://x"}, "down")},
		{"advanced-editing", w, h, advanced(nil, "enter")},
		{"list-empty", w, h, list(nil)},
		{"list-three", w, h, list([]string{"https://youtube.com/playlist?list=A", "https://youtube.com/playlist?list=B", "C:\\songs.txt"})},
		{"list-confirm-quit", w, h, list([]string{"https://youtube.com/playlist?list=A"}, "esc")},
		{"songs-empty", w, h, songs("")},
		{"songs-two", w, h, songs("Daft Punk - One More Time\nbad line\nAlexia - Uh La La La")},
		{"folder-browse", w, h, folder(false, "")},
		{"folder-creating", w, h, folder(true, "")},
		{"folder-error", w, h, folder(false, "open C:\\Users\\me\\Music: access denied")},
		{"config-first", w, h, prefs()},
		{"config-template", w, h, prefs("down", "down", "down", "down")},
		{"config-editing", w, h, prefs("down", "enter")},
	}

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
var knownGaps = map[string]string{
	"colours-editing": "a cell's own text cannot ask for the terminal's default colour (GAP default-fg)",
	"colours-error":   "a cell's own text cannot ask for the terminal's default colour (GAP default-fg)",
}

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
		diffs := sameCells(string(want), got, c.w, 8)
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
