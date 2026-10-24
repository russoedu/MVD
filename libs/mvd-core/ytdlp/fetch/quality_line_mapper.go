package fetch

import (
	"math"
	"strconv"
	"strings"
)

// QualityPrint makes yt-dlp write, as it begins to download, the height of the picture it chose
// and the bitrate of the sound, for ParseQualityLine. It goes to a file with --print-to-file, so
// the lines yt-dlp prints are not changed by it.
const QualityPrint = "before_dl:%(height)s|%(abr)s"

// ParseQualityLine reads what QualityPrint wrote: the height in pixels and the audio bitrate in
// kilobits a second, each 0 when yt-dlp had none ("NA"). found is false when it has neither.
func ParseQualityLine(text string) (height, audioKbps int, found bool) {
	first, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	parts := strings.Split(strings.TrimSpace(first), "|")
	if len(parts) != 2 {
		return 0, 0, false
	}
	number := func(s string) int {
		f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil || f <= 0 {
			return 0
		}

		return int(math.Round(f))
	}
	height, audioKbps = number(parts[0]), number(parts[1])

	return height, audioKbps, height > 0 || audioKbps > 0
}
