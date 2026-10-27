package tui

import (
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

	// The playlist editor, in the states of its tests (120 by 30 cells).
	editor := func(steps func(f *editorFixture)) func(t *testing.T) string {
		return func(t *testing.T) string {
			ApplyTheme(config.DefaultThemeColors())
			f := newEditorFixture(t, samePlan())
			steps(f)
			return f.screen()
		}
	}
	cases = append(cases,
		legacyCase{"editor-planning", 120, 30, editor(func(f *editorFixture) {})},
		legacyCase{"editor-first", 120, 30, editor(func(f *editorFixture) { f.planDone() })},
		legacyCase{"editor-decided", 120, 30, editor(func(f *editorFixture) {
			f.planDone()
			f.press("c")
			f.drive(key("r"))
			f.drive(tea.PasteMsg{Content: "https://www.youtube.com/watch?v=CHOSEN00001"})
			f.drive(key("enter"))
			f.skipNote()
			f.press("up", "up")
		})},
		legacyCase{"editor-replace-prompt", 120, 30, editor(func(f *editorFixture) { f.planDone(); f.press("r") })},
		legacyCase{"editor-ask-download", 120, 30, editor(func(f *editorFixture) { f.planDone(); f.press("c", "d") })},
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
