package main

import (
	"encoding/base64"
	"strings"
	"unicode/utf16"
)

// dialogCommand is a program that shows a folder chooser and prints the choice.
type dialogCommand struct {
	name string
	args []string
	// env is added to the program's environment. The starting folder travels here (or
	// as a separate argument), never inside the script text, so a folder name can
	// never be read as code.
	env []string
}

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
func folderDialogCommand(goos, start string, has func(string) bool) (cmd dialogCommand, ok bool) {
	switch goos {
	case "windows":
		if !has("powershell") {
			return dialogCommand{}, false
		}

		return dialogCommand{
			name: "powershell",
			args: []string{"-NoProfile", "-NonInteractive", "-STA", "-EncodedCommand", encodePowerShell(windowsScript)},
			env:  []string{startEnv + "=" + start},
		}, true
	case "darwin":
		if !has("osascript") {
			return dialogCommand{}, false
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

		return dialogCommand{name: "osascript", args: script}, true
	default:
		switch {
		case has("zenity"):
			args := []string{"--file-selection", "--directory", "--title=Choose a folder"}
			if start != "" {
				args = append(args, "--filename="+strings.TrimRight(start, "/")+"/")
			}

			return dialogCommand{name: "zenity", args: args}, true
		case has("kdialog"):
			args := []string{"--getexistingdirectory"}
			if start != "" {
				args = append(args, start)
			}

			return dialogCommand{name: "kdialog", args: args}, true
		}

		return dialogCommand{}, false
	}
}

// encodePowerShell is the form -EncodedCommand takes: the script as UTF-16LE, in
// base64. It avoids every quoting rule between this program and PowerShell.
func encodePowerShell(script string) string {
	units := utf16.Encode([]rune(script))
	raw := make([]byte, 0, len(units)*2)
	for _, u := range units {
		raw = append(raw, byte(u), byte(u>>8))
	}

	return base64.StdEncoding.EncodeToString(raw)
}
