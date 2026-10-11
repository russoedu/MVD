package musicsource

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"youtube-downloader/libs/mvd-core/official/trackpick"
)

func TestSearchVideosReadsYouTubeMusicRows(t *testing.T) {
	answer := `{"contents":{"sectionListRenderer":{"contents":[{"musicShelfRenderer":{"contents":[
	 {"musicResponsiveListItemRenderer":{"flexColumns":[
	   {"musicResponsiveListItemFlexColumnRenderer":{"text":{"runs":[{"text":"Killer"}]}}},
	   {"musicResponsiveListItemFlexColumnRenderer":{"text":{"runs":[{"text":"ATB"},{"text":" • "},{"text":"327K views"},{"text":" • "},{"text":"4:08"}]}}}],
	  "overlay":{"navigationEndpoint":{"watchEndpoint":{"videoId":"sJ0FFbRBQlE","watchEndpointMusicSupportedConfigs":{"musicVideoType":"MUSIC_VIDEO_TYPE_OMV"}}}}}},
	 {"musicResponsiveListItemRenderer":{"flexColumns":[
	   {"musicResponsiveListItemFlexColumnRenderer":{"text":{"runs":[{"text":"ATB - Killer 2000 - HQ"}]}}},
	   {"musicResponsiveListItemFlexColumnRenderer":{"text":{"runs":[{"text":"85KasiaD85"},{"text":" • "},{"text":"1.2M views"},{"text":" • "},{"text":"5:36"}]}}}],
	  "overlay":{"navigationEndpoint":{"watchEndpoint":{"videoId":"1eoPsEBErws","watchEndpointMusicSupportedConfigs":{"musicVideoType":"MUSIC_VIDEO_TYPE_UGC"}}}}}},
	 {"musicResponsiveListItemRenderer":{"flexColumns":[{"musicResponsiveListItemFlexColumnRenderer":{"text":{"runs":[{"text":"no id here"}]}}}]}}
	]}}]}}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") != "https://music.youtube.com" {
			t.Errorf("the request must look like YouTube Music's own")
		}
		_, _ = w.Write([]byte(answer))
	}))
	defer server.Close()

	client := NewYouTubeMusicClient()
	client.SearchURL = server.URL
	got, err := client.SearchVideos("Killer ATB")
	if err != nil || len(got) != 2 {
		t.Fatalf("got %+v, %v; want the two rows that name a video", got, err)
	}
	want := trackpick.SearchResult{ID: "sJ0FFbRBQlE", Title: "Killer", Channel: "ATB", Views: 327000, Duration: 248, MusicType: "OMV"}
	if got[0] != want {
		t.Errorf("got %+v, want %+v", got[0], want)
	}
	if got[1].MusicType != "UGC" || got[1].Views != 1200000 || got[1].Duration != 336 {
		t.Errorf("got %+v", got[1])
	}
}

func TestParseCount(t *testing.T) {
	cases := map[string]int64{"327K views": 327000, "1.2M views": 1200000, "540 views": 540, "2.1B views": 2100000000, "1,234 views": 1234, "": 0, "no views": 0}
	for in, want := range cases {
		if got := parseCount(in); got != want {
			t.Errorf("parseCount(%q) = %d, want %d", in, got, want)
		}
	}
}
