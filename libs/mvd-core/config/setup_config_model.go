// Package config loads and saves MVD's settings. The config lives in the OS
// app-data folder (see libs/mvd-core/appdir) as a simple key=value file, created
// with defaults on first run.
package config

import "path/filepath"

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
