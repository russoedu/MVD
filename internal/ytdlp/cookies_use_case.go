package ytdlp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ExportCookies asks yt-dlp to read the browser's cookies and save them as
// a Netscape cookie file. probeURL is a cheap URL for yt-dlp to touch
// (one playlist entry is listed), since the jar is only written at the end
// of a run. Later runs can then use CookieArgs instead of decrypting the
// browser's cookie store again.
func ExportCookies(ctx context.Context, bin, browserSpec, file, probeURL string, extraArgs []string) error {
	args := []string{
		"--cookies-from-browser", browserSpec,
		"--cookies", file,
		"--flat-playlist", "--playlist-items", "1",
		"--simulate", "--skip-download", "--no-warnings", "--quiet",
	}
	args = append(args, extraArgs...)
	args = append(args, probeURL)

	cmd := exec.CommandContext(ctx, bin, args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("cannot export cookies from %s: %s", browserSpec, lastLine(msg))
	}
	return nil
}

// CookieArgs returns the yt-dlp arguments that load a cookie file.
func CookieArgs(file string) []string {
	if file == "" {
		return nil
	}
	return []string{"--cookies", file}
}
