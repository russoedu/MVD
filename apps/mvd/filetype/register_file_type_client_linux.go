package filetype

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

// mimeFile is where the description of the kind of file goes, in the person's own mime
// folder (~/.local/share/mime/packages/mvd.xml).
func mimeFile(dataDir string) string {
	return filepath.Join(dataDir, "mime", "packages", "mvd.xml")
}

// userDataDir is ~/.local/share, or where XDG_DATA_HOME says.
func userDataDir() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share"), nil
}

// Register tells the system that .mvd files are download lists. The menu entry of the app
// (see apps/mvd/install) says that the app opens them, so the program is not needed here.
func Register(string) error {
	dir, err := userDataDir()
	if err != nil {
		return err
	}
	file := mimeFile(dir)
	if have, err := os.ReadFile(file); err == nil && string(have) == MimeXML() {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(file, []byte(MimeXML()), 0o644); err != nil {
		return err
	}
	refresh(dir)
	return nil
}

// Unregister takes the kind of file away again.
func Unregister() error {
	dir, err := userDataDir()
	if err != nil {
		return err
	}
	if err := os.Remove(mimeFile(dir)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	refresh(dir)
	return nil
}

// refresh has the system read the mime folder again; without the tool the files are read at
// the next sign in.
func refresh(dataDir string) {
	if tool, err := exec.LookPath("update-mime-database"); err == nil {
		_ = exec.Command(tool, filepath.Join(dataDir, "mime")).Run()
	}
}
