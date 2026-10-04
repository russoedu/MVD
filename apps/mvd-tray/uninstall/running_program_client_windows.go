package uninstall

import (
	"fmt"
	"os"
	"path/filepath"
)

// removeRunningProgram gets the running program out of the way. Windows will not delete
// a program that is running, but it does allow it to be renamed, so the program is moved
// into the temporary folder, which Windows empties by itself, and the folder it was in
// can then be removed. A temporary folder on another drive cannot take it, in which case
// it is renamed where it is, and that is reported as left behind.
//
// Nothing here starts another program or schedules anything for the next restart.
func removeRunningProgram(path string) (leftover string, err error) {
	aside := filepath.Join(os.TempDir(), fmt.Sprintf("mvd-removed-%d.exe", os.Getpid()))
	if err := os.Rename(path, aside); err == nil {
		return "", nil
	}

	inPlace := path + ".removed"
	if err := os.Rename(path, inPlace); err != nil {
		return "", err
	}

	return inPlace, nil
}
