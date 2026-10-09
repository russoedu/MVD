package install

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"youtube-downloader/apps/mvd/question"
)

// moveProbe is a moveEnvironment whose every action is recorded, and whose answers and
// failures the test chooses.
type moveProbe struct {
	env moveEnvironment

	answers   []question.Answer
	questions []string
	choices   [][]string
	told      []string
	installed []installTarget
	started   []startedProgram

	installErrs []error
	startErr    error
}

// answerAskAgain is the second choice of the question after "Not now".
const answerAskAgain = question.AnswerLeave

type startedProgram struct {
	target installTarget
	args   []string
}

func newMoveProbe(t *testing.T, goos string, answers ...question.Answer) *moveProbe {
	t.Helper()
	p := &moveProbe{answers: answers}
	root := t.TempDir()
	p.env = moveEnvironment{
		GOOS: goos, Version: "0.0.7", Window: true,
		Places: installPlaces{
			ProgramFiles: filepath.Join(root, "Program Files"),
			LocalAppData: filepath.Join(root, "Local"),
			Home:         filepath.Join(root, "home"),
		},
		AppDir: t.TempDir(),
		Exe:    filepath.Join(root, "Downloads", "mvd (1).exe"),
		Args:   []string{"-addr", "127.0.0.1:9000"},
		Ask: func(title, question string, choices []string) question.Answer {
			p.questions = append(p.questions, question)
			p.choices = append(p.choices, choices)
			// Unless a test says otherwise, "Not now" is followed by "ask me again".
			if len(p.answers) == 0 && len(choices) == 2 && choices[0] == "Never ask again" {
				return answerAskAgain
			}
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
	p := newMoveProbe(t, "windows", question.AnswerFirst)

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
	p := newMoveProbe(t, "darwin", question.AnswerSecond)

	if !offerMove(p.env) {
		t.Fatal("expected the move to happen")
	}

	if len(p.installed) != 1 || p.installed[0] != p.targets()[1] || p.installed[0].Everyone {
		t.Errorf("installed = %+v, want the second (own) place", p.installed)
	}
}

func TestTheQuestionNamesBothPlacesWhenTheAppMayInstallForEveryone(t *testing.T) {
	p := newMoveProbe(t, "windows", question.AnswerLeave)

	offerMove(p.env)

	question := p.questions[0]
	for _, want := range []string{filepath.Dir(p.env.Exe), p.targets()[0].Folder, p.targets()[1].Folder, "later, from the preferences"} {
		if !strings.Contains(question, want) {
			t.Errorf("the question lacks %q:\n%s", want, question)
		}
	}
	if got := strings.Join(p.choices[0], "|"); got != "For everyone|Just for me|Not now" {
		t.Errorf("choices = %s", got)
	}
}

func TestWithoutAdministratorRightsOnWindowsOnlyTheOwnFolderIsOfferedAndTheQuestionSaysHowToGetTheOther(t *testing.T) {
	p := newMoveProbe(t, "windows", question.AnswerLeave)
	p.env.Places.ProgramFiles = ""
	p.env.AdminHint = true

	offerMove(p.env)

	if got := strings.Join(p.choices[0], "|"); got != "Move it|Not now" {
		t.Errorf("choices = %s", got)
	}
	if !strings.Contains(p.questions[0], "Run as administrator") || !strings.Contains(p.questions[0], p.targets()[0].Folder) {
		t.Errorf("the question should say how to install for everyone and name the own folder:\n%s", p.questions[0])
	}
}

func TestTheHintIsLeftOutWhereItDoesNotApply(t *testing.T) {
	linux := newMoveProbe(t, "linux", question.AnswerLeave)
	offerMove(linux.env)
	if strings.Contains(linux.questions[0], "administrator") {
		t.Errorf("the Linux question mentions an administrator:\n%s", linux.questions[0])
	}

	elevated := newMoveProbe(t, "windows", question.AnswerLeave)
	offerMove(elevated.env)
	if strings.Contains(elevated.questions[0], "administrator") {
		t.Errorf("the question mentions an administrator although the app already runs as one:\n%s", elevated.questions[0])
	}
}

func TestThereIsNoAdministratorTalkOnMacAndOnlyTwoChoicesOnLinux(t *testing.T) {
	mac := newMoveProbe(t, "darwin", question.AnswerLeave)
	offerMove(mac.env)
	if strings.Contains(mac.questions[0], "administrator") {
		t.Errorf("macOS question mentions an administrator:\n%s", mac.questions[0])
	}

	linux := newMoveProbe(t, "linux", question.AnswerLeave)
	offerMove(linux.env)
	if got := strings.Join(linux.choices[0], "|"); got != "Move it|Not now" {
		t.Errorf("Linux choices = %s", got)
	}
	if !strings.Contains(linux.questions[0], linux.targets()[0].Folder) {
		t.Errorf("the question should name the folder:\n%s", linux.questions[0])
	}
}

func TestLinuxMovesToItsOnlyPlaceOnYes(t *testing.T) {
	p := newMoveProbe(t, "linux", question.AnswerFirst)

	if !offerMove(p.env) || len(p.installed) != 1 || p.installed[0].Kind != kindLinuxUser {
		t.Errorf("installed = %+v", p.installed)
	}
}

func TestNotNowDoesNothingAndIsFollowedByAskingWhetherToBeAskedAgain(t *testing.T) {
	p := newMoveProbe(t, "windows", question.AnswerLeave, question.AnswerLeave)

	if offerMove(p.env) {
		t.Fatal("moved although the person chose not now")
	}
	if len(p.installed)+len(p.started) != 0 {
		t.Errorf("something was done: installed=%v started=%v", p.installed, p.started)
	}
	if len(p.questions) != 2 || strings.Join(p.choices[1], "|") != "Never ask again|Ask me again" {
		t.Fatalf("questions %q, choices %q: want the move, then whether to ask again", p.questions, p.choices)
	}
	if _, err := os.Stat(filepath.Join(p.env.AppDir, movedMarkerName)); err == nil {
		t.Error("the person said to ask again, so nothing should be remembered")
	}
}

func TestAskMeAgainIsAskedAtTheNextStart(t *testing.T) {
	p := newMoveProbe(t, "windows", question.AnswerLeave, question.AnswerLeave, question.AnswerLeave, question.AnswerLeave)

	offerMove(p.env)
	offerMove(p.env)

	if len(p.questions) != 4 {
		t.Errorf("asked %d questions, want 4: the move and the follow-up at each of two starts", len(p.questions))
	}
}

func TestNeverAskAgainIsRememberedAndNotAskedAtTheNextStart(t *testing.T) {
	p := newMoveProbe(t, "windows", question.AnswerLeave, question.AnswerFirst)

	offerMove(p.env)
	if _, err := os.Stat(filepath.Join(p.env.AppDir, movedMarkerName)); err != nil {
		t.Fatalf("the person said never, it should be remembered: %v", err)
	}

	if offerMove(p.env) || len(p.questions) != 2 {
		t.Errorf("asked %d questions in all, want the two of the first start only", len(p.questions))
	}
}

func TestMovingOnRequestWorksAfterNeverAskAgainAndRemembersNothing(t *testing.T) {
	p := newMoveProbe(t, "windows", question.AnswerLeave, question.AnswerFirst, question.AnswerSecond)
	offerMove(p.env) // not now, never ask again

	p.env.Force = true
	if got := moveOffer(p.env); got != offerMoved {
		t.Fatalf("got %v, want the move to happen although the person said never", got)
	}
	if len(p.installed) != 1 || p.installed[0].Kind != kindWindowsUser {
		t.Errorf("installed %v, want the person's own place", p.installed)
	}
	if got := strings.Join(p.choices[2], "|"); got != "For everyone|Just for me|Not now" {
		t.Errorf("choices %q", got)
	}
	if !strings.Contains(p.questions[2], "You asked for this from the preferences.") {
		t.Errorf("the question should say the person asked: %q", p.questions[2])
	}
}

func TestMovingOnRequestWhenAlreadyInPlaceSaysSoAndAsksNothing(t *testing.T) {
	p := newMoveProbe(t, "windows")
	p.env.Force = true
	p.env.Exe = p.targets()[0].Program

	if got := moveOffer(p.env); got != offerAlreadyThere || len(p.questions) != 0 {
		t.Errorf("got %v after %d questions, want already there with none", got, len(p.questions))
	}
}

func TestDecliningARequestedMoveIsNotFollowedByQuestionsAndRemembersNothing(t *testing.T) {
	p := newMoveProbe(t, "windows", question.AnswerLeave)
	p.env.Force = true

	if got := moveOffer(p.env); got != offerDeclined || len(p.questions) != 1 {
		t.Errorf("got %v after %d questions, want declined after one", got, len(p.questions))
	}
	if _, err := os.Stat(filepath.Join(p.env.AppDir, movedMarkerName)); err == nil {
		t.Error("a requested move leaves no record")
	}
}

func TestTheQuestionIsRecordedBeforeItIsPutSoACrashCannotMakeItRepeat(t *testing.T) {
	p := newMoveProbe(t, "windows")
	marker := filepath.Join(p.env.AppDir, movedMarkerName)
	p.env.Ask = func(string, string, []string) question.Answer {
		if _, err := os.Stat(marker); err != nil {
			t.Error("the marker did not exist while the question was on screen")
		}

		return question.AnswerLeave
	}

	offerMove(p.env)
}

func TestWhenNothingCouldShowTheQuestionItIsNotCountedAsAskedAndIsTriedAgainNextTime(t *testing.T) {
	p := newMoveProbe(t, "linux", question.AnswerUnavailable, question.AnswerLeave, question.AnswerLeave)

	if offerMove(p.env) {
		t.Fatal("moved with no answer")
	}
	if _, err := os.Stat(filepath.Join(p.env.AppDir, movedMarkerName)); err == nil {
		t.Error("the marker was kept although nothing was asked")
	}

	offerMove(p.env)

	if len(p.questions) != 3 {
		t.Errorf("asked %d times, want 3 (the second start asks the move and the follow-up again)", len(p.questions))
	}
}

func TestNothingHappensWhereTheOfferDoesNotApply(t *testing.T) {
	cases := map[string]func(*moveProbe){
		"a dev build":     func(p *moveProbe) { p.env.Version = "dev" },
		"no window":       func(p *moveProbe) { p.env.Window = false },
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
		p := newMoveProbe(t, "windows", question.AnswerFirst)
		change(p)

		if offerMove(p.env) || len(p.questions) != 0 {
			t.Errorf("%s: moved or asked (%v)", name, p.questions)
		}
	}
}

func TestWhenInstallingForEveryoneFailsThePersonIsOfferedTheirOwnPlaceInstead(t *testing.T) {
	p := newMoveProbe(t, "windows", question.AnswerFirst, question.AnswerFirst)
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
	p := newMoveProbe(t, "windows", question.AnswerFirst, question.AnswerLeave)
	p.installErrs = []error{errors.New("administrator approval was declined")}

	if offerMove(p.env) {
		t.Fatal("moved although the person declined both")
	}
	if len(p.installed) != 1 || len(p.started) != 0 || len(p.told) != 0 {
		t.Errorf("installed=%v started=%v told=%v", p.installed, p.started, p.told)
	}
}

func TestAFailureOfThePersonsOwnPlaceIsReportedAndTheAppKeepsRunningWhereItIs(t *testing.T) {
	p := newMoveProbe(t, "windows", question.AnswerSecond)
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
	p := newMoveProbe(t, "windows", question.AnswerFirst)
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
	p := newMoveProbe(t, "windows", question.AnswerSecond)
	p.startErr = errors.New("blocked by antivirus")

	if offerMove(p.env) {
		t.Fatal("reported a move although the copy did not start")
	}
	if len(p.told) != 1 || !strings.Contains(p.told[0], "blocked by antivirus") {
		t.Errorf("told = %v", p.told)
	}
}

func TestAnUnwritableMarkerMeansNoQuestionRatherThanAQuestionAtEveryStart(t *testing.T) {
	p := newMoveProbe(t, "windows", question.AnswerFirst)
	p.env.AppDir = filepath.Join(p.env.AppDir, "does", "not", "exist")

	if offerMove(p.env) || len(p.questions) != 0 {
		t.Errorf("asked although it could not remember having asked: %v", p.questions)
	}
}

// --- the installers, on real files ---------------------------------------------------

func writeProgram(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "downloaded", "mvd (1).exe")
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
	if err != nil || string(entry) != desktopEntryWithIcon(target.Program, linuxIconPath(menu)) {
		t.Errorf("the entry is not the one for %s: %q (%v)", target.Program, entry, err)
	}
}

func TestInstallingForEveryoneOnWindowsIsAPlainCopyPlusAShortcutForEveryAccount(t *testing.T) {
	exe := writeProgram(t, "program")
	target := installTargets("windows", installPlaces{ProgramFiles: t.TempDir()})[0]
	link := filepath.Join(t.TempDir(), "ProgramData", "Start Menu", "MVD.lnk")
	var links [][2]string
	var registered []installTarget

	err := installWindowsSystem(target, exe, link, func(l, to string) error {
		links = append(links, [2]string{l, to})

		return nil
	}, func(t installTarget) error {
		registered = append(registered, t)

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
	if len(registered) != 1 || registered[0] != target {
		t.Errorf("Settings > Apps entry = %v", registered)
	}
}

func TestInstallingForEveryoneFailsCleanlyWhenTheCopyFails(t *testing.T) {
	target := installTargets("windows", installPlaces{ProgramFiles: t.TempDir()})[0]

	registered := false

	err := installWindowsSystem(target, filepath.Join(t.TempDir(), "missing.exe"), "x.lnk", func(string, string) error { return nil },
		func(installTarget) error { registered = true; return nil })

	if err == nil {
		t.Error("expected an error")
	}
	if registered {
		t.Error("a program that was never installed must not be listed in Settings > Apps")
	}
}

func TestOnWindowsTheOwnProgramsFolderGetsTheProgramAndAShortcutButAMissingShortcutIsForgiven(t *testing.T) {
	exe := writeProgram(t, "program")
	target := installTargets("windows", installPlaces{LocalAppData: t.TempDir()})[0]
	startMenu := filepath.Join(t.TempDir(), "Start Menu")
	var links [][2]string
	registered := false

	err := installWindowsUser(target, exe, startMenu, func(link, to string) error {
		links = append(links, [2]string{link, to})

		return errors.New("no PowerShell")
	}, func(installTarget) error {
		registered = true

		return errors.New("no registry")
	})

	if err != nil {
		t.Fatalf("a missing shortcut or entry must not fail the install: %v", err)
	}
	if !registered {
		t.Error("the Settings > Apps entry was not attempted")
	}
	if len(links) != 1 || links[0] != [2]string{filepath.Join(startMenu, "MVD.lnk"), target.Program} {
		t.Errorf("shortcut = %v", links)
	}
	if data, _ := os.ReadFile(target.Program); string(data) != "program" {
		t.Errorf("program = %q", data)
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
