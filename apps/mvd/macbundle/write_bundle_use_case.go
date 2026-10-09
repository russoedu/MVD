package macbundle

import (
	_ "embed"
	"os"
	"path/filepath"

	"youtube-downloader/apps/mvd/programfile"
)

// appIcon is the app's icon in the .icns form macOS reads (tools/make-icons.py makes it).
//
//go:embed mvd.icns
var appIcon []byte

// DefaultIcon is the app's own icon, as the bytes of an .icns file.
func DefaultIcon() []byte { return appIcon }

// Write builds the bundle at bundle around the program: the program itself, the
// Info.plist that makes macOS treat the folder as an application and, when icon is the
// path of an .icns file, that icon. An older bundle at the same place is replaced.
func Write(bundle, program, version, icon string) error {
	var data []byte
	if icon != "" {
		var err error
		if data, err = os.ReadFile(icon); err != nil {
			return err
		}
	}

	return WriteWithIcon(bundle, program, version, data)
}

// WriteWithIcon is Write with the icon given as the bytes of an .icns file, or none.
func WriteWithIcon(bundle, program, version string, icon []byte) error {
	if err := programfile.Place(program, ProgramPath(bundle)); err != nil {
		return err
	}
	if len(icon) > 0 {
		if err := os.MkdirAll(filepath.Dir(IconPath(bundle)), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(IconPath(bundle), icon, 0o644); err != nil {
			return err
		}
	}

	return os.WriteFile(InfoPlistPath(bundle), []byte(InfoPlist(version, len(icon) > 0)), 0o644)
}
