package musicfixer

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"youtube-downloader/libs/mvd-core/procwindow"
)

// Job is one video to give the music to.
type Job struct {
	// Exe is the music-fixer program; its folder must stay whole.
	Exe string
	// Video carries the song in low quality; Music is the high-quality copy.
	Video, Music string
	// OutDir receives <video name>-better.mp4 and result.json.
	OutDir string
	// FFmpeg is a folder holding ffmpeg and ffprobe, or the ffmpeg program; empty uses the PATH.
	FFmpeg string
}

// Result is what the app says it did.
type Result struct {
	// Better is the path of the new video.
	Better   string
	OffsetMs int
	Speed    float64
	Score    float64
}

// Fix runs the app and waits. onEvent, when not nil, gets the name of each action as it starts.
// Cancelling ctx asks the app to stop, and Fix returns ctx's error.
func Fix(ctx context.Context, job Job, onEvent func(action string)) (Result, error) {
	if err := os.MkdirAll(job.OutDir, 0o755); err != nil {
		return Result{}, err
	}
	resultFile := filepath.Join(job.OutDir, "result.json")
	args := []string{"--headless", "--videos", job.Video, "--input", "music=" + job.Music,
		"--out", resultFile, "--json-events"}
	if job.FFmpeg != "" {
		args = append(args, "--ffmpeg", job.FFmpeg)
	}
	cmd := exec.Command(job.Exe, args...)
	procwindow.Hide(cmd)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return Result{}, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Result{}, err
	}
	if err := cmd.Start(); err != nil {
		return Result{}, err
	}

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_, _ = io.WriteString(stdin, "cancel\n")
		case <-done:
		}
	}()
	runErr := readEvents(stdout, onEvent)
	waitErr := cmd.Wait()
	close(done)

	if ctx.Err() != nil {
		return Result{}, ctx.Err()
	}
	if runErr != nil {
		return Result{}, runErr
	}
	if waitErr != nil {
		return Result{}, fmt.Errorf("music-fixer failed: %w", waitErr)
	}
	return readResult(resultFile, job.OutDir)
}

// readEvents reads the event lines to the end and reports a run that did not work.
func readEvents(r io.Reader, onEvent func(string)) error {
	var failure error
	scan := bufio.NewScanner(r)
	scan.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for scan.Scan() {
		var e struct {
			Event  string `json:"event"`
			Action string `json:"action"`
			Error  any    `json:"error"`
			OK     *bool  `json:"ok"`
		}
		if json.Unmarshal(scan.Bytes(), &e) != nil {
			continue
		}
		switch {
		case e.Event == "action.started" && onEvent != nil:
			onEvent(e.Action)
		case e.Event == "video.finished" && e.Error != nil:
			failure = fmt.Errorf("music-fixer: %v", e.Error)
		case e.Event == "run.failed":
			failure = errors.New("music-fixer could not start the run")
		}
	}
	return failure
}

func readResult(file, outDir string) (Result, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Result{}, err
	}
	var records []struct {
		Better   string  `json:"better"`
		OffsetMs int     `json:"offsetMs"`
		Speed    float64 `json:"speed"`
		Score    float64 `json:"score"`
	}
	if err := json.Unmarshal(data, &records); err != nil || len(records) == 0 || strings.TrimSpace(records[0].Better) == "" {
		return Result{}, errors.New("music-fixer made no new video")
	}
	r := records[0]
	return Result{Better: filepath.Join(outDir, r.Better), OffsetMs: r.OffsetMs, Speed: r.Speed, Score: r.Score}, nil
}
