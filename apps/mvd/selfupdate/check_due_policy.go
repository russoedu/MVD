package selfupdate

import "time"

// checkEvery is how long to wait between two looks at GitHub.
const checkEvery = 24 * time.Hour

// due says whether it is time to look again: never looked, or a day ago. A clock that went
// back counts as due, so a wrong date cannot stop the check for good.
func due(last, now time.Time) bool {
	return last.IsZero() || now.Before(last) || now.Sub(last) >= checkEvery
}
