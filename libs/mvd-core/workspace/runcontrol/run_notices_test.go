package runcontrol

import "testing"

func TestAcceptedSongsSayTheyNeedDownloadWhileTheDownloadsAreHeld(t *testing.T) {
	cases := []struct {
		n    int
		held bool
		want string
	}{
		{1, true, "1 song accepted: press Download to start it"},
		{3, true, "3 songs accepted: press Download to start them"},
		{1, false, "1 song accepted: it will download"},
		{3, false, "3 songs accepted: they will download"},
	}
	for _, c := range cases {
		if got := AcceptedNotice(c.n, c.held); got != c.want {
			t.Errorf("AcceptedNotice(%d, %v) = %q, want %q", c.n, c.held, got, c.want)
		}
	}
}
