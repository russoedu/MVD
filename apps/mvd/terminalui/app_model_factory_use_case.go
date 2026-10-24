// Package terminalui serves mvd's terminal interface (the setup screens and the
// download screen, the same ones mvd-tui shows) to a browser window, through
// TReactUI.
package terminalui

import (
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/appdir"
	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/openurl"
	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/runner"
	"youtube-downloader/libs/mvd-core/sourcelist"
	"youtube-downloader/libs/mvd-core/tui"
	"youtube-downloader/libs/mvd-core/workspace"
)

// Files are where the settings and the download list live, the same files the
// terminal app and the settings page use.
type Files struct {
	Config, List string
	// Review, when set, is a playlist address or a saved .mvd file the first window opens
	// in the playlist editor, instead of on the list.
	Review string
}

// Host is what the app needs from the machine it runs on.
type Host struct {
	// Start starts a download run from the screens.
	Start tui.RunStarter
	// StartList starts the run that carries the download list on.
	StartList func(config.Config, []string, []engine.PlannedEntry, bool) (workspace.ListRun, error)
	// PickFolder is the operating system's own folder chooser, which the folder
	// settings open; nil to use the built-in browser.
	PickFolder tui.FolderPicker
	// Uninstall removes the app after asking; nil to offer no uninstall.
	Uninstall tui.Uninstaller
	// Move moves the app to its own folder after asking; nil to offer no move.
	Move tui.Mover
	// Plan and PlanDownload run the playlist editor's plan-only run and the download of
	// what it chose; Describe names a video the person picked. Each may be nil.
	Plan         tui.RunStarter
	PlanDownload tui.PlanDownloadStarter
	Describe     func(videoID string) (string, error)
	// PickFile is the operating system's own open-file window, which the editor's open
	// key shows; nil to ask for the file on the screen's bottom line.
	PickFile func(start string) (string, bool, error)
	// SaveFile is the operating system's own save-file window, which the editor's save
	// key shows; nil to ask for the path on the screen's bottom line.
	SaveFile func(dir, name string) (string, bool, error)
	// Quit is called when the person quits the app from its screens, so the window that
	// shows them closes with it; nil for none.
	Quit func()
	// Version is the version of the app, for its About window.
	Version string
	// Clipboard is the machine's clipboard; nil to use the terminal's own.
	Clipboard workspace.Clipboard
	// Modal is told when a popup opens and closes over the screen; nil for none.
	Modal func(open bool)
}

// NewModelFactory returns what the server calls when the first window connects
// (and again after the user quits the app from it). Each start reads the files
// again, so a change made elsewhere since is the one it shows.
func NewModelFactory(appDir string, files Files, host Host, logf func(string, ...interface{})) func() tea.Model {
	review := files.Review // only the first window opens on it
	return func() tea.Model {
		cfg, created, err := config.LoadOrCreate(files.Config, appDir, appdir.DefaultVideosDir())
		if err != nil {
			logf("cannot read the settings %s, starting from the defaults: %v", files.Config, err)
			cfg, created = config.Default(appDir, appdir.DefaultVideosDir()), true
		}
		tui.ApplyTheme(cfg.Colors)
		urls, _ := sourcelist.Load(files.List)
		list := downloadlist.New(urls)
		if review != "" {
			if strings.EqualFold(filepath.Ext(review), playlistfile.Extension) {
				if opened, err := downloadlist.Open(review); err == nil {
					list = opened
				} else {
					logf("cannot open %s: %v", review, err)
				}
			} else {
				list = downloadlist.New([]string{review})
			}
			review = ""
		}

		return accessibleApp{tui.NewWorkspaceApp(tui.WorkspaceInput{
			Setup: tui.SetupInput{Cfg: cfg, URLs: urls, CfgPath: files.Config, ListPath: files.List, OpenConfig: created, PickFolder: host.PickFolder, Uninstall: host.Uninstall, Move: host.Move},
			List:  list,
			Live:  runner.LiveSettings,
			Modal: host.Modal,
			Host: workspace.Host{
				Start: host.StartList,
				Editor: tui.EditorHost{
					Open:        openurl.Open,
					Describe:    host.Describe,
					SessionFile: filepath.Join(appDir, "plan"+playlistfile.Extension),
					SaveDir:     filepath.Join(appDir, "plans"),
					DecisionLog: filepath.Join(appDir, "editor-decisions.jsonl"),
				},
				ListDir:  filepath.Join(appDir, "lists"),
				PickFile: host.PickFile,
				SaveFile: host.SaveFile,
				Open:     openurl.Open,
				Quit:     host.Quit,
				Version:  host.Version,
				Clipboard: host.Clipboard,
			},
		})}
	}
}
