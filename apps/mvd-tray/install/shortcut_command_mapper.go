package install

import (
	"path/filepath"

	"youtube-downloader/apps/mvd-tray/oscommand"
)

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
func shortcutCommand(link, target string) oscommand.Command {
	return oscommand.Command{
		Name: "powershell",
		Args: []string{"-NoProfile", "-NonInteractive", "-EncodedCommand", oscommand.EncodePowerShell(shortcutScript)},
		Env: []string{
			"MVD_SHORTCUT=" + link,
			"MVD_TARGET=" + target,
			"MVD_FOLDER=" + filepath.Dir(target),
		},
	}
}
