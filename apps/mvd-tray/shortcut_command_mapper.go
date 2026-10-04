package main

import "path/filepath"

// shortcutScript creates a shortcut. The paths come from the environment, so a folder
// name can never be read as code.
const shortcutScript = `$ErrorActionPreference = 'Stop'
$shortcut = (New-Object -ComObject WScript.Shell).CreateShortcut($env:MVD_SHORTCUT)
$shortcut.TargetPath = $env:MVD_TARGET
$shortcut.WorkingDirectory = $env:MVD_FOLDER
$shortcut.Description = 'MVD, the music video downloader'
$shortcut.Save()
`

// shortcutCommand returns the command that makes a shortcut at link to target.
func shortcutCommand(link, target string) dialogCommand {
	return dialogCommand{
		name: "powershell",
		args: []string{"-NoProfile", "-NonInteractive", "-EncodedCommand", encodePowerShell(shortcutScript)},
		env: []string{
			"MVD_SHORTCUT=" + link,
			"MVD_TARGET=" + target,
			"MVD_FOLDER=" + filepath.Dir(target),
		},
	}
}
