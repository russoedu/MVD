package selfupdate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// AsideSuffix ends the name of the old program while the new one takes its place. Windows lets a
// running program be renamed but not deleted, so the old one waits under this name until the
// new one starts and removes it (see the call in main).
const AsideSuffix = ".old"

// replaceProgram puts the program at newProgram in the place of the one at exe, which may be
// the one that is running. Nothing changes unless the folder lets both steps happen: the new
// program is first copied next to the old one (which finds out at once that the folder is
// read-only), then the old one is renamed aside and the new one takes its name. If the second
// rename fails the old program gets its name back, so the install is never left without a program.
func replaceProgram(exe, newProgram string) error {
	staged := exe + ".new"
	if err := copyFile(newProgram, staged); err != nil {
		return fmt.Errorf("cannot write in %s: %w", filepath.Dir(exe), err)
	}
	aside := exe + AsideSuffix
	_ = os.Remove(aside) // one a previous update left behind
	if err := os.Rename(exe, aside); err != nil {
		_ = os.Remove(staged)

		return fmt.Errorf("cannot move the running program aside: %w", err)
	}
	if err := os.Rename(staged, exe); err != nil {
		restoreErr := os.Rename(aside, exe)
		_ = os.Remove(staged)
		if restoreErr != nil {
			return fmt.Errorf("cannot put the new program in place (%v), and the old one stays as %s: %w", err, aside, restoreErr)
		}

		return fmt.Errorf("cannot put the new program in place: %w", err)
	}

	return nil
}

func copyFile(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o700)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(to)
	}

	return err
}
