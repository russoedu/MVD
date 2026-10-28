package runcontrol

import "fmt"

// NothingRunning is what the person is told when they ask for something of a run and none is going.
const NothingRunning = "Nothing is running"

// PauseNotice is what is said once the downloads are held (paused) or let go on again.
func PauseNotice(paused bool) string {
	if paused {
		return "Paused: the songs on their way finish, and nothing else begins"
	}

	return "Download is on: the songs you accept start now"
}

// RetryNotice says how many failed songs were queued again.
func RetryNotice(n int) string { return fmt.Sprintf("%d songs queued again", n) }

// AcceptedNotice says how many waiting songs were accepted for download. While the downloads are
// held, they only begin once the person presses Download, and the notice says so.
func AcceptedNotice(n int, held bool) string {
	switch {
	case n == 1 && held:
		return "1 song accepted: press Download to start it"
	case held:
		return fmt.Sprintf("%d songs accepted: press Download to start them", n)
	case n == 1:
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
