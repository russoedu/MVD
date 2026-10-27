package workspace

import (
	"encoding/json"
	"strings"
	"testing"

	"youtube-downloader/libs/mvd-core/musicsearch"
	"youtube-downloader/libs/mvd-core/workspace/closequestion"
	"youtube-downloader/libs/mvd-core/workspace/windowsnapshot"
)

// intent sends what a themed window would.
func intent(m Model, in windowsnapshot.Intent) Model {
	next, _ := m.Update(WindowIntentMsg(in))

	return next.(Model)
}

// entryOf is a menu entry of the window, by its id.
func entryOf(t *testing.T, snap windowsnapshot.Snapshot, id string) windowsnapshot.MenuEntry {
	t.Helper()
	for _, menu := range snap.Menus {
		for _, entry := range menu.Entries {
			if entry.ID == id {
				return entry
			}
		}
	}
	t.Fatalf("no menu entry %q", id)

	return windowsnapshot.MenuEntry{}
}

func TestTheWindowHasTheMenusWithTheirKeysLettersAndWhatCanBeChosen(t *testing.T) {
	snap := listOf().Window()
	if len(snap.Menus) != 4 || snap.Menus[0].Label != "File" || snap.Menus[0].Mnemonic != "F" {
		t.Fatalf("menus = %+v", snap.Menus)
	}
	saveAs := entryOf(t, snap, entrySaveAs)
	if saveAs.Shortcut == nil || !saveAs.Shortcut.Ctrl || !saveAs.Shortcut.Shift || saveAs.Shortcut.Key != "S" {
		t.Errorf("Save as shortcut = %+v", saveAs.Shortcut)
	}
	if entryOf(t, snap, entryStart).Enabled {
		t.Error("Load playlists is offered for a list with no sources")
	}
	if !entryOf(t, snap, entryPreferences).SeparatorBefore || entryOf(t, snap, entryNew).SeparatorBefore {
		t.Error("the separators are not before the entries that follow them")
	}
}

func TestTheWindowsButtonsSayWhichIsTheDefaultAndWhyOneCannotBeUsed(t *testing.T) {
	snap := listOf().Window()
	if len(snap.Toolbar) != 2 {
		t.Fatalf("toolbar = %+v", snap.Toolbar)
	}
	load, add := snap.Toolbar[0], snap.Toolbar[1]
	if load.Enabled || load.DisabledReason != "Add a source first." {
		t.Errorf("Load playlists = %+v", load)
	}
	if !add.Enabled || add.Role != windowsnapshot.RoleDefault || add.ID != entryAddLink {
		t.Errorf("Add source = %+v", add)
	}
}

func TestTheWindowShowsTheSourcesOfTheList(t *testing.T) {
	snap := listOf("https://www.youtube.com/playlist?list=a", "https://www.youtube.com/playlist?list=b").Window()
	if snap.Sources.Heading != "2 sources" || len(snap.Sources.Rows) != 2 {
		t.Fatalf("sources = %+v", snap.Sources)
	}
	if snap.Activity != windowsnapshot.ActivityIdle || snap.Covered != nil {
		t.Errorf("activity %q, covered %+v", snap.Activity, snap.Covered)
	}
}

func TestAddingMusicFromAThemedWindowAddsWhatWasWritten(t *testing.T) {
	m := intent(listOf(), windowsnapshot.Intent{Kind: windowsnapshot.IntentButton, ID: entryAddLink})
	snap := m.Window()
	if snap.Adding == nil || snap.Adding.Mode != windowsnapshot.ModeURL || len(snap.Adding.Fields) != 1 {
		t.Fatalf("adding = %+v", snap.Adding)
	}
	m = intent(m, windowsnapshot.Intent{Kind: windowsnapshot.IntentAddingMode, Mode: windowsnapshot.ModeSearch})
	if fields := m.Window().Adding.Fields; len(fields) != 3 || fields[0].Field != windowsnapshot.FieldArtist {
		t.Fatalf("search fields = %+v", fields)
	}
	m = intent(m, windowsnapshot.Intent{Kind: windowsnapshot.IntentAddingText, Field: windowsnapshot.FieldArtist, Text: "Daft Punk"})
	m = intent(m, windowsnapshot.Intent{Kind: windowsnapshot.IntentAddingButton, ID: "add"})
	want := musicsearch.Query{Artist: "Daft Punk"}.Address()
	if got := addressesOf(m); m.addOpen || len(got) != 1 || got[0] != want {
		t.Errorf("window open %v, list %v, want [%s]", m.addOpen, got, want)
	}
}

func TestARefusedAdditionKeepsTheWindowOpenWithTheReason(t *testing.T) {
	m := intent(listOf(), windowsnapshot.Intent{Kind: windowsnapshot.IntentButton, ID: entryAddLink})
	m = intent(m, windowsnapshot.Intent{Kind: windowsnapshot.IntentAddingButton, ID: "add"})
	if adding := m.Window().Adding; adding == nil || adding.Problem == "" {
		t.Fatalf("adding = %+v", adding)
	}
	m = intent(m, windowsnapshot.Intent{Kind: windowsnapshot.IntentAddingButton, ID: "cancel"})
	if m.Window().Adding != nil {
		t.Error("Cancel did not close the window")
	}
}

func TestAboutAndTheShortcutsAreNotesOfTheThemeThatOKCloses(t *testing.T) {
	m := intent(listOf(), windowsnapshot.Intent{Kind: windowsnapshot.IntentMenu, ID: entryAbout})
	about := m.Window().Question
	if about == nil || about.ID != windowsnapshot.QuestionAbout || !strings.HasPrefix(about.Lines[1], "Version ") || m.Window().Covered != nil {
		t.Fatalf("about = %+v", about)
	}
	m = intent(m, windowsnapshot.Intent{Kind: windowsnapshot.IntentQuestionButton, ID: questionOK})
	if m.Window().Question != nil {
		t.Fatal("OK did not close About")
	}
	m = intent(m, windowsnapshot.Intent{Kind: windowsnapshot.IntentMenu, ID: entryShortcuts})
	keys := m.Window().Question
	if keys == nil || keys.Sections[0].Heading != "File" || keys.Sections[0].Rows[0].Does != entryNew {
		t.Fatalf("shortcuts = %+v", keys)
	}
}

func TestTheQuitQuestionIsAskedInTheThemeWithKeepingAsTheSafeAnswer(t *testing.T) {
	m := listOf()
	m.closing = closequestion.Ask()
	quit := m.Window().Question
	if quit == nil || quit.Tone != windowsnapshot.ToneCaution || quit.Buttons[1].Role != windowsnapshot.RoleDefault || quit.Buttons[1].ID != questionKeep {
		t.Fatalf("quit = %+v", quit)
	}
	m = intent(m, windowsnapshot.Intent{Kind: windowsnapshot.IntentQuestionButton, ID: questionKeep})
	if m.closing.On {
		t.Error("Keep downloading did not answer the question")
	}
}

func TestAQuestionWithALineToWriteTakesTheTextAndOK(t *testing.T) {
	m := listOf()
	m = m.ask(askAddLink, "Add a playlist or video link", "")
	ask := m.Window().Question
	if ask == nil || ask.Field == nil || ask.Title != "Add a playlist or video link" {
		t.Fatalf("ask = %+v", ask)
	}
	m = intent(m, windowsnapshot.Intent{Kind: windowsnapshot.IntentQuestionText, Text: "https://www.youtube.com/playlist?list=a"})
	if m.Window().Question.Field.Text != "https://www.youtube.com/playlist?list=a" {
		t.Fatalf("field = %+v", m.Window().Question.Field)
	}
	m = intent(m, windowsnapshot.Intent{Kind: windowsnapshot.IntentQuestionButton, ID: questionOK})
	if got := addressesOf(m); m.asking != askNone || len(got) != 1 {
		t.Errorf("asking %v, list %v", m.asking, got)
	}
}

func TestAMenuEntryThatCannotBeChosenDoesNothing(t *testing.T) {
	m := intent(listOf(), windowsnapshot.Intent{Kind: windowsnapshot.IntentMenu, ID: entryStart})
	if m.life.Starting || m.running() {
		t.Error("a list with no sources was started from a greyed entry")
	}
}

func TestTheSnapshotIsWrittenWithTheNamesTheWindowReads(t *testing.T) {
	data, err := json.Marshal(listOf("https://www.youtube.com/playlist?list=a").Window())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{`"title":`, `"menus":`, `"separatorBefore":true`, `"toolbar":`, `"sources":`, `"songs":`, `"details":`, `"status":`, `"activity":"idle"`} {
		if !strings.Contains(string(data), name) {
			t.Errorf("the snapshot has no %s:\n%s", name, data)
		}
	}
}
