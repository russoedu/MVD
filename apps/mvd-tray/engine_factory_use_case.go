package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"youtube-downloader/libs/mvd-core/appdir"
	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/runner"
	"youtube-downloader/libs/mvd-server/session"
)

// newEngineFactory returns what the session calls when the first URLs arrive.
//
// The config is read then, not at start-up, so a setting changed since the app
// opened is the one the run uses.
func newEngineFactory(ytDlpPath, appDir string, logf func(string, ...interface{})) session.Factory {
	return func(ctx context.Context, urls []string) (session.Engine, error) {
		cfg, _, err := config.LoadOrCreate(filepath.Join(appDir, "config.conf"), appDir, appdir.DefaultDownloadsDir())
		if err != nil {
			return nil, fmt.Errorf("cannot read the settings: %w", err)
		}
		if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
			return nil, fmt.Errorf("cannot create the download folder %s: %w", cfg.OutputDir, err)
		}
		return runner.BuildEngine(ctx, ytDlpPath, cfg, urls, logf)
	}
}
