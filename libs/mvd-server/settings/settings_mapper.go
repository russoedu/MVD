package settings

import "youtube-downloader/libs/mvd-core/config"

// From extracts the editable settings from a config.
func From(cfg config.Config) Settings {
	return Settings{
		OutputDir:                  cfg.OutputDir,
		VideoQuality:               cfg.VideoQuality,
		AudioQuality:               cfg.AudioQuality,
		MergeOutputFormat:          cfg.MergeOutputFormat,
		OutputTemplate:             cfg.OutputTemplate,
		MaxConcurrentDownloads:     cfg.MaxConcurrentDownloads,
		ConcurrentFragments:        cfg.ConcurrentFragments,
		DownloadOfficialMusicVideo: cfg.DownloadOfficialMusicVideo,
		AutoRetry:                  cfg.AutoRetry,
		Cookies:                    cookiesOf(cfg),
		CreateLogFile:              cfg.CreateLogFile,
		LogDir:                     cfg.LogDir,
	}
}

// Apply returns cfg with the editable settings replaced. Everything else in cfg (the
// raw format override, the cookie file path, extra yt-dlp arguments) is kept.
func Apply(cfg config.Config, s Settings) config.Config {
	cfg.OutputDir = s.OutputDir
	cfg.VideoQuality = s.VideoQuality
	cfg.AudioQuality = s.AudioQuality
	cfg.MergeOutputFormat = s.MergeOutputFormat
	cfg.OutputTemplate = s.OutputTemplate
	cfg.MaxConcurrentDownloads = s.MaxConcurrentDownloads
	cfg.ConcurrentFragments = s.ConcurrentFragments
	cfg.DownloadOfficialMusicVideo = s.DownloadOfficialMusicVideo
	cfg.AutoRetry = s.AutoRetry
	cfg.CreateLogFile = s.CreateLogFile
	cfg.LogDir = s.LogDir

	switch s.Cookies {
	case "all":
		cfg.AutoCookies, cfg.CookiesFromBrowser = true, ""
	case "off":
		cfg.AutoCookies, cfg.CookiesFromBrowser = false, ""
	default:
		cfg.AutoCookies, cfg.CookiesFromBrowser = false, s.Cookies
	}

	return cfg
}

func cookiesOf(cfg config.Config) string {
	switch {
	case cfg.CookiesFromBrowser != "":
		return cfg.CookiesFromBrowser
	case cfg.AutoCookies:
		return "all"
	default:
		return "off"
	}
}
