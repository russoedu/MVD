package install

import (
	"fmt"
	"os"
	"path/filepath"

	"youtube-downloader/apps/mvd/question"
)

// rememberInstalled records in the marker file which version was put in place.
func rememberInstalled(marker, version string) {
	_ = os.WriteFile(marker, []byte(markerState{InstalledVersion: version}.String()), 0o600)
}

// updateOffer is the question put when the app is started from somewhere other than its
// own folder, after it was moved there once: when a different version is installed,
// replace it with the one that is running. It answers offerMoved when the installed copy
// was replaced and started (the caller should exit), offerSkipped when it does not apply,
// and offerDeclined when the person said not now or the copy could not be replaced.
func updateOffer(env moveEnvironment, targets []installTarget, installed bool, state markerState, marker string) offerOutcome {
	var present *installTarget
	for i := range targets {
		if _, err := os.Stat(targets[i].Program); err == nil {
			present = &targets[i]
			break
		}
	}

	if !shouldOfferUpdate(updateSituation{
		Version: env.Version, Window: env.Window, MovedFrom: env.MovedFrom,
		Installed: installed, HasInstalledCopy: present != nil, Marker: state,
	}) {
		return offerSkipped
	}

	have := "another version"
	if state.InstalledVersion != "" {
		have = "version " + state.InstalledVersion
	}
	prompt := fmt.Sprintf("MVD %s is running from:\n\n%s\n\nYou have %s of MVD installed in:\n\n%s\n\nReplace the installed copy with this one?",
		env.Version, filepath.Dir(env.Exe), have, present.Folder)
	switch env.Ask("MVD", prompt, []string{"Update it", "Not now"}) {
	case question.AnswerUnavailable:
		// Nothing could be shown, so nothing was asked: try again at the next start.
		return offerCannot
	case question.AnswerFirst:
	default:
		state.DeclinedUpdate = env.Version
		_ = os.WriteFile(marker, []byte(state.String()), 0o600)

		return offerDeclined
	}

	if err := env.Install(*present, env.Exe); err != nil {
		env.Tell(fmt.Sprintf("MVD could not update the copy in %s: %v\n\nIf MVD is open, close it and start this one again. It will keep running from where it is.", present.Folder, err))

		return offerDeclined
	}
	rememberInstalled(marker, env.Version)

	args := append(append([]string{}, env.Args...), "-moved-from="+env.Exe)
	if err := env.Start(*present, args); err != nil {
		env.Tell(fmt.Sprintf("MVD was updated in %s but could not be started from there: %v\n\nIt will keep running from where it is.", present.Folder, err))

		return offerDeclined
	}

	return offerMoved
}
