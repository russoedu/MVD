package runstate

import (
	"strings"
	"testing"
)

func TestTheCommonFailuresAreSaidInEverydayWords(t *testing.T) {
	cases := map[string]string{
		"yt-dlp --flat-playlist failed: ERROR: [generic] 'hello this is not a link' is not a valid URL": "not a link MVD can use",
		"yt-dlp --flat-playlist failed: ERROR: [youtube] THISISNOTAR: This video is unavailable":      "unavailable",
		"ERROR: [youtube] abcdefghijk: Sign in to confirm your age. This video may be inappropriate":    "age-restricted",
		"ERROR: [youtube] abcdefghijk: Sign in to confirm you’re not a bot":                           "wants to be sure",
		"ERROR: [youtube] abcdefghijk: Private video. Sign in if you've been granted access":            "private",
		"ERROR: [youtube] abcdefghijk: The uploader has not made this video available in your country": "your country",
		"ERROR: unable to download video data: HTTP Error 429: Too Many Requests":                       "limiting requests",
		"ERROR: Unable to download webpage: <urlopen error [Errno 11001] getaddrinfo failed>":           "no connection",
	}
	for raw, want := range cases {
		got := PlainError(raw)
		if !strings.Contains(strings.ToLower(got), strings.ToLower(want)) {
			t.Errorf("%q\n  said %q, want something with %q", raw, got, want)
		}
		if strings.Contains(got, "yt-dlp") || strings.Contains(got, "--flat-playlist") {
			t.Errorf("the tool's name is still in %q", got)
		}
	}
}

func TestAnUnknownFailureKeepsItsMessageButLosesTheToolsWrapping(t *testing.T) {
	raw := "yt-dlp --flat-playlist failed: ERROR: [youtube] jNQXAC9IVRw: Something new went wrong"
	if got := PlainError(raw); got != "Something new went wrong" {
		t.Errorf("got %q", got)
	}
	if got := PlainError("disk full"); got != "Disk full" {
		t.Errorf("got %q", got)
	}
}

func TestNothingStaysNothing(t *testing.T) {
	if got := PlainError("   "); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestAMessageThatIsOnlyWrappingIsNotErased(t *testing.T) {
	if got := PlainError("ERROR: [generic]"); got == "" {
		t.Error("a message was erased")
	}
}
