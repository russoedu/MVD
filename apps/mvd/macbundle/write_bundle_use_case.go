package macbundle

import (
	"os"
	"path/filepath"

	"youtube-downloader/apps/mvd/programfile"
)

// Write builds the bundle at bundle around the program: the program itself, the
// Info.plist that makes macOS treat the folder as an application and, when icon is the
// path of an .icns file, that icon. An older bundle at the same place is replaced.
func Write(bundle, program, version, icon string) error {
	if err := programfile.Place(program, ProgramPath(bundle)); err != nil {
		return err
	}
	if icon != "" {
		data, err := os.ReadFile(icon)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(IconPath(bundle)), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(IconPath(bundle), data, 0o644); err != nil {
			return err
		}
	}

	return os.WriteFile(InfoPlistPath(bundle), []byte(InfoPlist(version, icon != "")), 0o644)
}
