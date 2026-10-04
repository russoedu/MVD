package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// moveProbe is a moveEnvironment whose every action is recorded.
type moveProbe struct {
	env moveEnvironment

	asked    []string
	told     []string
	copied   [][2]string
	links    [][2]string
	started  []startedProgram
	answer   bool
	copyErr  error
	linkErr  error
	startErr error
}

type startedProgram struct {
	path string
	args []string
}

func newMoveProbe(t *testing.T) *moveProbe {
	t.Helper()
	p := &moveProbe{answer: true}
	local := t.TempDir()
	p.env = moveEnvironment{
		GOOS: "windows", Version: "0.0.7", Tray: true,
		LocalAppData: local, StartMenu: filepath.Join(t.TempDir(), "Start Menu"), AppDir: t.TempDir(),
		Exe:  filepath.Join(t.TempDir(), "Downloads", "mvd-tray (1).exe"),
		Args: []string{"-addr", "127.0.0.1:9000"},
		Ask: func(title, text string) bool {
			p.asked = append(p.asked, text)

			return p.answer
		},
		Tell:     func(text string) { p.told = append(p.told, text) },
		Copy:     func(src, dst string) error { p.copied = append(p.copied, [2]string{src, dst}); return p.copyErr },
		Shortcut: func(link, target string) error { p.links = append(p.links, [2]string{link, target}); return p.linkErr },
		Start: func(path string, args []string) error {
			p.started = append(p.started, startedProgram{path, args})

			return p.startErr
		},
	}

	return p
}

func (p *moveProbe) target() string { return installFolder("windows", p.env.LocalAppData) }

func TestAgreeingMovesTheProgramAddsAShortcutAndStartsTheCopyWithTheOldPathAndTheSameArguments(t *testing.T) {
	p := newMoveProbe(t)

	moved := offerMove(p.env)

	if !moved {
		t.Fatal("expected the move to happen")
	}
	destination := filepath.Join(p.target(), "mvd-tray.exe")
	if len(p.copied) != 1 || p.copied[0] != [2]string{p.env.Exe, destination} {
		t.Errorf("copied = %v, want the program to land as %s whatever it was called", p.copied, destination)
	}
	if len(p.links) != 1 || p.links[0][0] != filepath.Join(p.env.StartMenu, "MVD.lnk") || p.links[0][1] != destination {
		t.Errorf("shortcut = %v", p.links)
	}
	if len(p.started) != 1 || p.started[0].path != destination {
		t.Fatalf("started = %+v", p.started)
	}
	if got := strings.Join(p.started[0].args, " "); got != "-addr 127.0.0.1:9000 -moved-from="+p.env.Exe {
		t.Errorf("the copy was started with %q, want the same arguments plus the old path", got)
	}
	if len(p.told) != 0 {
		t.Errorf("told = %v", p.told)
	}
}

func TestTheQuestionNamesWhereTheAppIsAndWhereItWouldGo(t *testing.T) {
	p := newMoveProbe(t)

	offerMove(p.env)

	if len(p.asked) != 1 || !strings.Contains(p.asked[0], filepath.Dir(p.env.Exe)) || !strings.Contains(p.asked[0], p.target()) {
		t.Errorf("question = %v", p.asked)
	}
	if !strings.Contains(p.asked[0], "not be asked again") {
		t.Error("the person should be told that No is final")
	}
}

func TestDecliningLeavesEverythingAloneAndIsNeverAskedAgain(t *testing.T) {
	p := newMoveProbe(t)
	p.answer = false

	if offerMove(p.env) {
		t.Fatal("moved although the answer was no")
	}
	if len(p.copied)+len(p.links)+len(p.started) != 0 {
		t.Errorf("something was done: copied=%v links=%v started=%v", p.copied, p.links, p.started)
	}

	if offerMove(p.env) || len(p.asked) != 1 {
		t.Errorf("asked %d times in total, want exactly once", len(p.asked))
	}
}

func TestTheQuestionIsRecordedBeforeItIsPutSoACrashCannotMakeItRepeat(t *testing.T) {
	p := newMoveProbe(t)
	marker := filepath.Join(p.env.AppDir, movedMarkerName)
	p.env.Ask = func(string, string) bool {
		if _, err := os.Stat(marker); err != nil {
			t.Error("the marker did not exist while the question was on screen")
		}

		return false
	}

	offerMove(p.env)
}

func TestNothingHappensWhereTheOfferDoesNotApply(t *testing.T) {
	cases := map[string]func(*moveEnvironment){
		"macOS":           func(e *moveEnvironment) { e.GOOS = "darwin" },
		"a dev build":     func(e *moveEnvironment) { e.Version = "dev" },
		"no tray":         func(e *moveEnvironment) { e.Tray = false },
		"a freshly moved": func(e *moveEnvironment) { e.MovedFrom = "old.exe" },
		"already installed": func(e *moveEnvironment) {
			e.Exe = filepath.Join(installFolder("windows", e.LocalAppData), "mvd-tray.exe")
		},
	}
	for name, change := range cases {
		p := newMoveProbe(t)
		change(&p.env)

		if offerMove(p.env) || len(p.asked) != 0 {
			t.Errorf("%s: moved or asked (%v)", name, p.asked)
		}
	}
}

func TestAFailedCopyIsReportedAndTheAppKeepsRunningWhereItIs(t *testing.T) {
	p := newMoveProbe(t)
	p.copyErr = errors.New("access denied")

	if offerMove(p.env) {
		t.Fatal("reported a move that did not happen")
	}
	if len(p.told) != 1 || !strings.Contains(p.told[0], "access denied") || !strings.Contains(p.told[0], "keep running") {
		t.Errorf("told = %v", p.told)
	}
	if len(p.started) != 0 || len(p.links) != 0 {
		t.Errorf("went on after a failed copy: started=%v links=%v", p.started, p.links)
	}
}

func TestAMissingShortcutDoesNotUndoTheMove(t *testing.T) {
	p := newMoveProbe(t)
	p.linkErr = errors.New("no PowerShell")

	if !offerMove(p.env) || len(p.started) != 1 {
		t.Errorf("the move should still go ahead: started=%v", p.started)
	}
	if len(p.told) != 0 {
		t.Errorf("a missing shortcut should not alarm the person: %v", p.told)
	}
}

func TestACopyThatCannotBeStartedIsReportedAndTheAppKeepsRunningWhereItIs(t *testing.T) {
	p := newMoveProbe(t)
	p.startErr = errors.New("blocked by antivirus")

	if offerMove(p.env) {
		t.Fatal("reported a move although the copy did not start")
	}
	if len(p.told) != 1 || !strings.Contains(p.told[0], "blocked by antivirus") {
		t.Errorf("told = %v", p.told)
	}
}

func TestAnUnwritableMarkerMeansNoQuestionRatherThanAQuestionAtEveryStart(t *testing.T) {
	p := newMoveProbe(t)
	p.env.AppDir = filepath.Join(p.env.AppDir, "does", "not", "exist")

	if offerMove(p.env) || len(p.asked) != 0 {
		t.Errorf("asked although it could not remember having asked: %v", p.asked)
	}
}

// --- copying -----------------------------------------------------------

func TestCopyExecutableCopiesTheBytesIntoAFolderItCreatesAndLeavesNothingElse(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.exe")
	if err := os.WriteFile(src, []byte("program"), 0o755); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "new", "folder", "mvd-tray.exe")

	if err := copyExecutable(src, dst); err != nil {
		t.Fatal(err)
	}

	if data, _ := os.ReadFile(dst); string(data) != "program" {
		t.Errorf("content = %q", data)
	}
	entries, _ := os.ReadDir(filepath.Dir(dst))
	if len(entries) != 1 {
		t.Errorf("the folder holds %d items, want only the program", len(entries))
	}
}

func TestCopyExecutableReplacesAnOlderCopy(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.exe")
	dst := filepath.Join(dir, "mvd-tray.exe")
	if err := os.WriteFile(src, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := copyExecutable(src, dst); err != nil {
		t.Fatal(err)
	}

	if data, _ := os.ReadFile(dst); string(data) != "new" {
		t.Errorf("content = %q", data)
	}
}

func TestCopyExecutableFailsCleanlyWhenThereIsNothingToCopy(t *testing.T) {
	dir := t.TempDir()

	if err := copyExecutable(filepath.Join(dir, "missing.exe"), filepath.Join(dir, "out", "mvd-tray.exe")); err == nil {
		t.Fatal("expected an error")
	}
	if _, err := os.Stat(filepath.Join(dir, "out", "mvd-tray.exe.part")); err == nil {
		t.Error("a partial file was left behind")
	}
}

// --- removing the old copy ---------------------------------------------------

func TestTheOldCopyIsRemovedOnceItStopsBeingInUse(t *testing.T) {
	attempts := 0
	var slept []time.Duration

	done := removeAfterExit("old.exe", 10, 200*time.Millisecond,
		func(string) error {
			attempts++
			if attempts < 4 {
				return errors.New("in use")
			}

			return nil
		},
		func(d time.Duration) { slept = append(slept, d) })

	if !done || attempts != 4 || len(slept) != 3 || slept[0] != 200*time.Millisecond {
		t.Errorf("done=%v attempts=%d slept=%v", done, attempts, slept)
	}
}

func TestAnOldCopyThatIsAlreadyGoneCountsAsRemoved(t *testing.T) {
	done := removeAfterExit("old.exe", 5, time.Millisecond,
		func(path string) error { return &os.PathError{Op: "remove", Path: path, Err: os.ErrNotExist} },
		func(time.Duration) {})

	if !done {
		t.Error("a program that is not there should count as removed")
	}
}

func TestAnOldCopyThatCannotBeRemovedIsGivenUpOnAfterTheAttempts(t *testing.T) {
	attempts := 0

	done := removeAfterExit("old.exe", 5, time.Millisecond,
		func(string) error { attempts++; return errors.New("in use") },
		func(time.Duration) {})

	if done || attempts != 5 {
		t.Errorf("done=%v attempts=%d, want false after 5", done, attempts)
	}
}
