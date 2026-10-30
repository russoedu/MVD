package cookies

import (
	"context"
	"os"

	"youtube-downloader/libs/mvd-core/ytdlp"
)

// LogFunc receives progress messages, printf style.
type LogFunc func(format string, a ...interface{})

// premiumProbeURL is a stable official upload whose format list shows whether a login is
// Premium (it then offers the 256 kbps audio).
const premiumProbeURL = "https://music.youtube.com/watch?v=BLZWkjBXfN8"

// Acquire exports cookies from each browser in turn (most reliable first) and keeps the
// export of the first browser whose login is Premium. When none is Premium it keeps the
// first export that carries a YouTube login. It returns the browser used and true, or ""
// and false when none worked, in which case any file it wrote is removed so the caller
// cleanly continues without cookies. probeURL is a cheap URL for yt-dlp to touch, since the
// jar is only written at the end of a run.
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
	// Each export goes to its own file, so a later browser cannot overwrite the first login.
	fallback, fallbackFile := "", ""
	for _, b := range browsers {
		if ctx.Err() != nil {
			break
		}
		exported := file + "." + b
		if err := ytdlp.ExportCookies(ctx, bin, b, exported, probeURL, extraArgs); err != nil {
			logf("%s: cookies unreadable, skipping", b)
			_ = os.Remove(exported)
			continue
		}
		if !hasYouTubeSession(exported) {
			logf("%s: no YouTube login found", b)
			_ = os.Remove(exported)
			continue
		}
		if ytdlp.HasPremiumAudio(ctx, bin, exported, premiumProbeURL, extraArgs) {
			if fallbackFile != "" {
				_ = os.Remove(fallbackFile)
			}
			return finish(file, b, exported)
		}
		logf("%s: logged in, but not Premium", b)
		if fallback == "" {
			fallback, fallbackFile = b, exported
		} else {
			_ = os.Remove(exported)
		}
	}
	if fallback != "" {
		return finish(file, fallback, fallbackFile)
	}
	_ = os.Remove(file) // an unusable cookie file is only in the way
	return "", false
}

// finish moves the chosen export to the cookie file.
func finish(file, browser, exported string) (string, bool) {
	if err := os.Rename(exported, file); err != nil {
		_ = os.Remove(exported)
		return "", false
	}
	return browser, true
}
