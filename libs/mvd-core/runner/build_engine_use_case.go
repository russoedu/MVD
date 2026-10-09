// Package runner assembles a ready-to-run engine from a config and a URL
// list: it resolves cookies, wires the official-video resolver and returns
// the engine. main and the TUI both use it so neither re-implements the wiring.
package runner

import (
	"context"
	"os"
	"path/filepath"

	"youtube-downloader/libs/mvd-core/appdir"
	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/cookies"
	"youtube-downloader/libs/mvd-core/engine"
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
	// Videos are downloaded and merged out of the library and moved in once complete.
	if appDir, err := appdir.Dir(); err == nil {
		opts.PartialDir = filepath.Join(appDir, "partial")
	}
	// The resolver is also what songs from other services are found with, so it is
	// built whether or not the official video option is on; only the option puts it in
	// front of the entries of YouTube playlists.
	cacheFile := ""
	if appDir, err := appdir.Dir(); err == nil {
		cacheFile = filepath.Join(appDir, "official-videos.json")
	}
	resolver, searcher := BuildResolver(ctx, ResolverInput{
		YtDlp: ytDlpPath, ExtraArgs: extraArgs, CookiesFile: cfg.CookiesFile, CookiesActive: cookiesActive, CacheFile: cacheFile,
	}, log)
	if cfg.DownloadOfficialMusicVideo {
		opts.Resolver = engineResolver{resolver}
	}
	opts.Tracks = newPlaylistTrackSource(searcher, resolver.Sources)

	return engine.New(opts)
}
