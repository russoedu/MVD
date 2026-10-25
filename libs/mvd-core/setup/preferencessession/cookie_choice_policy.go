package preferencessession

import "youtube-downloader/libs/mvd-core/config"

// CookieOptions are the choices for the cookies setting: every browser, none, or one browser.
var CookieOptions = []string{"all", "off", "firefox", "chrome", "edge", "brave", "chromium", "opera", "vivaldi", "safari"}

// ApplyCookieChoice sets the cookie settings to the choice opt.
func ApplyCookieChoice(cfg *config.Config, opt string) {
	switch opt {
	case "all":
		cfg.AutoCookies, cfg.CookiesFromBrowser = true, ""
	case "off":
		cfg.AutoCookies, cfg.CookiesFromBrowser = false, ""
	default:
		cfg.AutoCookies, cfg.CookiesFromBrowser = false, opt
	}
}

// CurrentCookieChoice is the choice the cookie settings are at.
func CurrentCookieChoice(cfg config.Config) string {
	switch {
	case cfg.CookiesFromBrowser != "":
		return cfg.CookiesFromBrowser
	case cfg.AutoCookies:
		return "all"
	default:
		return "off"
	}
}

// cookieLabel is how the cookie settings read on the screen.
func cookieLabel(cfg config.Config) string {
	switch {
	case cfg.CookiesFromBrowser != "":
		return cfg.CookiesFromBrowser
	case cfg.AutoCookies:
		return "all (try every browser)"
	default:
		return "off"
	}
}
