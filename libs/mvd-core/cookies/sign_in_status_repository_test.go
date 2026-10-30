package cookies

import "testing"

func TestSignInStateIsKeptBetweenRuns(t *testing.T) {
	app := t.TempDir()
	if got := SignInState(app); got != "" {
		t.Fatalf("before any sign-in the state is %q, want none kept", got)
	}
	for _, c := range []struct {
		signedIn, premium bool
		want              string
	}{{true, true, SignInPremium}, {true, false, SignInFree}, {false, false, SignInNone}} {
		SaveSignInState(app, c.signedIn, c.premium)
		if got := SignInState(app); got != c.want {
			t.Errorf("signedIn=%v premium=%v: state %q, want %q", c.signedIn, c.premium, got, c.want)
		}
	}
}
