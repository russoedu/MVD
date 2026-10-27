package runcontrol

import "fmt"

// NothingRunning is what the person is told when they ask for something of a run and none is going.
const NothingRunning = "Nothing is running"

// PauseNotice is what is said once the downloads are held (paused) or let go on again.
func PauseNotice(paused bool) string {
	if paused {
		return "Stopped: the songs on their way finish, and nothing else begins"
	}

	return "Downloading: what you accept goes on from here"
}

// RetryNotice says how many failed songs were queued again.
func RetryNotice(n int) string { return fmt.Sprintf("%d songs queued again", n) }

// AcceptedNotice says how many waiting songs were accepted for download.
func AcceptedNotice(n int) string {
	if n == 1 {
		return "1 song accepted: it will download"
	}

	return fmt.Sprintf("%d songs accepted: they will download", n)
}

// AddedNotice says how many sources were given to the run while it works.
func AddedNotice(n int) string {
	if n == 1 {
		return "Added: its songs are loading"
	}

	return fmt.Sprintf("Added %d sources: their songs are loading", n)
}
