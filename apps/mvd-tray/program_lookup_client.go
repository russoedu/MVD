package main

import "os/exec"

func hasProgram(name string) bool {
	_, err := exec.LookPath(name)

	return err == nil
}
