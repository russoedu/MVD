package config

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

func applyKey(cfg *Config, key, val, path string) {
	switch key {
	case "output_dir":
		if val != "" {
			cfg.OutputDir = val
		}
	case "video_quality":
		if validVideoPreset(val) {
			cfg.VideoQuality = val
		}
	case "audio_quality":
		if validAudioPreset(val) {
			cfg.AudioQuality = val
		}
	case "raw_format":
		cfg.RawFormat = val
	case "quality": // legacy single format -> raw override
		cfg.RawFormat = val
	case "merge_output_format":
		if val != "" {
			cfg.MergeOutputFormat = val
		}
	case "output_template":
		if val != "" {
			cfg.OutputTemplate = val
		}
	case "max_concurrent_downloads":
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			cfg.MaxConcurrentDownloads = n
		}
	case "concurrent_fragments":
		switch strings.ToLower(val) {
		case "off", "none", "false":
			cfg.ConcurrentFragments = 0
		default:
			if n, err := strconv.Atoi(val); err == nil && n >= 0 {
				cfg.ConcurrentFragments = n
			}
		}
	case "download_official_music_video":
		cfg.DownloadOfficialMusicVideo = parseBool(val)
	case "auto_retry":
		cfg.AutoRetry = parseBool(val)
	case "cookies_from_browser":
		switch strings.ToLower(val) {
		case "", "all", "auto":
			cfg.AutoCookies = true
			cfg.CookiesFromBrowser = ""
		case "off", "none", "false":
			cfg.AutoCookies = false
			cfg.CookiesFromBrowser = ""
		default:
			cfg.AutoCookies = false
			cfg.CookiesFromBrowser = val
		}
	case "cookies_file":
		switch strings.ToLower(val) {
		case "":
		case "off", "none", "false":
			cfg.CookiesFile = ""
		default:
			cfg.CookiesFile = val
		}
	case "create_log_file":
		cfg.CreateLogFile = parseBool(val)
	case "log_dir":
		if val != "" {
			cfg.LogDir = val
		}
	case "log_file": // legacy: a path or "off"
		switch strings.ToLower(val) {
		case "", "off", "none", "false":
			cfg.CreateLogFile = false
		default:
			cfg.CreateLogFile = true
			if d := filepath.Dir(val); d != "" && d != "." {
				cfg.LogDir = d
			}
		}
	case "extra_args":
		if val != "" {
			cfg.ExtraArgs = strings.Fields(val)
		}
	default:
		if slot := colourSlot(cfg, key); slot != nil && ValidColour(val) {
			*slot = strings.TrimSpace(val)
		}
	}
}

// colourSlot returns the colour setting a color_* key names, or nil.
func colourSlot(cfg *Config, key string) *string {
	switch key {
	case "color_accent":
		return &cfg.Colors.Accent
	case "color_focus":
		return &cfg.Colors.Focus
	case "color_highlight":
		return &cfg.Colors.Highlight
	case "color_success":
		return &cfg.Colors.Success
	case "color_error":
		return &cfg.Colors.Error
	case "color_dim":
		return &cfg.Colors.Dim
	case "color_text":
		return &cfg.Colors.Text
	case "color_selected":
		return &cfg.Colors.Selected
	}
	return nil
}

// configText is the commented key=value file Save writes for cfg.
func configText(cfg Config) string {
	var b strings.Builder
	b.WriteString("# MVD configuration. Edit in the app (preferences screen) or here.\n\n")
	w := func(k, v string) { fmt.Fprintf(&b, "%s=%s\n", k, v) }

	w("output_dir", cfg.OutputDir)
	w("video_quality", cfg.VideoQuality)
	w("audio_quality", cfg.AudioQuality)
	w("raw_format", cfg.RawFormat)
	w("merge_output_format", cfg.MergeOutputFormat)
	w("output_template", cfg.OutputTemplate)
	w("max_concurrent_downloads", strconv.Itoa(cfg.MaxConcurrentDownloads))
	w("concurrent_fragments", strconv.Itoa(cfg.ConcurrentFragments))
	w("download_official_music_video", strconv.FormatBool(cfg.DownloadOfficialMusicVideo))
	w("auto_retry", strconv.FormatBool(cfg.AutoRetry))
	w("cookies_from_browser", cookieSetting(cfg))
	w("cookies_file", cfg.CookiesFile)
	w("create_log_file", strconv.FormatBool(cfg.CreateLogFile))
	w("log_dir", cfg.LogDir)
	w("extra_args", strings.Join(cfg.ExtraArgs, " "))
	w("color_accent", cfg.Colors.Accent)
	w("color_focus", cfg.Colors.Focus)
	w("color_highlight", cfg.Colors.Highlight)
	w("color_success", cfg.Colors.Success)
	w("color_error", cfg.Colors.Error)
	w("color_dim", cfg.Colors.Dim)
	w("color_text", cfg.Colors.Text)
	w("color_selected", cfg.Colors.Selected)

	return b.String()
}

// cookieSetting serialises the cookie source for Save.
func cookieSetting(cfg Config) string {
	switch {
	case cfg.CookiesFromBrowser != "":
		return cfg.CookiesFromBrowser
	case cfg.AutoCookies:
		return "all"
	default:
		return "off"
	}
}

// parseBool accepts the usual spellings of a boolean config value.
func parseBool(val string) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true", "yes", "on", "y":
		return true
	}
	return false
}
