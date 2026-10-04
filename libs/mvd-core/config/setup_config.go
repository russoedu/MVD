// Package config loads and saves MVD's settings. The config lives in the OS
// app-data folder (see internal/appdir) as a simple key=value file, created
// with defaults on first run.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds every setting. Paths are absolute once resolved by Default.
type Config struct {
	OutputDir string
	// VideoQuality and AudioQuality are preset keys (see quality_preset_mapper).
	VideoQuality string
	AudioQuality string
	// RawFormat, when set, is a raw yt-dlp -f string that overrides the presets.
	RawFormat         string
	MergeOutputFormat string
	OutputTemplate    string

	MaxConcurrentDownloads int
	ConcurrentFragments    int

	DownloadOfficialMusicVideo bool
	AutoRetry                  bool

	// CookiesFromBrowser pins one yt-dlp browser spec; empty means none pinned.
	CookiesFromBrowser string
	// AutoCookies tries every installed browser. Ignored when a browser is pinned.
	AutoCookies bool
	// CookiesFile is where cookies are stored/read. Empty disables cookies.
	CookiesFile string

	CreateLogFile bool
	LogDir        string

	// ExtraArgs are advanced raw yt-dlp flags.
	ExtraArgs []string
}

// Default returns the settings used on first run, with paths rooted at the
// app-data directory and the OS Downloads folder.
func Default(appDir, downloadsDir string) Config {
	return Config{
		OutputDir:                  downloadsDir,
		VideoQuality:               "best",
		AudioQuality:               "best",
		RawFormat:                  "",
		MergeOutputFormat:          "mp4",
		OutputTemplate:             "%(playlist_title,playlist)s/%(playlist_index)02d - %(title)s.%(ext)s",
		MaxConcurrentDownloads:     4,
		ConcurrentFragments:        4,
		DownloadOfficialMusicVideo: false,
		AutoRetry:                  true,
		CookiesFromBrowser:         "",
		AutoCookies:                true,
		CookiesFile:                filepath.Join(appDir, "cookies.txt"),
		CreateLogFile:              true,
		LogDir:                     appDir,
		ExtraArgs:                  []string{"-4", "--js-runtimes", "deno,node"},
	}
}

// Format returns the yt-dlp -f string the config resolves to.
func (c Config) Format() string {
	return CompileFormat(c.VideoQuality, c.AudioQuality, c.RawFormat)
}

// LogFile returns the full log file path, or "" when logging is off.
func (c Config) LogFile() string {
	if !c.CreateLogFile || c.LogDir == "" {
		return ""
	}
	return filepath.Join(c.LogDir, "mvd-downloads.log")
}

// LoadOrCreate reads the config at path on top of the defaults. When the file
// does not exist it creates it (importing a legacy ./setup.conf if present) and
// reports created=true so the caller can open the config screen.
func LoadOrCreate(path, appDir, downloadsDir string) (cfg Config, created bool, err error) {
	base := Default(appDir, downloadsDir)

	if _, statErr := os.Stat(path); statErr == nil {
		cfg, err = parseInto(path, base)
		return cfg, false, err
	}

	cfg = base
	if legacy := "setup.conf"; fileExists(legacy) {
		if c, perr := parseInto(legacy, base); perr == nil {
			cfg = c
		}
	}
	if err = Save(cfg, path); err != nil {
		return cfg, false, err
	}
	return cfg, true, nil
}

// parseInto overlays the key=value file at path onto base.
func parseInto(path string, base Config) (Config, error) {
	cfg := base
	file, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		applyKey(&cfg, key, val, path)
	}
	return cfg, scanner.Err()
}

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
	}
}

// Save writes the config to path as a commented key=value file.
func Save(cfg Config, path string) error {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

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

	return os.WriteFile(path, []byte(b.String()), 0o644)
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

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
