package signin

import (
	"strings"
	"testing"
)

func TestOutcomeMessageSaysWhatTheLoginIs(t *testing.T) {
	cases := []struct {
		signedIn, premium bool
		want              string
	}{
		{false, false, "No YouTube login"},
		{true, false, "not Premium"},
		{true, true, "Premium account"},
	}
	for _, c := range cases {
		if got := outcomeMessage(c.signedIn, c.premium); !strings.Contains(got, c.want) {
			t.Errorf("outcomeMessage(%v, %v) = %q, want it to say %q", c.signedIn, c.premium, got, c.want)
		}
	}
}
