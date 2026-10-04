package settings

import (
	"reflect"
	"testing"

	"youtube-downloader/libs/mvd-core/config"
)

func TestSettingsRoundTripThroughAConfig(t *testing.T) {
	base := config.Default("/app", "/downloads")
	want := Settings{
		OutputDir: "/music", VideoQuality: "720p", AudioQuality: "low", MergeOutputFormat: "mkv",
		OutputTemplate: "%(title)s.%(ext)s", MaxConcurrentDownloads: 7, ConcurrentFragments: 0,
		DownloadOfficialMusicVideo: true, AutoRetry: false, Cookies: "firefox:default",
		CreateLogFile: false, LogDir: "/logs",
	}

	got := From(Apply(base, want))

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
}

func TestApplyKeepsTheSettingsThatAreNotExposed(t *testing.T) {
	base := config.Default("/app", "/downloads")
	base.RawFormat = "bv*+ba"
	base.CookiesFile = "/secret/cookies.txt"
	base.ExtraArgs = []string{"--proxy", "http://p"}

	got := Apply(base, From(base))

	if got.RawFormat != "bv*+ba" || got.CookiesFile != "/secret/cookies.txt" || !reflect.DeepEqual(got.ExtraArgs, base.ExtraArgs) {
		t.Errorf("advanced settings were changed: %+v", got)
	}
}

func TestCookiesMapToTheThreeCookieModes(t *testing.T) {
	base := config.Default("/app", "/downloads")
	for cookies, want := range map[string]struct {
		auto bool
		pin  string
	}{
		"all":     {true, ""},
		"off":     {false, ""},
		"firefox": {false, "firefox"},
	} {
		s := From(base)
		s.Cookies = cookies
		got := Apply(base, s)
		if got.AutoCookies != want.auto || got.CookiesFromBrowser != want.pin {
			t.Errorf("%s: auto=%v pin=%q", cookies, got.AutoCookies, got.CookiesFromBrowser)
		}
	}
}
