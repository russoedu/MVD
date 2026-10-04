package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// moveProbe is a moveEnvironment whose every action is recorded, and whose answers and
// failures the test chooses.
type moveProbe struct {
	env moveEnvironment

	answers   []answer
	questions []string
	choices   [][]string
	told      []string
	installed []installTarget
	started   []startedProgram

	installErrs []error
	startErr    error
}

type startedProgram struct {
	target installTarget
	args   []string
}

func newMoveProbe(t *testing.T, goos string, answers ...answer) *moveProbe {
	t.Helper()
	p := &moveProbe{answers: answers}
	root := t.TempDir()
	p.env = moveEnvironment{
		GOOS: goos, Version: "0.0.7", Tray: true,
		Places: installPlaces{
			ProgramFiles: filepath.Join(root, "Program Files"),
			LocalAppData: filepath.Join(root, "Local"),
			Home:         filepath.Join(root, "home"),
		},
		AppDir: t.TempDir(),
		Exe:    filepath.Join(root, "Downloads", "mvd-tray (1).exe"),
		Args:   []string{"-addr", "127.0.0.1:9000"},
		Ask: func(title, question string, choices []string) answer {
			p.questions = append(p.questions, question)
			p.choices = append(p.choices, choices)
			if len(p.answers) == 0 {
				t.Fatal("asked more questions than the test expected")
			}
			next := p.answers[0]
			p.answers = p.answers[1:]

			return next
		},
		Tell: func(text string) { p.told = append(p.told, text) },
		Install: func(target installTarget, exe string) error {
			p.installed = append(p.installed, target)
			if len(p.installErrs) == 0 {
				return nil
			}
			err := p.installErrs[0]
			p.installErrs = p.installErrs[1:]

			return err
		},
		Start: func(target installTarget, args []string) error {
			p.started = append(p.started, startedProgram{target, args})

			return p.startErr
		},
	}

	return p
}

func (p *moveProbe) targets() []installTarget { return installTargets(p.env.GOOS, p.env.Places) }

func TestChoosingEveryoneInstallsInTheSystemPlaceAndStartsItWithTheSameArgumentsPlusTheOldPath(t *testing.T) {
	p := newMoveProbe(t, "windows", answerFirst)

	if !offerMove(p.env) {
		t.Fatal("expected the move to happen")
	}

	if len(p.installed) != 1 || p.installed[0] != p.targets()[0] || !p.installed[0].Everyone {
		t.Fatalf("installed = %+v", p.installed)
	}
	if len(p.started) != 1 || p.started[0].target != p.targets()[0] {
		t.Fatalf("started = %+v", p.started)
	}
	if got := strings.Join(p.started[0].args, " "); got != "-addr 127.0.0.1:9000 -moved-from="+p.env.Exe {
		t.Errorf("started with %q, want the same arguments plus the old path", got)
	}
	if len(p.told) != 0 {
		t.Errorf("told = %v", p.told)
	}
}

func TestChoosingJustForMeInstallsInThePersonsOwnPlace(t *testing.T) {
	p := newMoveProbe(t, "darwin", answerSecond)

	if !offerMove(p.env) {
		t.Fatal("expected the move to happen")
	}

	if len(p.installed) != 1 || p.installed[0] != p.targets()[1] || p.installed[0].Everyone {
		t.Errorf("installed = %+v, want the second (own) place", p.installed)
	}
}

func TestTheQuestionNamesBothPlacesWhenTheAppMayInstallForEveryone(t *testing.T) {
	p := newMoveProbe(t, "windows", answerLeave)

	offerMove(p.env)

	question := p.questions[0]
	for _, want := range []string{filepath.Dir(p.env.Exe), p.targets()[0].Folder, p.targets()[1].Folder, "only asked once"} {
		if !strings.Contains(question, want) {
			t.Errorf("the question lacks %q:\n%s", want, question)
		}
	}
	if got := strings.Join(p.choices[0], "|"); got != "For everyone|Just for me|Leave it here" {
		t.Errorf("choices = %s", got)
	}
}

func TestWithoutAdministratorRightsOnWindowsOnlyTheOwnFolderIsOfferedAndTheQuestionSaysHowToGetTheOther(t *testing.T) {
	p := newMoveProbe(t, "windows", answerLeave)
	p.env.Places.ProgramFiles = ""
	p.env.AdminHint = true

	offerMove(p.env)

	if got := strings.Join(p.choices[0], "|"); got != "Move it|Leave it here" {
		t.Errorf("choices = %s", got)
	}
	if !strings.Contains(p.questions[0], "Run as administrator") || !strings.Contains(p.questions[0], p.targets()[0].Folder) {
		t.Errorf("the question should say how to install for everyone and name the own folder:\n%s", p.questions[0])
	}
}

func TestTheHintIsLeftOutWhereItDoesNotApply(t *testing.T) {
	linux := newMoveProbe(t, "linux", answerLeave)
	offerMove(linux.env)
	if strings.Contains(linux.questions[0], "administrator") {
		t.Errorf("the Linux question mentions an administrator:\n%s", linux.questions[0])
	}

	elevated := newMoveProbe(t, "windows", answerLeave)
	offerMove(elevated.env)
	if strings.Contains(elevated.questions[0], "administrator") {
		t.Errorf("the question mentions an administrator although the app already runs as one:\n%s", elevated.questions[0])
	}
}

func TestThereIsNoAdministratorTalkOnMacAndOnlyTwoChoicesOnLinux(t *testing.T) {
	mac := newMoveProbe(t, "darwin", answerLeave)
	offerMove(mac.env)
	if strings.Contains(mac.questions[0], "administrator") {
		t.Errorf("macOS question mentions an administrator:\n%s", mac.questions[0])
	}

	linux := newMoveProbe(t, "linux", answerLeave)
	offerMove(linux.env)
	if got := strings.Join(linux.choices[0], "|"); got != "Move it|Leave it here" {
		t.Errorf("Linux choices = %s", got)
	}
	if !strings.Contains(linux.questions[0], linux.targets()[0].Folder) {
		t.Errorf("the question should name the folder:\n%s", linux.questions[0])
	}
}

func TestLinuxMovesToItsOnlyPlaceOnYes(t *testing.T) {
	p := newMoveProbe(t, "linux", answerFirst)

	if !offerMove(p.env) || len(p.installed) != 1 || p.installed[0].Kind != kindLinuxUser {
		t.Errorf("installed = %+v", p.installed)
	}
}

func TestLeavingItDoesNothingAndIsNeverAskedAgain(t *testing.T) {
	p := newMoveProbe(t, "windows", answerLeave)

	if offerMove(p.env) {
		t.Fatal("moved although the person chose to leave it")
	}
	if len(p.installed)+len(p.started) != 0 {
		t.Errorf("something was done: installed=%v started=%v", p.installed, p.started)
	}

	if offerMove(p.env) || len(p.questions) != 1 {
		t.Errorf("asked %d times in total, want exactly once", len(p.questions))
	}
}

func TestTheQuestionIsRecordedBeforeItIsPutSoACrashCannotMakeItRepeat(t *testing.T) {
	p := newMoveProbe(t, "windows")
	marker := filepath.Join(p.env.AppDir, movedMarkerName)
	p.env.Ask = func(string, string, []string) answer {
		if _, err := os.Stat(marker); err != nil {
			t.Error("the marker did not exist while the question was on screen")
		}

		return answerLeave
	}

	offerMove(p.env)
}

func TestWhenNothingCouldShowTheQuestionItIsNotCountedAsAskedAndIsTriedAgainNextTime(t *testing.T) {
	p := newMoveProbe(t, "linux", answerUnavailable, answerLeave)

	if offerMove(p.env) {
		t.Fatal("moved with no answer")
	}
	if _, err := os.Stat(filepath.Join(p.env.AppDir, movedMarkerName)); err == nil {
		t.Error("the marker was kept although nothing was asked")
	}

	offerMove(p.env)

	if len(p.questions) != 2 {
		t.Errorf("asked %d times, want 2 (the second start tries again)", len(p.questions))
	}
}

func TestNothingHappensWhereTheOfferDoesNotApply(t *testing.T) {
	cases := map[string]func(*moveProbe){
		"a dev build":     func(p *moveProbe) { p.env.Version = "dev" },
		"no tray":         func(p *moveProbe) { p.env.Tray = false },
		"a freshly moved": func(p *moveProbe) { p.env.MovedFrom = "old.exe" },
		"nowhere to go":   func(p *moveProbe) { p.env.Places = installPlaces{} },
		"already in the system place": func(p *moveProbe) {
			p.env.Exe = p.targets()[0].Program
		},
		"already in the own place": func(p *moveProbe) {
			p.env.Exe = p.targets()[1].Program
		},
	}
	for name, change := range cases {
		p := newMoveProbe(t, "windows", answerFirst)
		change(p)

		if offerMove(p.env) || len(p.questions) != 0 {
			t.Errorf("%s: moved or asked (%v)", name, p.questions)
		}
	}
}

func TestWhenInstallingForEveryoneFailsThePersonIsOfferedTheirOwnPlaceInstead(t *testing.T) {
	p := newMoveProbe(t, "windows", answerFirst, answerFirst)
	p.installErrs = []error{errors.New("administrator approval was declined")}

	if !offerMove(p.env) {
		t.Fatal("expected the fallback to succeed")
	}

	if len(p.installed) != 2 || !p.installed[0].Everyone || p.installed[1] != p.targets()[1] {
		t.Fatalf("installed = %+v, want the system place then the own place", p.installed)
	}
	if len(p.questions) != 2 || !strings.Contains(p.questions[1], "administrator approval was declined") || !strings.Contains(p.questions[1], p.targets()[1].Folder) {
		t.Errorf("the second question should give the reason and the place: %v", p.questions)
	}
	if p.started[0].target != p.targets()[1] {
		t.Errorf("started %+v, want the own place", p.started)
	}
}

func TestDecliningTheFallbackLeavesTheAppWhereItIsWithoutNaggingWithAnError(t *testing.T) {
	p := newMoveProbe(t, "windows", answerFirst, answerLeave)
	p.installErrs = []error{errors.New("administrator approval was declined")}

	if offerMove(p.env) {
		t.Fatal("moved although the person declined both")
	}
	if len(p.installed) != 1 || len(p.started) != 0 || len(p.told) != 0 {
		t.Errorf("installed=%v started=%v told=%v", p.installed, p.started, p.told)
	}
}

func TestAFailureOfThePersonsOwnPlaceIsReportedAndTheAppKeepsRunningWhereItIs(t *testing.T) {
	p := newMoveProbe(t, "windows", answerSecond)
	p.installErrs = []error{errors.New("access denied")}

	if offerMove(p.env) {
		t.Fatal("reported a move that did not happen")
	}
	if len(p.told) != 1 || !strings.Contains(p.told[0], "access denied") || !strings.Contains(p.told[0], "keep running") {
		t.Errorf("told = %v", p.told)
	}
	if len(p.started) != 0 || len(p.questions) != 1 {
		t.Errorf("went on after a failure: started=%v questions=%v", p.started, p.questions)
	}
}

func TestFailingForEveryoneWithNoPlaceOfTheirOwnIsReported(t *testing.T) {
	p := newMoveProbe(t, "windows", answerFirst)
	p.env.Places.LocalAppData = ""
	p.installErrs = []error{errors.New("administrator approval was declined")}

	if offerMove(p.env) {
		t.Fatal("reported a move that did not happen")
	}
	if len(p.told) != 1 || !strings.Contains(p.told[0], "administrator approval was declined") {
		t.Errorf("told = %v", p.told)
	}
}

func TestAProgramThatCannotBeStartedFromItsNewPlaceIsReportedAndTheAppKeepsRunningWhereItIs(t *testing.T) {
	p := newMoveProbe(t, "windows", answerSecond)
	p.startErr = errors.New("blocked by antivirus")

	if offerMove(p.env) {
		t.Fatal("reported a move although the copy did not start")
	}
	if len(p.told) != 1 || !strings.Contains(p.told[0], "blocked by antivirus") {
		t.Errorf("told = %v", p.told)
	}
}

func TestAnUnwritableMarkerMeansNoQuestionRatherThanAQuestionAtEveryStart(t *testing.T) {
	p := newMoveProbe(t, "windows", answerFirst)
	p.env.AppDir = filepath.Join(p.env.AppDir, "does", "not", "exist")

	if offerMove(p.env) || len(p.questions) != 0 {
		t.Errorf("asked although it could not remember having asked: %v", p.questions)
	}
}

// --- the installers, on real files ---------------------------------------------------

func writeProgram(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "downloaded", "mvd-tray (1).exe")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestTheMacBundleHoldsTheProgramAndAnInfoPlistWithTheVersion(t *testing.T) {
	exe := writeProgram(t, "program")
	target := installTargets("darwin", installPlaces{Home: t.TempDir()})[1]

	if err := installMacBundle(target, exe, "0.0.7"); err != nil {
		t.Fatal(err)
	}

	if data, _ := os.ReadFile(target.Program); string(data) != "program" {
		t.Errorf("program = %q", data)
	}
	if plist, _ := os.ReadFile(filepath.Join(target.Folder, "Contents", "Info.plist")); !strings.Contains(string(plist), "<string>0.0.7</string>") {
		t.Errorf("Info.plist = %s", plist)
	}
	if info, err := os.Stat(target.Program); err != nil || info.Mode()&0o100 == 0 && os.PathSeparator == '/' {
		t.Errorf("the program in the bundle must be executable: %v %v", info, err)
	}
}

func TestAnOlderBundleIsReplaced(t *testing.T) {
	target := installTargets("darwin", installPlaces{Home: t.TempDir()})[1]
	if err := installMacBundle(target, writeProgram(t, "old"), "0.0.1"); err != nil {
		t.Fatal(err)
	}

	if err := installMacBundle(target, writeProgram(t, "new"), "0.0.2"); err != nil {
		t.Fatal(err)
	}

	if data, _ := os.ReadFile(target.Program); string(data) != "new" {
		t.Errorf("program = %q", data)
	}
}

func TestOnLinuxTheProgramGoesToLocalBinWithAMenuEntryPointingAtIt(t *testing.T) {
	exe := writeProgram(t, "program")
	target := installTargets("linux", installPlaces{Home: t.TempDir()})[0]
	menu := filepath.Join(t.TempDir(), "share", "applications")

	if err := installLinuxUser(target, exe, menu); err != nil {
		t.Fatal(err)
	}

	if data, _ := os.ReadFile(target.Program); string(data) != "program" {
		t.Errorf("program = %q", data)
	}
	entry, err := os.ReadFile(filepath.Join(menu, "mvd.desktop"))
	if err != nil || string(entry) != desktopEntry(target.Program) {
		t.Errorf("the entry is not the one for %s: %q (%v)", target.Program, entry, err)
	}
}

func TestInstallingForEveryoneOnWindowsIsAPlainCopyPlusAShortcutForEveryAccount(t *testing.T) {
	exe := writeProgram(t, "program")
	target := installTargets("windows", installPlaces{ProgramFiles: t.TempDir()})[0]
	link := filepath.Join(t.TempDir(), "ProgramData", "Start Menu", "MVD.lnk")
	var links [][2]string

	err := installWindowsSystem(target, exe, link, func(l, to string) error {
		links = append(links, [2]string{l, to})

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(target.Program); string(data) != "program" {
		t.Errorf("program = %q", data)
	}
	if len(links) != 1 || links[0] != [2]string{link, target.Program} {
		t.Errorf("shortcut = %v", links)
	}
}

func TestInstallingForEveryoneFailsCleanlyWhenTheCopyFails(t *testing.T) {
	target := installTargets("windows", installPlaces{ProgramFiles: t.TempDir()})[0]

	err := installWindowsSystem(target, filepath.Join(t.TempDir(), "missing.exe"), "x.lnk", func(string, string) error { return nil })

	if err == nil {
		t.Error("expected an error")
	}
}

func TestOnWindowsTheOwnProgramsFolderGetsTheProgramAndAShortcutButAMissingShortcutIsForgiven(t *testing.T) {
	exe := writeProgram(t, "program")
	target := installTargets("windows", installPlaces{LocalAppData: t.TempDir()})[0]
	startMenu := filepath.Join(t.TempDir(), "Start Menu")
	var links [][2]string

	err := installWindowsUser(target, exe, startMenu, func(link, to string) error {
		links = append(links, [2]string{link, to})

		return errors.New("no PowerShell")
	})

	if err != nil {
		t.Fatalf("a missing shortcut must not fail the install: %v", err)
	}
	if len(links) != 1 || links[0] != [2]string{filepath.Join(startMenu, "MVD.lnk"), target.Program} {
		t.Errorf("shortcut = %v", links)
	}
	if data, _ := os.ReadFile(target.Program); string(data) != "program" {
		t.Errorf("program = %q", data)
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
