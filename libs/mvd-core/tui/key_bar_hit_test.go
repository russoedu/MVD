package tui

import (
	"testing"
	"youtube-downloader/libs/mvd-core/keybar"

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
	hints := []keybar.Hint{{Key: "ctrl+s", Desc: "save"}, {Key: "↑↓", Desc: "move"}, {Key: "esc", Desc: "back"}, {Key: "q", Desc: ""}, {Key: "enter", Desc: "open"}}
	generated := map[string]func(x int) (int, bool){
		"addlinks": func(x int) (int, bool) {
			return addlinkspage.HintsHintAt(addlinkspage.Props{Hints: keybar.Of[addlinkspage.HintsHint](hints)}, x)
		},
		"advanced": func(x int) (int, bool) {
			return advancedpage.HintsHintAt(advancedpage.Props{Hints: keybar.Of[advancedpage.HintsHint](hints)}, x)
		},
		"colours": func(x int) (int, bool) {
			return colourspage.HintsHintAt(colourspage.Props{Hints: keybar.Of[colourspage.HintsHint](hints)}, x)
		},
		"download": func(x int) (int, bool) {
			return downloadpage.HintsHintAt(downloadpage.Props{Hints: keybar.Of[downloadpage.HintsHint](hints)}, x)
		},
		"editor": func(x int) (int, bool) {
			return editorpage.HintsHintAt(editorpage.Props{Hints: keybar.Of[editorpage.HintsHint](hints)}, x)
		},
		"editorplanning": func(x int) (int, bool) {
			return editorplanningpage.HintsHintAt(editorplanningpage.Props{Hints: keybar.Of[editorplanningpage.HintsHint](hints)}, x)
		},
		"folder": func(x int) (int, bool) {
			return folderpage.HintsHintAt(folderpage.Props{Hints: keybar.Of[folderpage.HintsHint](hints)}, x)
		},
		"list": func(x int) (int, bool) {
			return listpage.HintsHintAt(listpage.Props{Hints: keybar.Of[listpage.HintsHint](hints)}, x)
		},
		"preferences": func(x int) (int, bool) {
			return preferencespage.HintsHintAt(preferencespage.Props{Hints: keybar.Of[preferencespage.HintsHint](hints)}, x)
		},
		"songs": func(x int) (int, bool) {
			return songspage.HintsHintAt(songspage.Props{Hints: keybar.Of[songspage.HintsHint](hints)}, x)
		},
	}
	for name, at := range generated {
		for x := 0; x < 70; x++ {
			want, wantOK := keybar.At(hints, x)
			i, ok := at(x)
			if ok != wantOK || (ok && hints[i] != want) {
				t.Fatalf("%s: click at column %d: generated says (%d, %v), the app finds (%v, %v)", name, x, i, ok, want, wantOK)
			}
		}
	}
}
