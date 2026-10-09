// Package runner assembles a ready-to-run engine from a config and a URL
// list: it resolves cookies, wires the official-video resolver and returns
// the engine. main and the TUI both use it so neither re-implements the wiring.
package runner

import (
	"context"
	"os"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/cookies"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/official"
	"youtube-downloader/libs/mvd-core/ytdlp"
)

// BuildEngine resolves cookies and builds the engine for cfg and urls. logf
// receives human-readable progress (cookie discovery, etc.); it may be nil.
func BuildEngine(ctx context.Context, ytDlpPath string, cfg config.Config, urls []string, logf func(string, ...interface{})) (*engine.Engine, error) {
	log := func(format string, a ...interface{}) {
		if logf != nil {
			logf(format, a...)
		}
	}

	// A config written for an older yt-dlp may have flags in a form it no longer accepts.
	extraArgs := ytdlp.NormalizeArgs(cfg.ExtraArgs)
	probe := ""
	if len(urls) > 0 {
		probe = urls[0]
	}

	// Cookies: a pinned browser is re-exported; auto mode fills the file once
	// (reused on later runs); an existing file is used as-is.
	cookieMissing := true
	if cfg.CookiesFile != "" {
		if _, err := os.Stat(cfg.CookiesFile); err == nil {
			cookieMissing = false
		}
	}
	switch {
	case cfg.CookiesFromBrowser != "" && cfg.CookiesFile != "":
		log("exporting cookies from %s...", cfg.CookiesFromBrowser)
		if err := ytdlp.ExportCookies(ctx, ytDlpPath, cfg.CookiesFromBrowser, cfg.CookiesFile, probe, extraArgs); err != nil {
			log("warning: %v", err)
		}
	case cfg.AutoCookies && cfg.CookiesFile != "" && cookieMissing:
		log("looking for YouTube cookies in your browsers...")
		if b, ok := cookies.Acquire(ctx, ytDlpPath, cfg.CookiesFile, probe, extraArgs, cookies.InstalledBrowsers(), log); ok {
			log("found a YouTube login in %s", b)
		} else {
			log("no usable browser cookies found; continuing without them")
		}
	}

	cookiesActive := false
	if cfg.CookiesFile != "" {
		if _, err := os.Stat(cfg.CookiesFile); err == nil {
			extraArgs = append(extraArgs, ytdlp.CookieArgs(cfg.CookiesFile)...)
			cookiesActive = true
		}
	}

	opts := engine.Options{
		YtDlp:               ytDlpPath,
		URLs:                urls,
		OutputDir:           cfg.OutputDir,
		OutputTemplate:      cfg.OutputTemplate,
		Quality:             cfg.Format(),
		MergeOutputFormat:   cfg.MergeOutputFormat,
		ConcurrentFragments: cfg.ConcurrentFragments,
		ExtraArgs:           extraArgs,
		Workers:             cfg.MaxConcurrentDownloads,
		LogPath:             cfg.LogFile(),
		AutoRetry:           cfg.AutoRetry,
	}
	// The YouTube search behind both the official video lookup and the songs of
	// playlists from other services.
	searcher := func(query string) ([]official.SearchResult, error) {
		entries, err := ytdlp.ListPlaylist(ctx, ytDlpPath, "ytsearch10:"+query, extraArgs)
		out := make([]official.SearchResult, 0, len(entries))
		for _, e := range entries {
			channel := e.Channel
			if channel == "" {
				channel = e.Uploader
			}
			out = append(out, official.SearchResult{ID: e.ID, Title: e.Title, Channel: channel})
		}
		return out, err
	}
	if cfg.DownloadOfficialMusicVideo {
		resolver := official.NewResolver(nil)
		if cookiesActive {
			if _, err := resolver.UseCookies(cfg.CookiesFile); err != nil {
				log("warning: resolver cannot use cookies: %v", err)
			}
		}
		resolver.Dumper = func(videoID string) ([]official.DumpedPage, error) {
			pages, err := ytdlp.DumpPages(ctx, ytDlpPath, "https://www.youtube.com/watch?v="+videoID, extraArgs)
			out := make([]official.DumpedPage, 0, len(pages))
			for _, p := range pages {
				out = append(out, official.DumpedPage{URL: p.URL, Body: p.Body})
			}
			return out, err
		}
		resolver.Searcher = searcher
		resolver.VideoTypes = official.NewYouTubeMusicClient().VideoType
		opts.Resolver = resolver
	}
	opts.Tracks = newPlaylistTrackSource(searcher)

	return engine.New(opts)
}
