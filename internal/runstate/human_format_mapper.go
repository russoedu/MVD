package runstate

import (
	"fmt"
	"time"
)

// HumanBytes renders a byte count the way yt-dlp does (KiB, MiB, GiB).
func HumanBytes(b int64) string {
	const unit = 1024.0
	f := float64(b)
	for _, suffix := range []string{"B", "KiB", "MiB", "GiB"} {
		if f < unit || suffix == "GiB" {
			if suffix == "B" {
				return fmt.Sprintf("%d%s", b, suffix)
			}
			return fmt.Sprintf("%.1f%s", f, suffix)
		}
		f /= unit
	}
	return fmt.Sprintf("%.1fGiB", f)
}

// HumanSpeed renders bytes per second, "--" when unknown.
func HumanSpeed(bps float64) string {
	if bps <= 0 {
		return "--"
	}
	return HumanBytes(int64(bps)) + "/s"
}

// HumanETA renders seconds as m:ss or h:mm:ss, "--" when unknown.
func HumanETA(sec int) string {
	if sec < 0 {
		return "--"
	}
	if sec >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", sec/3600, (sec%3600)/60, sec%60)
	}
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}

// HumanDuration renders a duration as hh:mm:ss.
func HumanDuration(d time.Duration) string {
	sec := int(d.Seconds())
	return fmt.Sprintf("%02d:%02d:%02d", sec/3600, (sec%3600)/60, sec%60)
}

// ProgressLine renders "34.2% of 112.4MiB at 8.1MiB/s ETA 0:09".
func ProgressLine(en *Entry) string {
	if en.Total > 0 {
		return fmt.Sprintf("%5.1f%% of %s at %s ETA %s", en.Percent, HumanBytes(en.Total), HumanSpeed(en.Speed), HumanETA(en.ETA))
	}
	return fmt.Sprintf("%s downloaded at %s", HumanBytes(en.Downloaded), HumanSpeed(en.Speed))
}
