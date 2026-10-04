package uninstall

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"youtube-downloader/apps/mvd/install"
	"youtube-downloader/apps/mvd/question"
	"youtube-downloader/libs/mvd-server/api"
)

// world is a machine laid out in temporary folders: an installed app with a shortcut, an
// app-data folder, and a Downloads folder with a video in it.
type world struct {
	root      string
	installed string
	program   string
	shortcut  string
	appDir    string
	downloads string
	video     string

	answers   []question.Answer
	questions []string
	choices   [][]string
	informed  []string
	quits     int
	registry  int

	env Environment
}

func newWorld(t *testing.T, goos string, answers ...question.Answer) *world {
	t.Helper()
	root := t.TempDir()
	w := &world{
		root:      root,
		installed: filepath.Join(root, "Programs", "MVD"),
		shortcut:  filepath.Join(root, "menu", "MVD.lnk"),
		appDir:    filepath.Join(root, "roaming", "mvd"),
		downloads: filepath.Join(root, "Downloads"),
		answers:   answers,
	}
	w.program = filepath.Join(w.installed, "mvd.exe")
	w.video = filepath.Join(w.downloads, "song.mp4")
	for path, content := range map[string]string{
		w.program:                                "program",
		w.shortcut:                               "link",
		filepath.Join(w.installed, "extra.dll"):  "x",
		filepath.Join(w.appDir, "config.conf"):   "settings",
		filepath.Join(w.appDir, "list.txt"):      "urls",
		filepath.Join(w.appDir, "bin", "yt-dlp"): "tool",
		w.video:                                  "video",
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	w.env = Environment{
		GOOS: goos, Version: "1.0.0", Exe: w.program, AppDir: w.appDir,
		Keep: func() []string { return []string{w.downloads} },
		Footprint: func() install.Footprint {
			return install.Footprint{
				Places:  []install.Place{{Folder: w.installed, Program: w.program}},
				Entries: []string{w.shortcut},
			}
		},
		Ask: func(title, text string, choices []string) question.Answer {
			w.questions = append(w.questions, text)
			w.choices = append(w.choices, choices)
			if len(w.answers) == 0 {
				return question.AnswerLeave
			}
			next := w.answers[0]
			w.answers = w.answers[1:]

			return next
		},
		Inform:             func(text string) { w.informed = append(w.informed, text) },
		Exists:             func(path string) bool { _, err := os.Stat(path); return err == nil },
		ListDir:            listNames,
		RemoveAll:          os.RemoveAll,
		Remove:             os.Remove,
		RemoveProgram:      func(path string) (string, error) { return "", os.Remove(path) },
		RemoveRegistration: func() error { w.registry++; return nil },
		Quit:               func() { w.quits++ },
	}

	return w
}

func (w *world) exists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

func TestDecliningTheConfirmationRemovesNothing(t *testing.T) {
	w := newWorld(t, "windows", question.AnswerLeave)

	remove, err := Service{Env: w.env}.Confirm(nil)

	if !errors.Is(err, api.ErrUninstallDeclined) || remove != nil {
		t.Fatalf("err = %v, remove = %v", err, remove != nil)
	}
	for _, path := range []string{w.program, w.shortcut, filepath.Join(w.appDir, "config.conf"), w.video} {
		if !w.exists(path) {
			t.Errorf("%s was removed", path)
		}
	}
	if w.quits != 0 || len(w.informed) != 0 {
		t.Errorf("quits = %d, informed = %v", w.quits, w.informed)
	}
}

func TestAMachineThatCannotAskReportsThatInsteadOfRemovingAnything(t *testing.T) {
	w := newWorld(t, "linux", question.AnswerUnavailable)

	_, err := Service{Env: w.env}.Confirm(nil)

	if !errors.Is(err, api.ErrNoDialog) {
		t.Errorf("err = %v", err)
	}
	if !w.exists(w.program) {
		t.Error("the program was removed without anyone being asked")
	}
}

func TestWhenThePageChoosesThereIsOneQuestionThatListsEverythingAndNothingIsRemovedBeforeItIsAnswered(t *testing.T) {
	w := newWorld(t, "windows", question.AnswerFirst)
	deleteToo := true

	remove, err := Service{Env: w.env}.Confirm(&deleteToo)

	if err != nil {
		t.Fatal(err)
	}
	if len(w.questions) != 1 {
		t.Fatalf("questions = %d", len(w.questions))
	}
	for _, want := range []string{w.program, w.installed, w.shortcut, "Settings > Apps", w.appDir + " will be deleted", "never touched"} {
		if !strings.Contains(w.questions[0], want) {
			t.Errorf("the question does not mention %q:\n%s", want, w.questions[0])
		}
	}
	if !w.exists(w.program) || !w.exists(w.appDir) {
		t.Error("something was removed before the person had answered")
	}

	remove()

	for _, gone := range []string{w.program, w.installed, w.shortcut, w.appDir} {
		if w.exists(gone) {
			t.Errorf("%s is still there", gone)
		}
	}
	if !w.exists(w.video) || !w.exists(w.downloads) {
		t.Error("the downloads were touched")
	}
	if w.registry != 1 || w.quits != 1 || len(w.informed) != 1 || !strings.Contains(w.informed[0], "has been removed") {
		t.Errorf("registry = %d, quits = %d, informed = %v", w.registry, w.quits, w.informed)
	}
}

func TestWhenThePageKeepsThePreferencesTheyAreStillThereAfterwards(t *testing.T) {
	w := newWorld(t, "linux", question.AnswerFirst)
	keep := false

	remove, err := Service{Env: w.env}.Confirm(&keep)
	if err != nil {
		t.Fatal(err)
	}
	remove()

	for _, kept := range []string{filepath.Join(w.appDir, "config.conf"), filepath.Join(w.appDir, "list.txt"), filepath.Join(w.appDir, "bin", "yt-dlp"), w.video} {
		if !w.exists(kept) {
			t.Errorf("%s was removed", kept)
		}
	}
	if w.exists(w.program) {
		t.Error("the program is still there")
	}
	if !strings.Contains(w.informed[0], "were kept in "+w.appDir) {
		t.Errorf("informed = %s", w.informed[0])
	}
}

func TestWithoutAChoiceThePersonIsAskedAboutThePreferencesToo(t *testing.T) {
	w := newWorld(t, "linux", question.AnswerFirst, question.AnswerFirst)

	remove, err := Service{Env: w.env}.Confirm(nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(w.questions) != 2 || !strings.Contains(w.questions[1], "Also delete your preferences") || !strings.Contains(w.questions[1], "never deleted") {
		t.Fatalf("questions = %v", w.questions)
	}
	if got := strings.Join(w.choices[1], "|"); got != "Delete them too|Keep them" {
		t.Errorf("choices = %s", got)
	}
	remove()
	if w.exists(w.appDir) {
		t.Error("the preferences should have been deleted")
	}
	if !w.exists(w.video) {
		t.Error("the downloads were touched")
	}
}

func TestSayingNoToTheSecondQuestionKeepsThePreferences(t *testing.T) {
	w := newWorld(t, "linux", question.AnswerFirst, question.AnswerLeave)

	remove, err := Service{Env: w.env}.Confirm(nil)
	if err != nil {
		t.Fatal(err)
	}
	remove()

	if !w.exists(filepath.Join(w.appDir, "config.conf")) || w.exists(w.program) {
		t.Error("the program should be gone and the preferences kept")
	}
}

func TestAProblemDoesNotStopTheRestAndIsReportedWithWhatToDoAboutIt(t *testing.T) {
	w := newWorld(t, "darwin", question.AnswerFirst)
	denied := fs.ErrPermission
	w.env.RemoveAll = func(path string) error {
		if path == w.installed {
			return denied
		}

		return os.RemoveAll(path)
	}
	deleteToo := true

	remove, err := Service{Env: w.env}.Confirm(&deleteToo)
	if err != nil {
		t.Fatal(err)
	}
	remove()

	if w.exists(w.appDir) || w.exists(w.shortcut) {
		t.Error("the rest should still have been removed")
	}
	if !strings.Contains(w.informed[0], "could not be removed") || !strings.Contains(w.informed[0], "drag "+w.installed+" to the Trash") {
		t.Errorf("informed = %s", w.informed[0])
	}
	if w.quits != 1 {
		t.Errorf("quits = %d", w.quits)
	}
}

func TestADownloadsFolderInsideTheAppDataFolderSurvivesDeletingThePreferences(t *testing.T) {
	w := newWorld(t, "linux", question.AnswerFirst)
	w.downloads = filepath.Join(w.appDir, "videos")
	w.video = filepath.Join(w.downloads, "song.mp4")
	if err := os.MkdirAll(w.downloads, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(w.video, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	deleteToo := true

	remove, err := Service{Env: w.env}.Confirm(&deleteToo)
	if err != nil {
		t.Fatal(err)
	}
	remove()

	if !w.exists(w.video) {
		t.Fatal("the downloads were deleted along with the preferences")
	}
	if w.exists(filepath.Join(w.appDir, "config.conf")) {
		t.Error("the preferences should still have been deleted")
	}
	if !strings.Contains(w.informed[0], "holds your files") {
		t.Errorf("the person should be told why a folder stays: %s", w.informed[0])
	}
}

func TestADevelopersBuildKeepsItsProgram(t *testing.T) {
	w := newWorld(t, "linux", question.AnswerFirst)
	devProgram := filepath.Join(w.root, "go-build", "mvd")
	if err := os.MkdirAll(filepath.Dir(devProgram), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(devProgram, []byte("dev"), 0o755); err != nil {
		t.Fatal(err)
	}
	w.env.Version = "dev"
	w.env.Exe = devProgram
	keep := true

	remove, err := Service{Env: w.env}.Confirm(&keep)
	if err != nil {
		t.Fatal(err)
	}
	remove()

	if !w.exists(devProgram) {
		t.Error("the program of a development build was removed")
	}
	if strings.Contains(w.questions[0], devProgram) {
		t.Errorf("the question lists the development program:\n%s", w.questions[0])
	}
}

func TestAProgramLeftBehindIsReportedSoThePersonCanDeleteIt(t *testing.T) {
	w := newWorld(t, "windows", question.AnswerFirst)
	w.env.RemoveProgram = func(path string) (string, error) { return path + ".removed", nil }
	keep := false

	remove, err := Service{Env: w.env}.Confirm(&keep)
	if err != nil {
		t.Fatal(err)
	}
	remove()

	if !strings.Contains(w.informed[0], ".removed could not be deleted while MVD was running") {
		t.Errorf("informed = %s", w.informed[0])
	}
}
