package ytdlp

import (
	"strconv"
	"strings"
)

// ProgressTemplate makes yt-dlp print one machine readable line per tick.
const ProgressTemplate = "download:MVD|%(progress.downloaded_bytes)s|%(progress.total_bytes)s|%(progress.total_bytes_estimate)s|%(progress.speed)s|%(progress.eta)s|%(progress.fragment_index)s|%(progress.fragment_count)s"

// Progress is one decoded progress tick.
type Progress struct {
	Percent    float64
	Downloaded int64
	Total      int64
	Speed      float64 // bytes per second, 0 when unknown
	ETA        int     // seconds, -1 when unknown
	// FragmentIndex and FragmentCount say how far a part that comes in fragments is
	// (DASH streams do), and are 0 for one that does not.
	FragmentIndex, FragmentCount int
}

// ParseProgressLine decodes a line produced by ProgressTemplate.
func ParseProgressLine(line string) (Progress, bool) {
	if !strings.HasPrefix(line, "MVD|") {
		return Progress{}, false
	}
	parts := strings.Split(line, "|")
	// Six fields, or eight when the fragments are given as well.
	if len(parts) != 6 && len(parts) != 8 {
		return Progress{}, false
	}
	num := func(s string) float64 {
		s = strings.TrimSpace(s)
		if s == "" || s == "NA" || s == "None" {
			return -1
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return -1
		}
		return f
	}
	downloaded := num(parts[1])
	total := num(parts[2])
	if total <= 0 {
		total = num(parts[3])
	}
	speed := num(parts[4])
	eta := num(parts[5])

	p := Progress{ETA: -1}
	if downloaded >= 0 {
		p.Downloaded = int64(downloaded)
	}
	if total > 0 {
		p.Total = int64(total)
		p.Percent = float64(p.Downloaded) / total * 100
		if p.Percent > 100 {
			p.Percent = 100
		}
	}
	if speed > 0 {
		p.Speed = speed
	}
	if eta >= 0 {
		p.ETA = int(eta)
	}
	if len(parts) == 8 {
		if index := num(parts[6]); index > 0 {
			p.FragmentIndex = int(index)
		}
		if count := num(parts[7]); count > 0 {
			p.FragmentCount = int(count)
		}
	}
	return p, true
}

// IsPostProcessLine reports whether a yt-dlp line marks the start of
// post-processing (merging, conversion, metadata).
func IsPostProcessLine(line string) bool {
	for _, p := range []string{"[Merger]", "[ExtractAudio]", "[VideoConvertor]", "[VideoRemuxer]", "[Metadata]", "[EmbedThumbnail]"} {
		if strings.HasPrefix(line, p) {
			return true
		}
	}
	return false
}
