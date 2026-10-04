package main

import (
	"strings"

	"youtube-downloader/apps/mvd-tray/oscommand"
)

// startEnv is the variable the Windows script reads the starting folder from.
const startEnv = "MVD_START"

// windowsScript shows the classic folder chooser above every other window and prints
// the chosen path as UTF-8. Printing nothing means the person cancelled.
const windowsScript = `$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
Add-Type -AssemblyName System.Windows.Forms
$owner = New-Object System.Windows.Forms.Form
$owner.TopMost = $true
$owner.ShowInTaskbar = $false
$owner.WindowState = 'Minimized'
$owner.Show()
$dialog = New-Object System.Windows.Forms.FolderBrowserDialog
$dialog.Description = 'Choose a folder'
$dialog.ShowNewFolderButton = $true
if ($env:MVD_START -and (Test-Path -LiteralPath $env:MVD_START -PathType Container)) { $dialog.SelectedPath = $env:MVD_START }
$result = $dialog.ShowDialog($owner)
$owner.Close()
if ($result -eq [System.Windows.Forms.DialogResult]::OK) { [Console]::Out.Write($dialog.SelectedPath) }
`

// folderDialogCommand picks the way to show a folder chooser on goos, using only the
// programs has reports as installed. ok is false when there is none.
func folderDialogCommand(goos, start string, has func(string) bool) (cmd oscommand.Command, ok bool) {
	switch goos {
	case "windows":
		if !has("powershell") {
			return oscommand.Command{}, false
		}

		return oscommand.Command{
			Name: "powershell",
			Args: []string{"-NoProfile", "-NonInteractive", "-STA", "-EncodedCommand", oscommand.EncodePowerShell(windowsScript)},
			Env:  []string{startEnv + "=" + start},
		}, true
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
