package install

import "strings"

// desktopEntry is the applications-menu entry for the program on Linux. The program's
// path is quoted, as the format requires for one with spaces, so it is read as one word.
func desktopEntry(program string) string {
	quoted := program
	if strings.ContainsAny(program, " \t\"'\\$`") {
		escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "`", "\\`", "$", `\$`).Replace(program)
		quoted = `"` + escaped + `"`
	}

	return `[Desktop Entry]
Type=Application
Name=MVD
Comment=Music video downloader
Exec=` + quoted + `
Terminal=false
Categories=Network;AudioVideo;
`
}
