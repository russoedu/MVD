package main

import (
	"strings"

	"youtube-downloader/apps/mvd-tray/oscommand"
)

// Windows' balloon notifications cut the title at 63 characters and the text at 255.
const (
	noticeTitleLimit = 63
	noticeTextLimit  = 255
)

const (
	noticeTitleEnv = "MVD_NOTICE_TITLE"
	noticeTextEnv  = "MVD_NOTICE_TEXT"
	noticeIconEnv  = "MVD_NOTICE_ICON"
)

// windowsNoticeScript shows a balloon notification from a temporary notification-area
// icon and keeps it alive long enough to be read. The words come from the environment,
// so what they contain can never be read as code.
const windowsNoticeScript = `$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Windows.Forms, System.Drawing
$icon = New-Object System.Windows.Forms.NotifyIcon
$icon.Icon = [System.Drawing.SystemIcons]::Information
$icon.BalloonTipIcon = $env:MVD_NOTICE_ICON
$icon.BalloonTipTitle = $env:MVD_NOTICE_TITLE
$icon.BalloonTipText = $env:MVD_NOTICE_TEXT
$icon.Visible = $true
$icon.ShowBalloonTip(10000)
Start-Sleep -Seconds 11
$icon.Dispose()
`

// notificationCommand picks the way to show n on goos, using only the programs has
// reports as installed. ok is false when there is none.
func notificationCommand(goos string, n userNotice, has func(string) bool) (cmd oscommand.Command, ok bool) {
	switch goos {
	case "windows":
		if !has("powershell") {
			return oscommand.Command{}, false
		}
		icon := "Info"
		if n.Failure {
			icon = "Error"
		}

		return oscommand.Command{
			Name: "powershell",
			Args: []string{"-NoProfile", "-NonInteractive", "-EncodedCommand", oscommand.EncodePowerShell(windowsNoticeScript)},
			Env: []string{
				noticeTitleEnv + "=" + clip(n.Title, noticeTitleLimit),
				noticeTextEnv + "=" + clip(n.Text, noticeTextLimit),
				noticeIconEnv + "=" + icon,
			},
		}, true
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
