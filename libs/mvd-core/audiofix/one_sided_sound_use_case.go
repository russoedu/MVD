package audiofix

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"youtube-downloader/libs/mvd-core/procwindow"
)

// FFmpegFor returns the ffmpeg to use: the one that sits beside yt-dlp (where the app
// keeps both), else the one on the PATH, else "".
func FFmpegFor(ytDlpPath string) string {
	name := "ffmpeg"
	if strings.HasSuffix(strings.ToLower(ytDlpPath), ".exe") {
		name += ".exe"
	}
	if ytDlpPath != "" {
		beside := filepath.Join(filepath.Dir(ytDlpPath), name)
		if _, err := os.Stat(beside); err == nil {
			return beside
		}
	}
	if found, err := exec.LookPath("ffmpeg"); err == nil {
		return found
	}
	return ""
}

// OneSided reports whether the file's sound plays on one side only: a stereo track
// whose other channel is silent. The file is only read.
func OneSided(ctx context.Context, ffmpeg, path string) (bool, error) {
	levels, err := channelLevels(ctx, ffmpeg, path)
	if err != nil {
		return false, err
	}
	_, oneSided := SilentChannel(levels)
	return oneSided, nil
}

// channelLevels asks ffmpeg for the RMS level of each channel of the first sound track.
func channelLevels(ctx context.Context, ffmpeg, path string) ([]float64, error) {
	cmd := exec.CommandContext(ctx, ffmpeg, "-hide_banner", "-nostats", "-i", path, "-map", "0:a:0",
		"-af", "astats=measure_overall=none:measure_perchannel=RMS_level", "-f", "null", "-")
	procwindow.Hide(cmd)
	var out bytes.Buffer
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg could not read the sound: %s", lastLine(out.String()))
	}
	return ParseChannelLevels(out.String()), nil
}

func lastLine(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}
