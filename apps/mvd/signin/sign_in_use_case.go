package signin

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"youtube-downloader/libs/mvd-core/cookies"
)

// Run opens the sign-in window as a program run of its own, waits until the person closes it,
// and says what the login came to: none, a login without Premium, or a Premium one. ytArgs are
// the extra arguments yt-dlp is run with (the JavaScript runtimes).
func Run(ctx context.Context, appDir, ytDlpPath string, ytArgs []string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if err := exec.CommandContext(ctx, exe, "-"+Flag, cookies.SignInProfileDir(appDir)).Run(); err != nil {
		return "", fmt.Errorf("the sign-in window did not run: %w", err)
	}
	signedIn, premium := cookies.CheckSignIn(ctx, ytDlpPath, cookies.SignInSpec(appDir), ytArgs)

	return outcomeMessage(signedIn, premium), nil
}

func outcomeMessage(signedIn, premium bool) string {
	switch {
	case !signedIn:
		return "No YouTube login was found. Try again and wait until YouTube Music has opened before closing the window."
	case premium:
		return "Signed in with a Premium account. Pick MVD sign-in in the setting above to download the music at the higher quality."
	}

	return "Signed in, but this account is not Premium, so the music keeps the normal quality. Pick MVD sign-in in the setting above to use the login anyway."
}
