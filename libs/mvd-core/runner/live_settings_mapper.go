package runner

import (
	"os"
	"time"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/ytdlp"
	"youtube-downloader/libs/mvd-core/ytdlp/fetch"
)

// LiveSettings is what a run that is going on is told when the settings change: the
// ones the run reads again at every download (the others, such as where the log goes or
// how many playlists are listed at once, are taken from the settings when a run starts).
func LiveSettings(cfg config.Config) engine.LiveSettings {
	extra := fetch.NormalizeArgs(cfg.ExtraArgs)
	if cfg.CookiesFile != "" {
		if _, err := os.Stat(cfg.CookiesFile); err == nil {
			extra = append(extra, ytdlp.CookieArgs(cfg.CookiesFile)...)
		}
	}
	return engine.LiveSettings{
		Workers:             cfg.MaxConcurrentDownloads,
		Quality:             cfg.Format(),
		MergeOutputFormat:   cfg.MergeOutputFormat,
		OutputDir:           cfg.OutputDir,
		OutputTemplate:      cfg.OutputTemplate,
		ConcurrentFragments: cfg.ConcurrentFragments,
		ExtraArgs:           extra,
		AutoRetry:           cfg.AutoRetry,
		RetryCycles:         cfg.RetryCycles,
		RetryWait:           time.Duration(cfg.RetryWaitSeconds) * time.Second,
		StartTimeout:        time.Duration(cfg.StartTimeoutSeconds) * time.Second,
		LookForOfficial:     cfg.LooksForOfficial(),
		OfficialOnly:        cfg.OfficialVideo == config.OfficialOnly,
	}
}
