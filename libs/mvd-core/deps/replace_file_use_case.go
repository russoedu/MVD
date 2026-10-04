package deps

import "os"

// replaceFile puts the file at partial in place of the one at final, without ever
// deleting a program that may be running.
//
// A running program cannot be overwritten or deleted on Windows, but it can be renamed,
// so the old copy is moved aside first. Processes already started keep running from it
// and every process started afterwards finds the new one. The old copy is deleted if
// that is possible now, and otherwise a later start removes it. If the new file cannot
// be put in place, the old one is put back.
func replaceFile(partial, final string) error {
	old := final + ".old"
	_ = os.Remove(old)

	if _, err := os.Stat(final); err == nil {
		if err := os.Rename(final, old); err != nil {
			return err
		}
	}
	if err := os.Rename(partial, final); err != nil {
		_ = os.Rename(old, final)

		return err
	}
	_ = os.Remove(old)

	return nil
}
