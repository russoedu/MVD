package install

import (
	"errors"
	"io/fs"
	"os"
	"time"
)

// removeAfterExit deletes the program a move left behind, once the copy that started the
// move has exited. Windows will not delete a running program, so the first attempts
// fail; they are repeated until one works or the attempts run out. A program that is
// already gone counts as done, and one that cannot be removed is left where it is.
func removeAfterExit(path string, attempts int, interval time.Duration, remove func(string) error, sleep func(time.Duration)) bool {
	for range attempts {
		err := remove(path)
		if err == nil || errors.Is(err, fs.ErrNotExist) {
			return true
		}
		sleep(interval)
	}

	return false
}

// removeMovedProgram is removeAfterExit with the real file system and about ten seconds
// of patience.
func removeMovedProgram(path string) bool {
	return removeAfterExit(path, 50, 200*time.Millisecond, os.Remove, time.Sleep)
}
