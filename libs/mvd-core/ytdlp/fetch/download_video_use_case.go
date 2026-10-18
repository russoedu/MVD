// Package fetch downloads a video with yt-dlp: the arguments and the output template it is
// started with, and the reading of its progress lines.
package fetch

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"youtube-downloader/libs/mvd-core/procwindow"
)

// Download runs yt-dlp with the given arguments and hands every line of its
// combined stdout and stderr to onLine. When yt-dlp fails, the returned
// error carries its last "ERROR:" line.
func Download(ctx context.Context, bin string, args []string, onLine func(line string)) error {
	return DownloadWithStartTimeout(ctx, bin, args, 0, onLine)
}

// DownloadWithStartTimeout is Download that gives up on a download which has not begun
// to fetch data (no progress line) within startTimeout: yt-dlp is stopped and the error
// says it timed out. 0 waits as long as it takes.
func DownloadWithStartTimeout(parent context.Context, bin string, args []string, startTimeout time.Duration, onLine func(line string)) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	var timedOut atomic.Bool
	var started atomic.Bool
	if startTimeout > 0 {
		timer := time.AfterFunc(startTimeout, func() {
			if !started.Load() {
				timedOut.Store(true)
				cancel()
			}
		})
		defer timer.Stop()
	}
	cmd := exec.CommandContext(ctx, bin, args...)
	procwindow.Hide(cmd)
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		_ = pw.Close()
		return fmt.Errorf("cannot start yt-dlp: %w", err)
	}

	lastError := ""
	scanDone := make(chan struct{})
	go func() {
		defer close(scanDone)
		scanner := bufio.NewScanner(pr)
		scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
		for scanner.Scan() {
			line := strings.TrimRight(scanner.Text(), "\r")
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "ERROR:") {
				lastError = strings.TrimSpace(strings.TrimPrefix(line, "ERROR:"))
			}
			if _, ok := ParseProgressLine(line); ok {
				started.Store(true)
			}
			onLine(line)
		}
	}()

	waitErr := cmd.Wait()
	_ = pw.Close()
	<-scanDone

	if waitErr != nil {
		if timedOut.Load() && parent.Err() == nil {
			return fmt.Errorf("timed out: the download did not start within %s", startTimeout)
		}
		if ctx.Err() != nil {
			return fmt.Errorf("cancelled")
		}
		if lastError != "" {
			return fmt.Errorf("%s", lastError)
		}
		return waitErr
	}
	return nil
}
