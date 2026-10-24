package notification

import (
	"strings"

	"youtube-downloader/apps/mvd/oscommand"
)

// Windows' balloon notifications cut the title at 63 characters and the text at 255.
const (
	noticeTitleLimit = 63
	noticeTextLimit  = 255
)

// notificationCommand picks the program that shows n on goos, using only the programs has
// reports as installed. ok is false when there is none, as on Windows, which needs no program.
func notificationCommand(goos string, n Notice, has func(string) bool) (cmd oscommand.Command, ok bool) {
	switch goos {
	case "windows":
		// Windows shows its balloon through the system's own interface, not a program.
		return oscommand.Command{}, false
	case "darwin":
		if !has("osascript") {
			return oscommand.Command{}, false
		}

		// The words are argv, not part of the AppleScript text.
		return oscommand.Command{Name: "osascript", Args: []string{
			"-e", "on run argv",
			"-e", "display notification (item 1 of argv) with title (item 2 of argv)",
			"-e", "end run",
			n.Text, n.Title,
		}}, true
	default:
		if !has("notify-send") {
			return oscommand.Command{}, false
		}
		urgency := "normal"
		if n.Failure {
			urgency = "critical"
		}

		return oscommand.Command{Name: "notify-send", Args: []string{"--app-name=MVD", "--urgency=" + urgency, "--", n.Title, n.Text}}, true
	}
}

// clip shortens s to at most limit characters, ending with "..." when it had to cut.
func clip(s string, limit int) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= limit {
		return string(runes)
	}

	return string(runes[:limit-3]) + "..."
}
