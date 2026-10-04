package uninstall

import (
	"os"
	"path/filepath"
	"runtime"

	"youtube-downloader/apps/mvd/console"
	"youtube-downloader/apps/mvd/install"
	"youtube-downloader/apps/mvd/notification"
	"youtube-downloader/apps/mvd/question"
)

// Here is the service against the real machine. configPath is the settings file, which
// says where the person's downloads are, and quit ends the app once it has been removed
// (it may be nil, where the process ends anyway).
func Here(version, appDir, configPath string, quit func()) Service {
	exe, _ := os.Executable()
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}

	return Service{Env: Environment{
		GOOS: runtime.GOOS, Version: version, Exe: exe, AppDir: appDir,
		Keep:      keptFolders(configPath, appDir),
		Footprint: install.FootprintHere,
		Ask:       question.Ask,
		Inform:    inform,
		Exists: func(path string) bool {
			_, err := os.Stat(path)

			return err == nil
		},
		ListDir:            listNames,
		RemoveAll:          os.RemoveAll,
		Remove:             os.Remove,
		RemoveProgram:      removeRunningProgram,
		RemoveRegistration: install.RemoveUninstallEntries,
		Quit:               quit,
	}}
}

// inform tells the person how the removal went: in a message box where the app has no
// terminal (Windows), and in a notification elsewhere, where the app may be gone by the
// time a dialog could be read.
func inform(text string) {
	if runtime.GOOS == "windows" {
		console.ShowInfo(text)

		return
	}
	notification.Notify(notification.Notice{Title: "MVD", Text: text})
}

func listNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names, nil
}
