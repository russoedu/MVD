package ytdlp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ListPlaylist asks yt-dlp for the contents of a playlist without
// downloading anything.
func ListPlaylist(ctx context.Context, bin, playlistURL string, extraArgs []string) ([]PlaylistEntry, error) {
	args := []string{"--flat-playlist", "-j", "--no-warnings"}
	args = append(args, extraArgs...)
	args = append(args, playlistURL)

	cmd := exec.CommandContext(ctx, bin, args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("yt-dlp --flat-playlist failed: %s", lastLine(msg))
	}
	return ParsePlaylistEntries(strings.NewReader(string(out)))
}

func lastLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.LastIndex(s, "\n"); i >= 0 {
		return strings.TrimSpace(s[i+1:])
	}
	return s
}
