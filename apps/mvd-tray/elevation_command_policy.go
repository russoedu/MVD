package main

import (
	"strconv"
	"strings"
)

// exitDeclined is what the elevation command exits with when the person did not approve
// the administrator prompt, or the elevated step could not be started.
const exitDeclined = 1223

// quotePowerShell writes s as a PowerShell single-quoted string: nothing inside one is
// interpreted, and a quote is written by doubling it. Windows paths may contain quotes,
// so a path can never end the string and become code.
func quotePowerShell(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// elevatedInstallScript is what runs with administrator rights: it puts the program at
// dest, creating its folder, and adds a shortcut for every account. It does only that,
// to the three paths given. A copy that fails stops the script with an error, which
// becomes a non-zero exit code.
func elevatedInstallScript(source, dest, link string) string {
	return `$ErrorActionPreference = 'Stop'
$source = ` + quotePowerShell(source) + `
$dest = ` + quotePowerShell(dest) + `
$link = ` + quotePowerShell(link) + `
New-Item -ItemType Directory -Force -Path (Split-Path -Parent $dest) | Out-Null
Copy-Item -LiteralPath $source -Destination $dest -Force
New-Item -ItemType Directory -Force -Path (Split-Path -Parent $link) | Out-Null
$shortcut = (New-Object -ComObject WScript.Shell).CreateShortcut($link)
$shortcut.TargetPath = $dest
$shortcut.WorkingDirectory = Split-Path -Parent $dest
$shortcut.Description = 'MVD, the music video downloader'
$shortcut.Save()
`
}

// elevationCommand returns the command that runs elevatedInstallScript with
// administrator approval. Windows PowerShell does the elevating, so the program itself
// carries no code for restarting itself with administrator rights, which antivirus
// treats as a mark of malware in a program that is not signed. The person sees the
// standard prompt and, if they decline, the command exits with exitDeclined.
func elevationCommand(source, dest, link string) dialogCommand {
	inner := encodePowerShell(elevatedInstallScript(source, dest, link))
	outer := `$ErrorActionPreference = 'Stop'
try {
    $step = Start-Process -FilePath 'powershell.exe' -Verb RunAs -WindowStyle Hidden -PassThru -Wait -ArgumentList @('-NoProfile', '-NonInteractive', '-EncodedCommand', '` + inner + `')
    exit $step.ExitCode
} catch {
    exit ` + strconv.Itoa(exitDeclined) + `
}
`

	return dialogCommand{
		name: "powershell",
		args: []string{"-NoProfile", "-NonInteractive", "-EncodedCommand", encodePowerShell(outer)},
	}
}
