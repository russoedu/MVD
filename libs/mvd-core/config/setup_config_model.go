// Package config loads and saves MVD's settings. The config lives in the OS
// app-data folder (see libs/mvd-core/appdir) as a simple key=value file, created
// with defaults on first run.
package config

import (
	"path/filepath"
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

	// ListWorkers, NameWorkers and PickWorkers are how many playlists are listed, songs
	// named (artist and title looked up in music databases) and versions picked (the
	// official video, else the best quality) at the same time, ahead of the downloads.
	ListWorkers int
	NameWorkers int
	PickWorkers int

	// OfficialVideo says what to do about the official video of a song: OfficialYes
	// takes it when there is one and the song as it is otherwise, OfficialNo never
	// looks, OfficialOnly downloads only songs that have one.
	OfficialVideo OfficialMode
	// SaveNotFound, with OfficialOnly, writes the songs that have no official video to
	// a CSV file in the output folder (the playlist, the artist, the title and the
	// address of the video, when there is one).
	SaveNotFound bool
	AutoRetry    bool

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

	// Colors are the colours of the terminal interface.
	Colors ThemeColors
}

// Default returns the settings used on first run, with paths rooted at the
// app-data directory and the OS Downloads folder.
func Default(appDir, downloadsDir string) Config {
	return Config{
		OutputDir:              downloadsDir,
		VideoQuality:           "best",
		AudioQuality:           "best",
		RawFormat:              "",
		MergeOutputFormat:      "mp4",
		OutputTemplate:         "%(playlist_title,playlist)s/%(playlist_index)02d - %(title)s.%(ext)s",
		MaxConcurrentDownloads: 4,
		ConcurrentFragments:    4,
		ListWorkers:            4,
		NameWorkers:            8,
		PickWorkers:            4,
		OfficialVideo:          OfficialYes,
		SaveNotFound:           true,
		AutoRetry:              true,
		CookiesFromBrowser:     "",
		AutoCookies:            true,
		CookiesFile:            filepath.Join(appDir, "cookies.txt"),
		CreateLogFile:          true,
		LogDir:                 appDir,
		ExtraArgs:              []string{"-4", "--js-runtimes", "deno", "--js-runtimes", "node"},
		Colors:                 DefaultThemeColors(),
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

// OfficialMode is how the official video of a song is treated.
type OfficialMode string

const (
	// OfficialYes downloads the official video of a song when there is one, and the
	// song as it is when there is not.
	OfficialYes OfficialMode = "yes"
	// OfficialNo downloads every song as it is, without looking for its official video.
	OfficialNo OfficialMode = "no"
	// OfficialOnly downloads only the songs that have an official video.
	OfficialOnly OfficialMode = "only"
)

// OfficialModes are the choices, in the order the preferences show them.
var OfficialModes = []string{string(OfficialYes), string(OfficialNo), string(OfficialOnly)}

// LooksForOfficial reports whether the official video of a song is looked for at all.
func (c Config) LooksForOfficial() bool { return c.OfficialVideo != OfficialNo }

// officialModeOf reads a setting: yes, no or only, and the true and false of the older
// setting. Anything else is the default, yes.
func officialModeOf(value string) OfficialMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "no", "false", "off", "0":
		return OfficialNo
	case "only":
		return OfficialOnly
	}
	return OfficialYes
}
