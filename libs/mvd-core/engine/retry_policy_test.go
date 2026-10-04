package engine

import "testing"

func TestClassify(t *testing.T) {
	cases := map[string]RetryClass{
		"[youtube] x: Private video. Sign in if you've been granted access":       ClassPermanent,
		"[youtube] x: This video has been removed by the uploader":                ClassPermanent,
		"[youtube] x: This video is not available in your country":                ClassPermanent,
		"[youtube] x: Join this channel to get access to members-only content":    ClassPermanent,
		"[youtube] x: Sign in to confirm your age":                                ClassPermanent,
		"[youtube] x: This video is no longer available due to a copyright claim": ClassPermanent,
		"[youtube] x: HTTP Error 429: Too Many Requests":                          ClassDeferred,
		"[youtube] x: Sign in to confirm you're not a bot":                        ClassDeferred,
		"[youtube] x: Unable to download webpage: HTTP Error 403: Forbidden":      ClassDeferred,
		"[youtube] x: Unable to download webpage: The read operation timed out":   ClassDeferred,
		"[youtube] x: HTTP Error 503: Service Unavailable":                        ClassDeferred,
		"[youtube] x: Video unavailable":                                          ClassImmediate,
		"[youtube] x: Unable to download webpage":                                 ClassImmediate,
		"[youtube] x: Some brand new error nobody has seen":                       ClassImmediate,
		"": ClassImmediate,
	}
	for msg, want := range cases {
		if got := classify(msg); got != want {
			t.Errorf("classify(%q) = %d, want %d", msg, got, want)
		}
	}
}
