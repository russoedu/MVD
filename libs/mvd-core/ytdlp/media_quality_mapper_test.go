package ytdlp

import "testing"

func TestMediaQualityIsTheBestPictureAndTheBestAudio(t *testing.T) {
	data := []byte(`{"formats":[
	  {"vcodec":"none","acodec":"opus","abr":160.4,"tbr":160.4,"height":null},
	  {"vcodec":"none","acodec":"mp4a.40.2","abr":129.5,"height":null},
	  {"vcodec":"avc1","acodec":"none","height":720,"tbr":1500},
	  {"vcodec":"vp09","acodec":"none","height":1080,"tbr":2500},
	  {"vcodec":"avc1","acodec":"mp4a.40.2","height":360,"abr":96,"tbr":600}]}`)

	got, err := ParseMediaQuality(data)
	if err != nil || got != (MediaQuality{Height: 1080, AudioKbps: 160}) {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestMediaQualityOfAnAudioOnlyFormatWithoutABitrateUsesTheTotalRate(t *testing.T) {
	got, err := ParseMediaQuality([]byte(`{"formats":[{"vcodec":"none","acodec":"opus","tbr":128}]}`))
	if err != nil || got != (MediaQuality{AudioKbps: 128}) {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestMediaQualityOfAMuxedFormatWithoutABitrateIgnoresTheTotalRate(t *testing.T) {
	got, err := ParseMediaQuality([]byte(`{"formats":[{"vcodec":"avc1","acodec":"aac","height":480,"tbr":900}]}`))
	if err != nil || got != (MediaQuality{Height: 480}) {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestMediaQualityOfGarbageIsAnError(t *testing.T) {
	if _, err := ParseMediaQuality([]byte("nope")); err == nil {
		t.Error("expected an error")
	}
}
