package ytdlp

import (
	"context"
	"os/exec"
	"strings"

	"youtube-downloader/libs/mvd-core/procwindow"
)

// premiumAudioIDs are the format ids YouTube Music serves at about 256 kbps, and only
// to a Premium login: 141 (AAC) and 774 (Opus).
var premiumAudioIDs = map[string]bool{"141": true, "774": true}

// HasPremiumAudio lists the formats of videoURL with the cookie file and reports whether
// they include the Premium audio ones. Any failure counts as "no".
func HasPremiumAudio(ctx context.Context, bin, cookieFile, videoURL string, extraArgs []string) bool {
	args := []string{"--cookies", cookieFile, "-F", "--no-warnings", "--quiet"}
	args = append(args, extraArgs...)
	cmd := exec.CommandContext(ctx, bin, append(args, videoURL)...)
	procwindow.Hide(cmd)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n") {
		if f := strings.Fields(line); len(f) > 0 && premiumAudioIDs[f[0]] {
			return true
		}
	}
	return false
}
