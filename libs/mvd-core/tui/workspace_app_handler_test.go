package tui

import (
	"path/filepath"
	"testing"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/engine"
)

func TestPreferencesApplyAtOnceExceptWhereFilesGo(t *testing.T) {
	cfg := config.Config{OutputDir: "/music", MaxConcurrentDownloads: 2}
	var live engine.LiveSettings
	a := NewWorkspaceApp(WorkspaceInput{
		Setup: SetupInput{Cfg: cfg, CfgPath: filepath.Join(t.TempDir(), "config.conf")},
		Live: func(c config.Config) engine.LiveSettings {
			live = engine.LiveSettings{Workers: c.MaxConcurrentDownloads}
			return live
		},
	}).(workspaceApp)

	changed := cfg
	changed.MaxConcurrentDownloads = 5
	changed.OutputDir = "/elsewhere"
	a = a.apply(changed, false)
	if a.cfg.MaxConcurrentDownloads != 5 || live.Workers != 5 {
		t.Errorf("the number of downloads was not applied: %d, %d", a.cfg.MaxConcurrentDownloads, live.Workers)
	}
	if a.cfg.OutputDir != "/music" {
		t.Errorf("the folder changed before it was confirmed: %q", a.cfg.OutputDir)
	}
	a = a.apply(changed, true)
	if a.cfg.OutputDir != "/elsewhere" {
		t.Errorf("the folder was not applied once confirmed: %q", a.cfg.OutputDir)
	}
}
