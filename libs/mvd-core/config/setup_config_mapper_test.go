package config

import (
	"path/filepath"
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
