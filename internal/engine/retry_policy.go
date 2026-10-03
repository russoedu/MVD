package engine

import "strings"

// RetryClass says how a failed download should be retried.
type RetryClass int

const (
	// ClassImmediate is a one-off glitch worth retrying once, right away.
	// It is also the default for unrecognised errors.
	ClassImmediate RetryClass = iota
	// ClassDeferred needs a cooldown (rate limit, bot check), so it is
	// retried in a sweep after the whole backlog drains.
	ClassDeferred
	// ClassPermanent will not change on retry (private, removed, geo-blocked).
	ClassPermanent
)

// permanentSignals mark failures that a retry cannot fix.
var permanentSignals = []string{
	"private video",
	"removed by",
	"has been removed",
	"no longer available",
	"account associated with this video has been terminated",
	"terminated",
	"members-only",
	"members only",
	"join this channel",
	"not available in your country",
	"not available in your location",
	"blocked it in your country",
	"inappropriate for some users", // age gate
	"confirm your age",
	"age-restricted",
	"violat", // violating / terms violation
	"copyright",
}

// deferredSignals mark transient failures that need time before a retry:
// rate limiting and bot checks reset after the backlog drains.
var deferredSignals = []string{
	"429",
	"too many requests",
	"http error 403",
	"sign in to confirm you", // "...you're not a bot"
	"not a bot",
	"temporarily",
	"timed out",
	"timeout",
	"connection reset",
	"connection refused",
	"connection error",
	"network",
	"http error 5", // 5xx
	"read error",
}

// classify buckets a yt-dlp error message. Permanent is checked before
// deferred so an age gate ("sign in to confirm your age") is not mistaken
// for a bot check ("sign in to confirm you're not a bot").
func classify(errMsg string) RetryClass {
	s := strings.ToLower(errMsg)
	for _, sig := range permanentSignals {
		if strings.Contains(s, sig) {
			return ClassPermanent
		}
	}
	for _, sig := range deferredSignals {
		if strings.Contains(s, sig) {
			return ClassDeferred
		}
	}
	return ClassImmediate
}
