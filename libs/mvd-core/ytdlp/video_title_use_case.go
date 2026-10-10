package ytdlp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"youtube-downloader/libs/mvd-core/procwindow"
)

// VideoTitle asks yt-dlp for the title of a video, without downloading it.
func VideoTitle(ctx context.Context, bin, videoID string, extraArgs []string) (string, error) {
	args := []string{"--print", "%(title)s", "--no-playlist", "--skip-download", "--no-warnings"}
	args = append(args, extraArgs...)
	args = append(args, "https://www.youtube.com/watch?v="+videoID)

	cmd := exec.CommandContext(ctx, bin, args...)
	procwindow.Hide(cmd)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	title := strings.TrimSpace(string(out))
	if err != nil || title == "" {
		msg := lastLine(strings.TrimSpace(stderr.String()))
		if msg == "" {
			msg = "no title"
		}
		return "", fmt.Errorf("yt-dlp could not read the video: %s", msg)
	}
	return title, nil
}
