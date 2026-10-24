package folderdialog

import (
	"strings"

	"youtube-downloader/apps/mvd/oscommand"
)

// fileDialogCommand picks the way to show an open-file window on goos, using only the
// programs has reports as installed. extension (without its dot) and label describe the
// files to show first; every file can be chosen too. ok is false when there is none.
func fileDialogCommand(goos, start, extension, label string, has func(string) bool) (cmd oscommand.Command, ok bool) {
	switch goos {
	case "darwin":
		if !has("osascript") {
			return oscommand.Command{}, false
		}
		// The folder is argv, not part of the AppleScript text.
		script := []string{
			"-e", "on run argv",
			"-e", `if (count of argv) > 0 then`,
			"-e", `return POSIX path of (choose file of type {"` + extension + `"} with prompt "Open a file" default location (POSIX file (item 1 of argv)))`,
			"-e", "end if",
			"-e", `return POSIX path of (choose file of type {"` + extension + `"} with prompt "Open a file")`,
			"-e", "end run",
		}
		if start != "" {
			script = append(script, start)
		}

		return oscommand.Command{Name: "osascript", Args: script}, true
	default:
		switch {
		case has("zenity"):
			args := []string{"--file-selection", "--title=Open a file",
				"--file-filter=" + label + " | *." + extension, "--file-filter=All files | *"}
			if start != "" {
				args = append(args, "--filename="+strings.TrimRight(start, "/")+"/")
			}

			return oscommand.Command{Name: "zenity", Args: args}, true
		case has("kdialog"):
			dir := start
			if dir == "" {
				dir = "."
			}

			return oscommand.Command{Name: "kdialog", Args: []string{"--getopenfilename", dir, "*." + extension + "|" + label}}, true
		}

		return oscommand.Command{}, false
	}
}
