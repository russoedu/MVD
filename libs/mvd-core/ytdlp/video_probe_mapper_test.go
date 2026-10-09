package ytdlp

import "testing"

func TestVideoProbeCarriesTheBestStoryboard(t *testing.T) {
	data := []byte(`{"formats":[
	  {"format_id":"251","vcodec":"none","acodec":"opus","abr":160},
	  {"format_id":"137","vcodec":"avc1","acodec":"none","height":1080},
	  {"format_id":"sb3","width":48,"height":27,"rows":10,"columns":10,"fragments":[{"url":"https://x/sb3.jpg","duration":200}]},
	  {"format_id":"sb2","width":80,"height":45,"rows":10,"columns":10,"fragments":[{"url":"https://x/a.jpg","duration":100},{"url":"https://x/b.jpg","duration":88.5}]},
	  {"format_id":"sb1","width":160,"height":90,"rows":5,"columns":5,"fragments":[{"url":"https://x/c.jpg","duration":25}]}]}`)

	got, err := ParseVideoProbe(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Quality != (MediaQuality{Height: 1080, AudioKbps: 160}) {
		t.Errorf("quality: %+v", got.Quality)
	}
	sb := got.Storyboard
	if sb.Width != 80 || sb.Rows != 10 || sb.Columns != 10 || len(sb.Fragments) != 2 || sb.Fragments[1].DurationSec != 88.5 {
		t.Errorf("storyboard: %+v", sb)
	}
}

func TestVideoProbeWithoutAStoryboardHasAnEmptyOne(t *testing.T) {
	got, err := ParseVideoProbe([]byte(`{"formats":[{"format_id":"18","vcodec":"avc1","acodec":"aac","height":360}]}`))
	if err != nil || len(got.Storyboard.Fragments) != 0 {
		t.Errorf("got %+v, %v", got, err)
	}
}
