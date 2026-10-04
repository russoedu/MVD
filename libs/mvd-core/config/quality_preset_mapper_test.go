package config

import "testing"

func TestCompileFormat(t *testing.T) {
	cases := []struct {
		video, audio, raw, want string
	}{
		{"best", "best", "", "bestvideo+bestaudio/best"},
		{"1080p", "best", "", "bestvideo[height<=1080]+bestaudio/best[height<=1080]"},
		{"720p", "medium", "", "bestvideo[height<=720]+bestaudio[abr<=128]/best[height<=720]"},
		{"best", "low", "", "bestvideo+bestaudio[abr<=96]/best"},
		{"nonsense", "nonsense", "", "bestvideo+bestaudio/best"}, // unknown -> best
		{"1080p", "medium", "worstvideo", "worstvideo"},          // raw override wins
	}
	for _, c := range cases {
		if got := CompileFormat(c.video, c.audio, c.raw); got != c.want {
			t.Errorf("CompileFormat(%q,%q,%q) = %q, want %q", c.video, c.audio, c.raw, got, c.want)
		}
	}
}

func TestPresetValidation(t *testing.T) {
	if !validVideoPreset("1080p") || validVideoPreset("999p") {
		t.Error("video preset validation wrong")
	}
	if !validAudioPreset("medium") || validAudioPreset("ultra") {
		t.Error("audio preset validation wrong")
	}
	if len(VideoPresets) == 0 || len(AudioPresets) == 0 {
		t.Error("preset lists should not be empty")
	}
}
