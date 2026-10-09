package install

import (
	"path/filepath"
	"strings"
)

// linuxIconPath is where the icon goes for the menu entry in applicationsDir, in the
// icon theme folders next to it (~/.local/share/icons/hicolor/256x256/apps/mvd.png).
func linuxIconPath(applicationsDir string) string {
	return filepath.Join(filepath.Dir(applicationsDir), "icons", "hicolor", "256x256", "apps", "mvd.png")
}

// desktopEntry is the applications-menu entry for the program on Linux, without an icon.
func desktopEntry(program string) string { return desktopEntryWithIcon(program, "") }

// desktopEntryWithIcon is the applications-menu entry for the program on Linux, showing
// the picture at icon when it is not empty. The program's path is quoted, as the format
// requires for one with spaces, so it is read as one word.
func desktopEntryWithIcon(program, icon string) string {
	quoted := program
	if strings.ContainsAny(program, " \t\"'\\$`") {
		escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "`", "\\`", "$", `\$`).Replace(program)
		quoted = `"` + escaped + `"`
	}

	iconLine := ""
	if icon != "" {
		iconLine = "Icon=" + icon + "\n"
	}

	return `[Desktop Entry]
Type=Application
Name=MVD
Comment=Music video downloader
Exec=` + quoted + `
` + iconLine + `Terminal=false
Categories=Network;AudioVideo;
`
}
