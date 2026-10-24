package folderdialog

import (
	"path/filepath"
	"strings"

	"youtube-downloader/apps/mvd/oscommand"
)

// saveDialogCommand picks the way to show a save-file window on goos, using only the
// programs has reports as installed. name is the file name offered (with its extension),
// start the folder it opens in; extension (without its dot) and label describe the file
// type. ok is false when there is none.
func saveDialogCommand(goos, start, name, extension, label string, has func(string) bool) (cmd oscommand.Command, ok bool) {
	switch goos {
	case "darwin":
		if !has("osascript") {
			return oscommand.Command{}, false
		}
		// The folder and the name are argv, not part of the AppleScript text.
		script := []string{
			"-e", "on run argv",
			"-e", `if (count of argv) > 1 then`,
			"-e", `return POSIX path of (choose file name with prompt "Save as" default name (item 2 of argv) default location (POSIX file (item 1 of argv)))`,
			"-e", "end if",
			"-e", `return POSIX path of (choose file name with prompt "Save as")`,
			"-e", "end run",
		}
		if start != "" && name != "" {
			script = append(script, start, name)
		}

		return oscommand.Command{Name: "osascript", Args: script}, true
	default:
		suggested := name
		if start != "" {
			suggested = strings.TrimRight(start, "/") + "/" + name
		}
		switch {
		case has("zenity"):
			args := []string{"--file-selection", "--save", "--confirm-overwrite", "--title=Save as",
				"--file-filter=" + label + " | *." + extension, "--file-filter=All files | *"}
			if suggested != "" {
				args = append(args, "--filename="+suggested)
			}

			return oscommand.Command{Name: "zenity", Args: args}, true
		case has("kdialog"):
			if suggested == "" {
				suggested = "."
			}

			return oscommand.Command{Name: "kdialog", Args: []string{"--getsavefilename", suggested, "*." + extension + "|" + label}}, true
		}

		return oscommand.Command{}, false
	}
}

// withExtension adds the extension (without its dot) to a name that has none.
func withExtension(path, extension string) string {
	if filepath.Ext(path) == "" && path != "" {
		return path + "." + strings.TrimPrefix(extension, ".")
	}

	return path
}
