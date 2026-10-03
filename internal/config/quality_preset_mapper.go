package config

import "fmt"

// VideoPresets are the Video Quality choices, best first. Each caps the
// picture height; "best" leaves it uncapped.
var VideoPresets = []string{"best", "2160p", "1440p", "1080p", "720p", "480p"}

// AudioPresets are the Audio Quality choices, best first. Each caps the
// audio bitrate; "best" leaves it uncapped.
var AudioPresets = []string{"best", "high", "medium", "low"}

// videoHeight maps a video preset to its height cap (0 = uncapped).
var videoHeight = map[string]int{
	"best": 0, "2160p": 2160, "1440p": 1440, "1080p": 1080, "720p": 720, "480p": 480,
}

// audioBitrate maps an audio preset to its bitrate cap in kbps (0 = uncapped).
var audioBitrate = map[string]int{
	"best": 0, "high": 192, "medium": 128, "low": 96,
}

// CompileFormat turns the Video and Audio presets into a yt-dlp -f string.
// A non-empty raw override is returned verbatim. Unknown presets fall back to
// "best". Example: 1080p + medium ->
// "bestvideo[height<=1080]+bestaudio[abr<=128]/best[height<=1080]".
func CompileFormat(video, audio, raw string) string {
	if raw != "" {
		return raw
	}

	h, ok := videoHeight[video]
	if !ok {
		h = 0
	}
	abr, ok := audioBitrate[audio]
	if !ok {
		abr = 0
	}

	vsel := "bestvideo"
	fallback := "best"
	if h > 0 {
		vsel = fmt.Sprintf("bestvideo[height<=%d]", h)
		fallback = fmt.Sprintf("best[height<=%d]", h)
	}
	asel := "bestaudio"
	if abr > 0 {
		asel = fmt.Sprintf("bestaudio[abr<=%d]", abr)
	}

	return vsel + "+" + asel + "/" + fallback
}

// validVideoPreset reports whether v is a known video preset.
func validVideoPreset(v string) bool {
	_, ok := videoHeight[v]
	return ok
}

// validAudioPreset reports whether a is a known audio preset.
func validAudioPreset(a string) bool {
	_, ok := audioBitrate[a]
	return ok
}
