package cookies

import (
	"os"
	"path/filepath"
	"strings"
)

// Sign-in states, as kept between runs.
const (
	SignInNone    = "none"    // the window was used but no YouTube login was found
	SignInFree    = "free"    // a login without Premium
	SignInPremium = "premium" // a Premium login
)

func signInStatusFile(appDir string) string { return filepath.Join(appDir, "youtube-sign-in.status") }

// SaveSignInState keeps what the last sign-in came to, so the preferences can show it later.
func SaveSignInState(appDir string, signedIn, premium bool) {
	state := SignInNone
	switch {
	case signedIn && premium:
		state = SignInPremium
	case signedIn:
		state = SignInFree
	}
	_ = os.WriteFile(signInStatusFile(appDir), []byte(state), 0o600)
}

// SignInState is what the last sign-in came to: "" when none was made.
func SignInState(appDir string) string {
	data, err := os.ReadFile(signInStatusFile(appDir))
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}
