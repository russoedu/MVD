package official

import "testing"

func TestATitleThatSaysItIsInHighQualityIsRecognised(t *testing.T) {
	for title, want := range map[string]bool{
		"Fun Factory - Don't Go Away (HD, 1080p, 16:9)": true,
		"Song [4K]":           true,
		"Song (Good quality)": false,
		"Song - Hdr Mix":      false,
		"Song":                false,
	} {
		if got := AdvertisesQuality(title); got != want {
			t.Errorf("%q: got %v, want %v", title, got, want)
		}
	}
}
