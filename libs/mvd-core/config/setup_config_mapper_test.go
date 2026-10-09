package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestParseBool(t *testing.T) {
	for _, v := range []string{"true", "True", "yes", "1", "on"} {
		if !parseBool(v) {
			t.Errorf("%q should be true", v)
		}
	}
	for _, v := range []string{"false", "0", "no", "", "maybe"} {
		if parseBool(v) {
			t.Errorf("%q should be false", v)
		}
	}
}

func TestCookieSettingRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.conf")
	base := Default(dir, dir)
	for in, want := range map[string]struct {
		auto bool
		pin  string
	}{
		"all":     {true, ""},
		"off":     {false, ""},
		"firefox": {false, "firefox"},
	} {
		c := base
		applyKey(&c, "cookies_from_browser", in, path)
		if c.AutoCookies != want.auto || c.CookiesFromBrowser != want.pin {
			t.Errorf("%q -> auto=%v pin=%q", in, c.AutoCookies, c.CookiesFromBrowser)
		}
		if got := cookieSetting(c); got != in {
			t.Errorf("round-trip %q -> %q", in, got)
		}
	}
}

func TestTheOfficialVideoDefaultsToYes(t *testing.T) {
	cfg := Default("/app", "/dl")
	if cfg.OfficialVideo != OfficialYes || !cfg.SaveNotFound || !cfg.LooksForOfficial() {
		t.Errorf("defaults: %+v", cfg)
	}
}

func TestTheOfficialVideoSettingReadsYesNoOnlyAndTheOlderTrueAndFalse(t *testing.T) {
	cases := map[string]OfficialMode{
		"official_video=yes":                     OfficialYes,
		"official_video=no":                      OfficialNo,
		"official_video=only":                    OfficialOnly,
		"official_video=ONLY":                    OfficialOnly,
		"download_official_music_video=true":     OfficialYes,
		"download_official_music_video=false":    OfficialNo,
		"official_video=something else entirely": OfficialYes,
	}
	for line, want := range cases {
		cfg := Default("/app", "/dl")
		key, value, _ := strings.Cut(line, "=")
		cfg.OfficialVideo = OfficialNo // so that a line that does nothing shows
		applyKey(&cfg, key, value, "/app/config.conf")
		if cfg.OfficialVideo != want {
			t.Errorf("%s: got %q, want %q", line, cfg.OfficialVideo, want)
		}
	}
}
