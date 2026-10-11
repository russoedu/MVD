package tui

import (
	"testing"

	editorpage "youtube-downloader/libs/mvd-core/editor/editor"
	editorplanningpage "youtube-downloader/libs/mvd-core/editor/editorplanning"
	addlinkspage "youtube-downloader/libs/mvd-core/run/addlinks"
	downloadpage "youtube-downloader/libs/mvd-core/run/download"
	advancedpage "youtube-downloader/libs/mvd-core/setup/advanced"
	colourspage "youtube-downloader/libs/mvd-core/setup/colours"
	folderpage "youtube-downloader/libs/mvd-core/setup/folder"
	listpage "youtube-downloader/libs/mvd-core/setup/list"
	preferencespage "youtube-downloader/libs/mvd-core/setup/preferences"
	songspage "youtube-downloader/libs/mvd-core/setup/songs"
)

// TestKeyBarClicksLandWhereTheDesignDrawsThem keeps hintAt, which the models use to
// find the key under a click, equal to the hit test Cuppa generates for each screen.
func TestKeyBarClicksLandWhereTheDesignDrawsThem(t *testing.T) {
	hints := []keyHint{{"ctrl+s", "save"}, {"↑↓", "move"}, {"esc", "back"}, {"q", ""}, {"enter", "open"}}
	generated := map[string]func(x int) (int, bool){
		"addlinks": func(x int) (int, bool) {
			return addlinkspage.HintsHintAt(addlinkspage.Props{Hints: hintsOf[addlinkspage.HintsHint](hints)}, x)
		},
		"advanced": func(x int) (int, bool) {
			return advancedpage.HintsHintAt(advancedpage.Props{Hints: hintsOf[advancedpage.HintsHint](hints)}, x)
		},
		"colours": func(x int) (int, bool) {
			return colourspage.HintsHintAt(colourspage.Props{Hints: hintsOf[colourspage.HintsHint](hints)}, x)
		},
		"download": func(x int) (int, bool) {
			return downloadpage.HintsHintAt(downloadpage.Props{Hints: hintsOf[downloadpage.HintsHint](hints)}, x)
		},
		"editor": func(x int) (int, bool) {
			return editorpage.HintsHintAt(editorpage.Props{Hints: hintsOf[editorpage.HintsHint](hints)}, x)
		},
		"editorplanning": func(x int) (int, bool) {
			return editorplanningpage.HintsHintAt(editorplanningpage.Props{Hints: hintsOf[editorplanningpage.HintsHint](hints)}, x)
		},
		"folder": func(x int) (int, bool) {
			return folderpage.HintsHintAt(folderpage.Props{Hints: hintsOf[folderpage.HintsHint](hints)}, x)
		},
		"list": func(x int) (int, bool) {
			return listpage.HintsHintAt(listpage.Props{Hints: hintsOf[listpage.HintsHint](hints)}, x)
		},
		"preferences": func(x int) (int, bool) {
			return preferencespage.HintsHintAt(preferencespage.Props{Hints: hintsOf[preferencespage.HintsHint](hints)}, x)
		},
		"songs": func(x int) (int, bool) {
			return songspage.HintsHintAt(songspage.Props{Hints: hintsOf[songspage.HintsHint](hints)}, x)
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
