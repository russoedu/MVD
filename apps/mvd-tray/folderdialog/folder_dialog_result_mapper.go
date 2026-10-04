package folderdialog

import "strings"

// folderChoice reads what a chooser program printed. Empty output is a cancel. A
// trailing separator is dropped (AppleScript and zenity add one) unless the path is
// a root such as "/" or "C:\".
func folderChoice(output string) (path string, chosen bool) {
	path = strings.TrimSpace(output)
	if path == "" {
		return "", false
	}
	for len(path) > 1 && (strings.HasSuffix(path, "/") || strings.HasSuffix(path, `\`)) {
		trimmed := path[:len(path)-1]
		if strings.HasSuffix(trimmed, ":") {
			break
		}
		path = trimmed
	}

	return path, true
}

// cancelledByExit reports whether a chooser that exited with a failure was simply
// closed by the person. zenity and kdialog exit 1 on cancel and print nothing;
// osascript exits 1 and says "User canceled" (error -128).
func cancelledByExit(goos string, exitCode int, stderr string) bool {
	if exitCode != 1 {
		return false
	}
	if goos == "darwin" {
		return strings.Contains(stderr, "-128") || strings.Contains(strings.ToLower(stderr), "user canceled")
	}

	return goos != "windows" && strings.TrimSpace(stderr) == ""
}
