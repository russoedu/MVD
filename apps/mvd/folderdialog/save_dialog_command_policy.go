package folderdialog

import (
	"path/filepath"
	"strings"

	"youtube-downloader/apps/mvd/oscommand"
)

// nameEnv is the variable the Windows script reads the suggested file name from.
const nameEnv = "MVD_NAME"

// windowsSaveScript shows the system's save-file window (the one of File Explorer) above
// every other window and prints the chosen path as UTF-8. Printing nothing means the
// person cancelled. The window asks before replacing a file that exists.
const windowsSaveScript = `$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
Add-Type -AssemblyName System.Windows.Forms
$owner = New-Object System.Windows.Forms.Form
$owner.TopMost = $true
$owner.ShowInTaskbar = $false
$owner.WindowState = 'Minimized'
$owner.Show()
$dialog = New-Object System.Windows.Forms.SaveFileDialog
$dialog.Title = 'Save as'
$dialog.OverwritePrompt = $true
$dialog.Filter = $env:MVD_FILTER
$dialog.DefaultExt = ($env:MVD_FILTER -split '\|')[1].TrimStart('*', '.')
$dialog.AddExtension = $true
$start = $env:MVD_START
if ($start -and (Test-Path -LiteralPath $start -PathType Container)) { $dialog.InitialDirectory = $start }
if ($env:MVD_NAME) { $dialog.FileName = $env:MVD_NAME }
$path = $null
if ($dialog.ShowDialog($owner) -eq [System.Windows.Forms.DialogResult]::OK) { $path = $dialog.FileName }
$owner.Close()
if ($path) { [Console]::Out.Write($path) }
`

// saveDialogCommand picks the way to show a save-file window on goos, using only the
// programs has reports as installed. name is the file name offered (with its extension),
// start the folder it opens in; extension (without its dot) and label describe the file
// type. ok is false when there is none.
func saveDialogCommand(goos, start, name, extension, label string, has func(string) bool) (cmd oscommand.Command, ok bool) {
	switch goos {
	case "windows":
		if !has("powershell") {
			return oscommand.Command{}, false
		}
		filter := label + " (*." + extension + ")|*." + extension + "|All files (*.*)|*.*"

		return oscommand.Command{
			Name: "powershell",
			Args: []string{"-NoProfile", "-NonInteractive", "-STA", "-EncodedCommand", oscommand.EncodePowerShell(windowsSaveScript)},
			Env:  []string{startEnv + "=" + start, filterEnv + "=" + filter, nameEnv + "=" + name},
		}, true
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
