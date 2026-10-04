package oscommand

import "os/exec"

// HasProgram reports whether name is a program that can be run.
func HasProgram(name string) bool {
	_, err := exec.LookPath(name)

	return err == nil
}
