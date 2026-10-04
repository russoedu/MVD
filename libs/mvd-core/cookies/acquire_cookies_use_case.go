package cookies

import (
	"context"
	"os"

	"youtube-downloader/libs/mvd-core/ytdlp"
)

// LogFunc receives progress messages, printf style.
type LogFunc func(format string, a ...interface{})

// Acquire exports cookies from each browser in turn (most reliable first)
// and keeps the first export that carries a live YouTube login. It returns
// the browser used and true, or "" and false when none worked, in which case
// any file it wrote is removed so the caller cleanly continues without
// cookies. probeURL is a cheap URL for yt-dlp to touch, since the jar is
// only written at the end of a run.
func Acquire(ctx context.Context, bin, file, probeURL string, extraArgs, browsers []string, log LogFunc) (string, bool) {
	logf := func(format string, a ...interface{}) {
		if log != nil {
			log(format, a...)
		}
	}
	if len(browsers) == 0 {
		logf("no browser cookie stores found")
		return "", false
	}
	for _, b := range browsers {
		if ctx.Err() != nil {
			return "", false
		}
		if err := ytdlp.ExportCookies(ctx, bin, b, file, probeURL, extraArgs); err != nil {
			logf("%s: cookies unreadable, skipping", b)
			continue
		}
		if !hasYouTubeSession(file) {
			logf("%s: no YouTube login found", b)
			continue
		}
		return b, true
	}
	_ = os.Remove(file) // an unusable cookie file is only in the way
	return "", false
}
