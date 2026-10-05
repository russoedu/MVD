package terminalui

import (
	"context"
	"fmt"
	"os"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runner"
	"youtube-downloader/libs/mvd-core/tui"
)

// engineRun is a download run the app model can close: Close cancels it and
// waits until the engine has stopped and its events are drained, as mvd-tui does.
// It is the engine itself, so the download screen can also add links to it.
type engineRun struct {
	*engine.Engine
	cancel context.CancelFunc
	events <-chan interface{}
	done   chan struct{}
}

func (r engineRun) Close() {
	r.cancel()
	for range r.events {
	}
	<-r.done
}

// StartDownloadRun builds the engine for the settings and list and runs it, for
// the app model's start key.
func StartDownloadRun(parent context.Context, ytDlpPath string, logf func(string, ...interface{})) tui.RunStarter {
	return func(cfg config.Config, urls []string) (tui.Run, error) {
		if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
			return nil, fmt.Errorf("cannot create the download folder %s: %w", cfg.OutputDir, err)
		}
		ctx, cancel := context.WithCancel(parent)
		eng, err := runner.BuildEngine(ctx, ytDlpPath, cfg, urls, logf)
		if err != nil {
			cancel()
			return nil, err
		}
		done := make(chan struct{})
		go func() { eng.Run(ctx); close(done) }()

		return engineRun{Engine: eng, cancel: cancel, events: eng.Events(), done: done}, nil
	}
}
