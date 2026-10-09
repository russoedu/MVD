package songid

import (
	"errors"
	"sync"
	"time"
)

// leftAloneFor is how long a service that refused a request is left alone: its
// callers move on to the next database.
const leftAloneFor = time.Minute

// errLeftAlone is what a call to a service that is being left alone answers.
var errLeftAlone = errors.New("the service asked to be left alone for a while")

// pacer keeps the requests to one service a minimum time apart, however many workers
// ask at once, and lets a service that said "too many" be left alone for a while.
type pacer struct {
	mu           sync.Mutex
	next         time.Time
	blockedUntil time.Time
}

// wait reserves the next free slot, at least interval after the one before it, and
// sleeps until it comes.
func (p *pacer) wait(interval time.Duration) {
	if interval <= 0 {
		return
	}
	p.mu.Lock()
	slot := time.Now()
	if p.next.After(slot) {
		slot = p.next
	}
	p.next = slot.Add(interval)
	p.mu.Unlock()

	if gap := time.Until(slot); gap > 0 {
		time.Sleep(gap)
	}
}

// blocked says whether the service is being left alone.
func (p *pacer) blocked() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return time.Now().Before(p.blockedUntil)
}

// block leaves the service alone for d, after it refused a request.
func (p *pacer) block(d time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if until := time.Now().Add(d); until.After(p.blockedUntil) {
		p.blockedUntil = until
	}
}

// refusedBy says whether a status is a service asking to be left alone.
func refusedBy(status int) bool {
	return status == 403 || status == 429 || status == 503
}
