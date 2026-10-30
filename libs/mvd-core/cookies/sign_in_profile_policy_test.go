package cookies

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSignInSpecPointsAtTheWebViewCookieFolder(t *testing.T) {
	spec := SignInSpec(filepath.Join("app"))
	want := filepath.Join("app", "youtube-sign-in", "EBWebView", "Default")
	if spec != "chromium:"+want {
		t.Errorf("spec = %q, want chromium:%s", spec, want)
	}
}

func TestHasSignInNeedsTheCookieDatabase(t *testing.T) {
	app := t.TempDir()
	if HasSignIn(app) {
		t.Fatal("no sign-in was made yet")
	}
	db := filepath.Join(SignInProfileDir(app), "EBWebView", "Default", "Network", "Cookies")
	if err := os.MkdirAll(filepath.Dir(db), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(db, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !HasSignIn(app) {
		t.Error("the profile has a cookie database, so a sign-in exists")
	}
	if !strings.HasPrefix(SignInSpec(app), "chromium:") {
		t.Error("the spec is read as a Chromium profile")
	}
}
