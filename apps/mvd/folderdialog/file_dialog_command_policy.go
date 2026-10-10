package folderdialog

import (
	"strings"

	"youtube-downloader/apps/mvd/oscommand"
)

// filterEnv is the variable the Windows script reads the file filter from.
const filterEnv = "MVD_FILTER"

// windowsFileScript shows the system's open-file window (the one of File Explorer) above
// every other window and prints the chosen file as UTF-8. Printing nothing means the
// person cancelled.
const windowsFileScript = `$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
Add-Type -AssemblyName System.Windows.Forms
$owner = New-Object System.Windows.Forms.Form
$owner.TopMost = $true
$owner.ShowInTaskbar = $false
$owner.WindowState = 'Minimized'
$owner.Show()
$dialog = New-Object System.Windows.Forms.OpenFileDialog
$dialog.Title = 'Open a file'
$dialog.CheckFileExists = $true
$dialog.Filter = $env:MVD_FILTER
$start = $env:MVD_START
if ($start -and (Test-Path -LiteralPath $start -PathType Container)) { $dialog.InitialDirectory = $start }
$path = $null
if ($dialog.ShowDialog($owner) -eq [System.Windows.Forms.DialogResult]::OK) { $path = $dialog.FileName }
$owner.Close()
if ($path) { [Console]::Out.Write($path) }
`

// fileDialogCommand picks the way to show an open-file window on goos, using only the
// programs has reports as installed. extension (without its dot) and label describe the
// files to show first; every file can be chosen too. ok is false when there is none.
func fileDialogCommand(goos, start, extension, label string, has func(string) bool) (cmd oscommand.Command, ok bool) {
	switch goos {
	case "windows":
		if !has("powershell") {
			return oscommand.Command{}, false
		}
		filter := label + " (*." + extension + ")|*." + extension + "|All files (*.*)|*.*"

		return oscommand.Command{
			Name: "powershell",
			Args: []string{"-NoProfile", "-NonInteractive", "-STA", "-EncodedCommand", oscommand.EncodePowerShell(windowsFileScript)},
			Env:  []string{startEnv + "=" + start, filterEnv + "=" + filter},
		}, true
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
