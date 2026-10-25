package advancedsession

import (
	"fmt"
	"strconv"
	"strings"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/settings"
)

// The settings the screen lists, in order.
const (
	itemExtraArgs = iota
	itemFragments
	itemAutoRetry
	itemRetryCycles
	itemRetryWait
	itemStartTimeout
)

// Labels are what the settings are called, in the order the screen lists them.
var Labels = []string{"Extra yt-dlp args", "Parallel fragments per video", "Auto Retry", "Retry cycles", "Retry wait (s x attempt)", "Start timeout (s)"}

// Session is the "if you know what you're doing" screen: raw yt-dlp args, the fragment count and
// auto-retry.
type Session struct {
	// Cfg is the settings as they stand, including changes not yet saved.
	Cfg config.Config
	// Cursor is the setting the person is on.
	Cursor int
	// Mode is how the setting of the cursor is being changed; EditNone when it is not. Text is what
	// is typed while it is.
	Mode settings.Mode
	Text string
}

// New is the session for the settings cfg.
func New(cfg config.Config) Session { return Session{Cfg: cfg} }

// Editing says whether a setting is being typed.
func (s Session) Editing() bool { return s.Mode != settings.EditNone }

// Display is what setting i reads as.
func (s Session) Display(i int) string {
	switch i {
	case itemExtraArgs:
		if len(s.Cfg.ExtraArgs) == 0 {
			return "(none)"
		}

		return strings.Join(s.Cfg.ExtraArgs, " ")
	case itemFragments:
		return strconv.Itoa(s.Cfg.ConcurrentFragments)
	case itemAutoRetry:
		return settings.YesNo(s.Cfg.AutoRetry)
	case itemRetryCycles:
		return strconv.Itoa(s.Cfg.RetryCycles)
	case itemRetryWait:
		return retryWaitText(s.Cfg.RetryWaitSeconds)
	case itemStartTimeout:
		return strconv.Itoa(s.Cfg.StartTimeoutSeconds)
	}

	return ""
}

// retryWaitText is the wait setting with what it comes to: the waits of the first
// retries, since each is the setting times the number of the retry.
func retryWaitText(seconds int) string {
	return fmt.Sprintf("%d  (%ds, %ds, %ds, ...)", seconds, seconds, 2*seconds, 3*seconds)
}

// current is the number the setting under the cursor holds now.
func (s Session) current() int {
	switch s.Cursor {
	case itemRetryCycles:
		return s.Cfg.RetryCycles
	case itemRetryWait:
		return s.Cfg.RetryWaitSeconds
	case itemStartTimeout:
		return s.Cfg.StartTimeoutSeconds
	}

	return s.Cfg.ConcurrentFragments
}

// typedNumber is the number typed, or the one the setting holds when what is typed is not one.
func (s Session) typedNumber() int {
	n, err := strconv.Atoi(strings.TrimSpace(s.Text))
	if err != nil || n < 0 {
		return s.current()
	}

	return n
}
