package ytdlp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"youtube-downloader/libs/mvd-core/procwindow"
)

// VideoDetails asks yt-dlp for the title and the channel of a video, without downloading it.
func VideoDetails(ctx context.Context, bin, videoID string) (title, channel string, err error) {
	args := []string{"--print", "%(title)s", "--print", "%(channel)s", "--no-playlist", "--skip-download", "--no-warnings",
		"https://www.youtube.com/watch?v=" + videoID}
	cmd := exec.CommandContext(ctx, bin, args...)
	procwindow.Hide(cmd)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, runErr := cmd.Output()
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if runErr != nil || len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		msg := LastLine(strings.TrimSpace(stderr.String()))
		if msg == "" {
			msg = "no title"
		}

		return "", "", fmt.Errorf("yt-dlp could not read the video: %s", msg)
	}
	title = strings.TrimSpace(lines[0])
	if len(lines) > 1 && strings.TrimSpace(lines[1]) != "NA" {
		channel = strings.TrimSpace(lines[1])
	}

	return title, channel, nil
}
