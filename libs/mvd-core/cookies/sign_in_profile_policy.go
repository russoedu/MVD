package cookies

import (
	"os"
	"path/filepath"
)

// SignInChoice is the value of the cookies setting that means "the login made inside MVD".
const SignInChoice = "mvd"

// SignInProfileDir is the folder the sign-in window of MVD keeps its own browser profile in,
// apart from the profile of the main window, whose cookie database is always open.
func SignInProfileDir(appDir string) string { return filepath.Join(appDir, "youtube-sign-in") }

// SignInSpec is the yt-dlp browser spec that reads the cookies of the sign-in profile: the
// folder of the cookie database that WebView2 keeps inside it.
func SignInSpec(appDir string) string {
	return "chromium:" + filepath.Join(SignInProfileDir(appDir), "EBWebView", "Default")
}

// HasSignIn reports whether a sign-in window has been used, so a profile exists to read.
func HasSignIn(appDir string) bool {
	_, err := os.Stat(filepath.Join(SignInProfileDir(appDir), "EBWebView", "Default", "Network", "Cookies"))

	return err == nil
}
