package install

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"youtube-downloader/apps/mvd/macbundle"
	"youtube-downloader/apps/mvd/programfile"
	"youtube-downloader/apps/mvd/question"
)

// movedMarkerName is the file in the app-data folder that records the person has been
// asked to move the app, whatever they answered.
const movedMarkerName = "install-offered"

// moveEnvironment is what moving the app needs from the machine, so that every branch
// can be tested without a screen or a real installation.
type moveEnvironment struct {
	GOOS      string
	Version   string
	Window    bool
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

	// Force is true when the person asked to move the app (from the preferences, or with
	// -move): they are asked whatever they answered before, and nothing is remembered.
	Force bool

	// Ask puts a question with the given choices, best one first, to the person.
	Ask func(title, question string, choices []string) question.Answer
	// Tell shows the person a problem.
	Tell func(text string)
	// Install puts the program in the target, with whatever else that system needs.
	Install func(target installTarget, exe string) error
	// Start starts the program that was installed.
	Start func(target installTarget, args []string) error
}

// offerOutcome is how an offer to move the app went.
type offerOutcome int

const (
	// offerSkipped: the offer does not apply (a developer's build, a script, asked and
	// told never again, already in place), so nobody was asked.
	offerSkipped offerOutcome = iota
	// offerMoved: the app was installed and the copy started; the caller should exit.
	offerMoved
	// offerDeclined: the person was asked and the app stays where it is.
	offerDeclined
	// offerAlreadyThere: asked for by the person, and the app is already in its place.
	offerAlreadyThere
	// offerCannot: there is nowhere to move to, or nothing could show the question.
	offerCannot
)

// offerMove is moveOffer for a caller that only wants to know whether the app moved.
func offerMove(env moveEnvironment) (moved bool) {
	return moveOffer(env) == offerMoved
}

// moveOffer asks the person, the first time the app is started from somewhere it does
// not belong, whether to move it, and where: for everyone on the computer, or just for
// them. If they agree it installs itself, starts the copy and reports offerMoved, and the
// caller should exit: the copy takes over, and removes this one once it has gone (see
// removeAfterExit).
//
// "Not now" is followed by one more question: ask again at the next start, or never.
// Never is remembered in the app-data folder. A person who asks to move the app (env.Force,
// from the preferences or -move) is asked whatever was answered before, and nothing is
// remembered.
//
// The question is recorded as asked before it is put, so that however the app ends it is
// not asked in a loop; the record is taken back when the person says to ask again, or
// when nothing could show the question. If installing for everyone does not work (the
// administrator prompt was declined, or the account may not write there) the person is
// offered the place of their own instead. If nothing works the app keeps running from
// where it is and says why.
func moveOffer(env moveEnvironment) offerOutcome {
	targets := installTargets(env.GOOS, env.Places)
	if len(targets) == 0 {
		return offerCannot
	}
	installed := isInstalled(env.GOOS, env.Exe, targets)
	marker := filepath.Join(env.AppDir, movedMarkerName)

	if env.Force {
		if installed {
			return offerAlreadyThere
		}
	} else {
		_, markerErr := os.Stat(marker)
		if !shouldOfferMove(moveSituation{
			Version: env.Version, Window: env.Window, MovedFrom: env.MovedFrom,
			Asked: markerErr == nil, HasTarget: true, Installed: installed,
		}) {
			return offerSkipped
		}
		if err := os.WriteFile(marker, []byte("asked\n"), 0o600); err != nil {
			// Not being able to remember would mean asking at every start.
			return offerSkipped
		}
	}

	prompt, choices := moveQuestion(filepath.Dir(env.Exe), targets, env.AdminHint, env.Force)
	picked := env.Ask("MVD", prompt, choices)
	switch picked {
	case question.AnswerUnavailable:
		// Nothing could be shown, so nothing was asked: try again at the next start.
		if !env.Force {
			_ = os.Remove(marker)
		}

		return offerCannot
	case question.AnswerLeave:
		if !env.Force {
			askAgainLater(env, marker)
		}

		return offerDeclined
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
			return offerDeclined
		}
		chosen = fallback
		err = env.Install(chosen, env.Exe)
	}
	if err != nil {
		env.Tell(fmt.Sprintf("MVD could not move itself to %s: %v\n\nIt will keep running from where it is.", chosen.Folder, err))

		return offerDeclined
	}

	args := append(append([]string{}, env.Args...), "-moved-from="+env.Exe)
	if err := env.Start(chosen, args); err != nil {
		env.Tell(fmt.Sprintf("MVD was installed in %s but could not be started from there: %v\n\nIt will keep running from where it is.", chosen.Folder, err))

		return offerDeclined
	}

	return offerMoved
}

// askAgainLater is the question after "Not now": whether to be asked again at the next
// start. Only "Never ask again" keeps the record that the person was asked; closing the
// box, or no way to show it, leaves things as they were and asks again.
func askAgainLater(env moveEnvironment, marker string) {
	picked := env.Ask("MVD",
		"Ask again the next time MVD starts?\n\nYou can always move it later: press m on the preferences, or start MVD with -move.",
		[]string{"Never ask again", "Ask me again"})
	if picked != question.AnswerFirst {
		_ = os.Remove(marker)
	}
}

// moveQuestion words the question and the choices. With two places to go it offers both
// and a way out; with one it offers that and a way out, and may add how to get the place
// for everyone.
func moveQuestion(from string, targets []installTarget, adminHint, requested bool) (string, []string) {
	text := fmt.Sprintf("MVD is running from:\n\n%s\n\n", from)
	footer := "You can also do this later, from the preferences."
	if requested {
		footer = "You asked for this from the preferences."
	}

	if len(targets) < 2 {
		text += fmt.Sprintf("Move it to its own folder, where this system keeps programs?\n\n%s\n\n", targets[0].Folder)
		if adminHint {
			text += "To install it for everyone on this computer instead, say No, then start MVD as an administrator " +
				"(right-click it and choose Run as administrator). You will not be asked again, so this is the only time to choose.\n\n"
		}
		text += footer

		return text, []string{"Move it", "Not now"}
	}

	text += "Where should it live?\n\n"
	text += fmt.Sprintf("For everyone on this computer:\n%s\n\n", targets[0].Folder)
	text += fmt.Sprintf("Just for you:\n%s\n\n", targets[1].Folder)
	text += footer

	return text, []string{"For everyone", "Just for me", "Not now"}
}

// installWindowsUser puts the program in the person's own Programs folder, adds a
// shortcut for them to the Start menu and lists it in Settings > Apps. A missing
// shortcut or entry is a loss of convenience, not a reason to undo the install.
func installWindowsUser(target installTarget, exe, startMenu string, shortcut func(link, target string) error, register func(installTarget) error) error {
	if err := programfile.Place(exe, target.Program); err != nil {
		return err
	}
	_ = shortcut(filepath.Join(startMenu, "MVD.lnk"), target.Program)
	_ = register(target)

	return nil
}

// installWindowsSystem puts the program in Program Files, adds a shortcut for every
// account and lists it in Settings > Apps. It is a plain copy: it only works, and is
// only offered, when the app was started as an administrator.
func installWindowsSystem(target installTarget, exe, allUsersLink string, shortcut func(link, target string) error, register func(installTarget) error) error {
	if err := programfile.Place(exe, target.Program); err != nil {
		return err
	}
	_ = shortcut(allUsersLink, target.Program)
	_ = register(target)

	return nil
}

// installMacBundle builds MVD.app around the program, with the app's icon.
func installMacBundle(target installTarget, exe, version string) error {
	return macbundle.WriteWithIcon(target.Folder, exe, version, macbundle.DefaultIcon())
}

// linuxIcon is the app's icon for the applications menu of Linux.
//
//go:embed linux_icon.png
var linuxIcon []byte

// installLinuxUser puts the program in ~/.local/bin and an entry for it, with its
// icon, in the applications menu.
func installLinuxUser(target installTarget, exe, applicationsDir string) error {
	if err := programfile.Place(exe, target.Program); err != nil {
		return err
	}
	if err := os.MkdirAll(applicationsDir, 0o755); err != nil {
		return err
	}

	// The icon is best effort: without it the entry still works and shows the generic one.
	icon := linuxIconPath(applicationsDir)
	if err := os.MkdirAll(filepath.Dir(icon), 0o755); err != nil || os.WriteFile(icon, linuxIcon, 0o644) != nil {
		icon = ""
	}

	return os.WriteFile(filepath.Join(applicationsDir, "mvd.desktop"), []byte(desktopEntryWithIcon(target.Program, icon)), 0o644)
}
