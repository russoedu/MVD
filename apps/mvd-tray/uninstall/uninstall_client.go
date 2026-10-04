package uninstall

import (
	"os"
	"path/filepath"
	"runtime"

	"youtube-downloader/apps/mvd-tray/console"
	"youtube-downloader/apps/mvd-tray/install"
	"youtube-downloader/apps/mvd-tray/notification"
	"youtube-downloader/apps/mvd-tray/question"
)

// Here is the service against the real machine. keep says which folders hold the
// person's own files, quit ends the app once it has been removed, and it may be nil.
func Here(version, appDir string, keep func() []string, quit func()) Service {
	exe, _ := os.Executable()
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}

	return Service{Env: Environment{
		GOOS: runtime.GOOS, Version: version, Exe: exe, AppDir: appDir,
		Keep:      keep,
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
