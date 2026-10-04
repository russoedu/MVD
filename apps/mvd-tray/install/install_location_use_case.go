package install

import (
	"fmt"
	"os"
	"path/filepath"

	"youtube-downloader/apps/mvd-tray/programfile"
	"youtube-downloader/apps/mvd-tray/question"
)

// movedMarkerName is the file in the app-data folder that records the person has been
// asked to move the app, whatever they answered.
const movedMarkerName = "install-offered"

// moveEnvironment is what moving the app needs from the machine, so that every branch
// can be tested without a screen or a real installation.
type moveEnvironment struct {
	GOOS      string
	Version   string
	Tray      bool
	MovedFrom string
	Places    installPlaces
	// AdminHint adds to the question how to install for everyone, on a system where
	// that needs the app to be started as an administrator and it was not.
	AdminHint bool
	// AppDir is where the "already asked" marker is kept.
	AppDir string
	// Exe is the running program and Args the arguments it was started with.
	Exe  string
	Args []string

	// Ask puts a question with the given choices, best one first, to the person.
	Ask func(title, question string, choices []string) question.Answer
	// Tell shows the person a problem.
	Tell func(text string)
	// Install puts the program in the target, with whatever else that system needs.
	Install func(target installTarget, exe string) error
	// Start starts the program that was installed.
	Start func(target installTarget, args []string) error
}

// offerMove asks the person, the first time the app is started from somewhere it does
// not belong, whether to move it, and where: for everyone on the computer, or just for
// them. If they agree it installs itself, starts the copy and returns true, and the
// caller should exit: the copy takes over, and removes this one once it has gone (see
// removeAfterExit).
//
// The question is recorded as asked before it is put, so whatever the answer, and
// however the app ends, it is not asked again. If installing for everyone does not work
// (the administrator prompt was declined, or the account may not write there) the person
// is offered the place of their own instead. If nothing works the app keeps running from
// where it is and says why.
func offerMove(env moveEnvironment) (moved bool) {
	targets := installTargets(env.GOOS, env.Places)
	marker := filepath.Join(env.AppDir, movedMarkerName)
	_, markerErr := os.Stat(marker)

	if !shouldOfferMove(moveSituation{
		Version: env.Version, Tray: env.Tray, MovedFrom: env.MovedFrom,
		Asked: markerErr == nil, HasTarget: len(targets) > 0,
		Installed: isInstalled(env.GOOS, env.Exe, targets),
	}) {
		return false
	}
	if err := os.WriteFile(marker, []byte("asked\n"), 0o600); err != nil {
		// Not being able to remember would mean asking at every start.
		return false
	}

	prompt, choices := moveQuestion(filepath.Dir(env.Exe), targets, env.AdminHint)
	picked := env.Ask("MVD", prompt, choices)
	switch picked {
	case question.AnswerUnavailable:
		// Nothing could be shown, so nothing was asked: try again at the next start.
		_ = os.Remove(marker)

		return false
	case question.AnswerLeave:
		return false
	}

	chosen := targets[0]
	if picked == question.AnswerSecond && len(targets) > 1 {
		chosen = targets[1]
	}
	err := env.Install(chosen, env.Exe)
	if err != nil && chosen.Everyone && len(targets) > 1 {
		fallback := targets[1]
		again := env.Ask("MVD", fmt.Sprintf(
			"MVD could not be installed for everyone: %v\n\nInstall it just for you instead?\n\n%s", err, fallback.Folder),
			[]string{"Install just for me", "Leave it here"})
		if again != question.AnswerFirst {
			return false
		}
		chosen = fallback
		err = env.Install(chosen, env.Exe)
	}
	if err != nil {
		env.Tell(fmt.Sprintf("MVD could not move itself to %s: %v\n\nIt will keep running from where it is.", chosen.Folder, err))

		return false
	}

	args := append(append([]string{}, env.Args...), "-moved-from="+env.Exe)
	if err := env.Start(chosen, args); err != nil {
		env.Tell(fmt.Sprintf("MVD was installed in %s but could not be started from there: %v\n\nIt will keep running from where it is.", chosen.Folder, err))

		return false
	}

	return true
}

// moveQuestion words the question and the choices. With two places to go it offers both
// and a way out; with one it offers that and a way out, and may add how to get the place
// for everyone.
func moveQuestion(from string, targets []installTarget, adminHint bool) (string, []string) {
	text := fmt.Sprintf("MVD is running from:\n\n%s\n\n", from)

	if len(targets) < 2 {
		text += fmt.Sprintf("Move it to its own folder, where this system keeps programs?\n\n%s\n\n", targets[0].Folder)
		if adminHint {
			text += "To install it for everyone on this computer instead, say No, then start MVD as an administrator " +
				"(right-click it and choose Run as administrator). You will not be asked again, so this is the only time to choose.\n\n"
		}
		text += "You are only asked once."

		return text, []string{"Move it", "Leave it here"}
	}

	text += "Where should it live?\n\n"
	text += fmt.Sprintf("For everyone on this computer:\n%s\n\n", targets[0].Folder)
	text += fmt.Sprintf("Just for you:\n%s\n\n", targets[1].Folder)
	text += "You are only asked once."

	return text, []string{"For everyone", "Just for me", "Leave it here"}
}

// installWindowsUser puts the program in the person's own Programs folder and adds a
// shortcut for them to the Start menu. A missing shortcut is a loss of convenience, not
// a reason to undo the install.
func installWindowsUser(target installTarget, exe, startMenu string, shortcut func(link, target string) error) error {
	if err := programfile.Place(exe, target.Program); err != nil {
		return err
	}
	_ = shortcut(filepath.Join(startMenu, "MVD.lnk"), target.Program)

	return nil
}

// installWindowsSystem puts the program in Program Files and adds a shortcut for every
// account. It is a plain copy: it only works, and is only offered, when the app was
// started as an administrator.
func installWindowsSystem(target installTarget, exe, allUsersLink string, shortcut func(link, target string) error) error {
	if err := programfile.Place(exe, target.Program); err != nil {
		return err
	}
	_ = shortcut(allUsersLink, target.Program)

	return nil
}

// installMacBundle builds MVD.app around the program: the program itself, and the
// Info.plist that makes macOS treat the folder as an application.
func installMacBundle(target installTarget, exe, version string) error {
	if err := programfile.Place(exe, target.Program); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(target.Folder, "Contents", "Info.plist"), []byte(infoPlist(version)), 0o644)
}

// installLinuxUser puts the program in ~/.local/bin and an entry for it in the
// applications menu.
func installLinuxUser(target installTarget, exe, applicationsDir string) error {
	if err := programfile.Place(exe, target.Program); err != nil {
		return err
	}
	if err := os.MkdirAll(applicationsDir, 0o755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(applicationsDir, "mvd.desktop"), []byte(desktopEntry(target.Program)), 0o644)
}
