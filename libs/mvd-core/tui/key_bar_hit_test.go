package tui

import (
	"testing"

	"youtube-downloader/libs/mvd-core/screens"
)

// TestKeyBarClicksLandWhereTheDesignDrawsThem keeps hintAt, which the models use to
// find the key under a click, equal to the hit test Cuppa generates for each screen.
func TestKeyBarClicksLandWhereTheDesignDrawsThem(t *testing.T) {
	hints := []keyHint{{"ctrl+s", "save"}, {"↑↓", "move"}, {"esc", "back"}, {"q", ""}, {"enter", "open"}}
	generated := map[string]func(x int) (int, bool){
		"addlinks": func(x int) (int, bool) {
			return screens.AddLinksHintsHintAt(screens.AddLinksProps{Hints: hintsOf[screens.AddLinksHintsHint](hints)}, x)
		},
		"advanced": func(x int) (int, bool) {
			return screens.AdvancedHintsHintAt(screens.AdvancedProps{Hints: hintsOf[screens.AdvancedHintsHint](hints)}, x)
		},
		"colours": func(x int) (int, bool) {
			return screens.ColoursHintsHintAt(screens.ColoursProps{Hints: hintsOf[screens.ColoursHintsHint](hints)}, x)
		},
		"download": func(x int) (int, bool) {
			return screens.DownloadHintsHintAt(screens.DownloadProps{Hints: hintsOf[screens.DownloadHintsHint](hints)}, x)
		},
		"editor": func(x int) (int, bool) {
			return screens.EditorHintsHintAt(screens.EditorProps{Hints: hintsOf[screens.EditorHintsHint](hints)}, x)
		},
		"editorplanning": func(x int) (int, bool) {
			return screens.EditorPlanningHintsHintAt(screens.EditorPlanningProps{Hints: hintsOf[screens.EditorPlanningHintsHint](hints)}, x)
		},
		"folder": func(x int) (int, bool) {
			return screens.FolderHintsHintAt(screens.FolderProps{Hints: hintsOf[screens.FolderHintsHint](hints)}, x)
		},
		"list": func(x int) (int, bool) {
			return screens.ListHintsHintAt(screens.ListProps{Hints: hintsOf[screens.ListHintsHint](hints)}, x)
		},
		"preferences": func(x int) (int, bool) {
			return screens.PreferencesHintsHintAt(screens.PreferencesProps{Hints: hintsOf[screens.PreferencesHintsHint](hints)}, x)
		},
		"songs": func(x int) (int, bool) {
			return screens.SongsHintsHintAt(screens.SongsProps{Hints: hintsOf[screens.SongsHintsHint](hints)}, x)
		},
	}
	for name, at := range generated {
		for x := 0; x < 70; x++ {
			want, wantOK := hintAt(hints, x)
			i, ok := at(x)
			if ok != wantOK || (ok && hints[i] != want) {
				t.Fatalf("%s: click at column %d: generated says (%d, %v), the app finds (%v, %v)", name, x, i, ok, want, wantOK)
			}
		}
	}
}
