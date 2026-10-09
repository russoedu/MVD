package ytdlp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"youtube-downloader/libs/mvd-core/procwindow"
)

// ProbeVideo asks yt-dlp what a video offers to download and what its storyboard is,
// without downloading it.
func ProbeVideo(ctx context.Context, bin, videoID string, extraArgs []string) (VideoProbe, error) {
	args := []string{"-J", "--no-playlist", "--skip-download", "--no-warnings"}
	args = append(args, extraArgs...)
	args = append(args, "https://www.youtube.com/watch?v="+videoID)

	cmd := exec.CommandContext(ctx, bin, args...)
	procwindow.Hide(cmd)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return VideoProbe{}, fmt.Errorf("yt-dlp -J failed: %s", lastLine(msg))
	}
	return ParseVideoProbe(out)
}
