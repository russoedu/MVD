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

// Repair looks at the sound of the file at path and, when only one side plays, rewrites
// the file so the audible channel plays on both sides. The picture is copied as it is.
// It returns true when the file was rewritten, and leaves the file alone on any error.
func Repair(ctx context.Context, ffmpeg, path string) (bool, error) {
	levels, err := channelLevels(ctx, ffmpeg, path)
	if err != nil {
		return false, err
	}
	silent, ok := SilentChannel(levels)
	if !ok {
		return false, nil
	}

	audible := 1 - silent
	ext := filepath.Ext(path)
	tmp := strings.TrimSuffix(path, ext) + ".mvd-fixing" + ext
	args := []string{"-hide_banner", "-nostats", "-y", "-i", path, "-map", "0", "-map_metadata", "0",
		"-c", "copy", "-c:a", audioCodecFor(ext), "-b:a", "192k",
		"-af", fmt.Sprintf("pan=stereo|c0=c%d|c1=c%d", audible, audible), tmp}
	cmd := exec.CommandContext(ctx, ffmpeg, args...)
	procwindow.Hide(cmd)
	var out bytes.Buffer
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		_ = os.Remove(tmp)
		return false, fmt.Errorf("ffmpeg could not rewrite the sound: %s", lastLine(out.String()))
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return false, err
	}
	return true, nil
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

// audioCodecFor is the codec a file of this kind holds its sound in once it is rewritten.
func audioCodecFor(ext string) string {
	switch strings.ToLower(ext) {
	case ".mp3":
		return "libmp3lame"
	case ".webm", ".opus", ".ogg":
		return "libopus"
	}
	return "aac"
}

func lastLine(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}
