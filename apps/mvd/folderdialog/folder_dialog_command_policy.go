package folderdialog

import (
	"strings"

	"youtube-downloader/apps/mvd/oscommand"
)

// folderDialogCommand picks the way to show a folder chooser on goos, using only the
// programs has reports as installed. ok is false when there is none. Windows is not here: it
// shows its windows itself (see native_dialog_client_windows.go).
func folderDialogCommand(goos, start string, has func(string) bool) (cmd oscommand.Command, ok bool) {
	switch goos {
	case "darwin":
		if !has("osascript") {
			return oscommand.Command{}, false
		}
		// The folder is argv, not part of the AppleScript text.
		script := []string{
			"-e", "on run argv",
			"-e", `if (count of argv) > 0 then`,
			"-e", `return POSIX path of (choose folder with prompt "Choose a folder" default location (POSIX file (item 1 of argv)))`,
			"-e", "end if",
			"-e", `return POSIX path of (choose folder with prompt "Choose a folder")`,
			"-e", "end run",
		}
		if start != "" {
			script = append(script, start)
		}

		return oscommand.Command{Name: "osascript", Args: script}, true
	default:
		switch {
		case has("zenity"):
			args := []string{"--file-selection", "--directory", "--title=Choose a folder"}
			if start != "" {
				args = append(args, "--filename="+strings.TrimRight(start, "/")+"/")
			}

			return oscommand.Command{Name: "zenity", Args: args}, true
		case has("kdialog"):
			args := []string{"--getexistingdirectory"}
			if start != "" {
				args = append(args, start)
			}

			return oscommand.Command{Name: "kdialog", Args: args}, true
		}

		return oscommand.Command{}, false
	}
}
