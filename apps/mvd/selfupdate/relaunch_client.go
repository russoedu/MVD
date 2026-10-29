package selfupdate

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// AfterExitFlag is the flag the new program is started with: it waits for the old one (its
// process number) to end before it opens, so the two never hold the single window at once.
const AfterExitFlag = "after-exit"

// WaitForExit waits until the process pid has ended, at most timeout. A process that is already
// gone, or that cannot be looked at, counts as ended.
func WaitForExit(pid int, timeout time.Duration) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	done := make(chan struct{})
	go func() {
		_, _ = p.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
	}
}

// Relaunch starts the program updated in place, to wait for this one to end and then open. It is
// an ordinary visible program in its own folder, started without a shell and without hiding
// anything. It returns the error of starting it.
func Relaunch(exe string) error {
	cmd := exec.Command(exe, "-"+AfterExitFlag+"="+strconv.Itoa(os.Getpid()))
	cmd.Dir = filepath.Dir(exe)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()

	return nil
}
