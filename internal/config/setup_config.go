// Package config loads the two configuration files the app reads at start:
// setup.conf (settings) and downloads.conf (playlist URLs).
package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds the settings read from setup.conf.
type Config struct {
	OutputDir              string
	Quality                string
	MergeOutputFormat      string
	OutputTemplate         string
	MaxConcurrentDownloads int
	ExtraArgs              []string
	// DownloadOfficialMusicVideo replaces auto-generated "- Topic" art
	// tracks with the official music video linked from their description.
	DownloadOfficialMusicVideo bool
	// LogFile receives every line of output. Empty disables it.
	LogFile string
}

// Default returns the settings used when setup.conf lacks a key.
func Default() Config {
	return Config{
		OutputDir:              "./downloads",
		Quality:                "bestvideo+bestaudio/best",
		MergeOutputFormat:      "mp4",
		OutputTemplate:         "%(playlist_title,playlist)s/%(playlist_index)02d - %(title)s.%(ext)s",
		MaxConcurrentDownloads: 3,
		ExtraArgs:              []string{"-4", "--js-runtimes", "deno,node"},
		LogFile:                "mvd.log",
	}
}

// LoadSetup reads setup.conf as key=value pairs on top of Default. A
// missing file is not an error: the defaults are used and a warning printed.
func LoadSetup(path string) (Config, error) {
	cfg := Default()

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Warning: '%s' not found. Using default configurations.\n", path)
			return cfg, nil
		}
		return cfg, err
	}
	defer file.Close()

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

		switch key {
		case "output_dir":
			if val != "" {
				cfg.OutputDir = val
			}
		case "quality":
			if val != "" {
				cfg.Quality = val
			}
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
		case "extra_args":
			if val != "" {
				cfg.ExtraArgs = strings.Fields(val)
			}
		case "download_official_music_video":
			cfg.DownloadOfficialMusicVideo = parseBool(val)
		case "log_file":
			switch strings.ToLower(val) {
			case "":
			case "off", "none", "false":
				cfg.LogFile = ""
			default:
				cfg.LogFile = val
			}
		}
	}

	return cfg, scanner.Err()
}

// parseBool accepts the usual spellings of a boolean config value.
func parseBool(val string) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true", "yes", "on", "y":
		return true
	}
	return false
}
