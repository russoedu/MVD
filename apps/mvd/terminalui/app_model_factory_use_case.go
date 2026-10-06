// Package terminalui serves mvd's terminal interface (the setup screens and the
// download screen, the same ones mvd-tui shows) to a browser window, through
// TReactUI.
package terminalui

import (
	tea "github.com/charmbracelet/bubbletea"

	"youtube-downloader/libs/mvd-core/appdir"
	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/sourcelist"
	"youtube-downloader/libs/mvd-core/tui"
)

// Files are where the settings and the download list live, the same files the
// terminal app and the settings page use.
type Files struct {
	Config, List string
}

// Host is what the app needs from the machine it runs on.
type Host struct {
	// Start starts a download run from the screens.
	Start tui.RunStarter
	// PickFolder is the operating system's own folder chooser, which the folder
	// settings open; nil to use the built-in browser.
	PickFolder tui.FolderPicker
	// Uninstall removes the app after asking; nil to offer no uninstall.
	Uninstall tui.Uninstaller
}

// NewModelFactory returns what the server calls when the first window connects
// (and again after the user quits the app from it). Each start reads the files
// again, so a change made elsewhere since is the one it shows.
func NewModelFactory(appDir string, files Files, host Host, logf func(string, ...interface{})) func() tea.Model {
	return func() tea.Model {
		cfg, created, err := config.LoadOrCreate(files.Config, appDir, appdir.DefaultDownloadsDir())
		if err != nil {
			logf("cannot read the settings %s, starting from the defaults: %v", files.Config, err)
			cfg, created = config.Default(appDir, appdir.DefaultDownloadsDir()), true
		}
		urls, _ := sourcelist.Load(files.List)

		return accessibleApp{tui.NewAppModel(tui.AppInput{
			Setup: tui.SetupInput{Cfg: cfg, URLs: urls, CfgPath: files.Config, ListPath: files.List, OpenConfig: created, PickFolder: host.PickFolder, Uninstall: host.Uninstall},
			Start: host.Start,
		})}
	}
}
