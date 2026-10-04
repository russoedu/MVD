package ytdlp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Download runs yt-dlp with the given arguments and hands every line of its
// combined stdout and stderr to onLine. When yt-dlp fails, the returned
// error carries its last "ERROR:" line.
func Download(ctx context.Context, bin string, args []string, onLine func(line string)) error {
	cmd := exec.CommandContext(ctx, bin, args...)
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
			onLine(line)
		}
	}()

	waitErr := cmd.Wait()
	_ = pw.Close()
	<-scanDone

	if waitErr != nil {
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
