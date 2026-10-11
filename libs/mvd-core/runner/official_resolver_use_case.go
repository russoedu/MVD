package runner

import (
	"context"
	"time"
	"youtube-downloader/libs/mvd-core/official/musicsource"
	"youtube-downloader/libs/mvd-core/official/trackfind"
	"youtube-downloader/libs/mvd-core/official/trackpick"

	"youtube-downloader/libs/mvd-core/official"
	"youtube-downloader/libs/mvd-core/songid"
	"youtube-downloader/libs/mvd-core/stillpicture"
	"youtube-downloader/libs/mvd-core/ytdlp"
)

// ResolverInput is what the official video lookup is built from.
type ResolverInput struct {
	YtDlp     string
	ExtraArgs []string
	// CookiesFile is used by the resolver when CookiesActive says it exists.
	CookiesFile   string
	CookiesActive bool
	// CacheFile keeps the answers of earlier runs for 90 days; empty for no cache, which is
	// what a check that must see YouTube as it is now wants.
	CacheFile string
	// SkipQuality leaves out the look-up of the formats of the uploads of a song, which
	// only chooses among uploads when none is official: a check that only wants to know
	// whether the official video is still found has no use for it.
	SkipQuality bool
}

// BuildResolver wires the official video lookup against the real sites, the way the app
// uses it, and returns it with the YouTube search it and the songs of other services share.
func BuildResolver(ctx context.Context, in ResolverInput, log func(string, ...interface{})) (*official.Resolver, func(string) ([]trackpick.SearchResult, error)) {
	searcher := func(query string) ([]trackpick.SearchResult, error) {
		entries, err := ytdlp.ListPlaylist(ctx, in.YtDlp, "ytsearch10:"+query, in.ExtraArgs)
		out := make([]trackpick.SearchResult, 0, len(entries))
		for _, e := range entries {
			channel := e.Channel
			if channel == "" {
				channel = e.Uploader
			}
			out = append(out, trackpick.SearchResult{ID: e.ID, Title: e.Title, Channel: channel, Duration: int(e.Duration), Views: e.ViewCount, Verified: e.ChannelIsVerified})
		}
		return out, err
	}
	resolver := official.NewResolver(nil)
	if in.CookiesActive {
		if _, err := resolver.UseCookies(in.CookiesFile); err != nil {
			log("warning: resolver cannot use cookies: %v", err)
		}
	}
	resolver.Dumper = func(videoID string) ([]official.DumpedPage, error) {
		pages, err := ytdlp.DumpPages(ctx, in.YtDlp, "https://www.youtube.com/watch?v="+videoID, in.ExtraArgs)
		out := make([]official.DumpedPage, 0, len(pages))
		for _, p := range pages {
			out = append(out, official.DumpedPage{URL: p.URL, Body: p.Body})
		}
		return out, err
	}
	music := musicsource.NewYouTubeMusicClient()
	resolver.Searcher = searcher
	resolver.TrackInfos = music.Describe
	resolver.Sources = trackfind.Sources{
		Known: resolver.KnownCandidates(musicsource.NewWikidataClient()),
		Music: music.SearchVideos,
		// A music database names the song of an upload.
		Identify: songid.NewIdentifier().Identify,
		Type: func(videoID string) (string, error) {
			info, err := music.Describe(videoID)
			return info.Type, err
		},
	}
	if !in.SkipQuality {
		// What yt-dlp says of an upload (the formats and the storyboard) is asked once.
		probes := newVideoProbes(ctx, in.YtDlp, in.ExtraArgs)
		// The formats of the uploads of a song tell which has the best picture and sound.
		resolver.Sources.Quality = func(videoID string) (trackpick.Quality, error) {
			probe, err := probes.of(videoID)
			return trackpick.Quality{Height: probe.Quality.Height, AudioKbps: probe.Quality.AudioKbps}, err
		}
		// A video that does not really move (a picture with the song over it, a lyric
		// video over one background) is not a better version of the song. Its motion is
		// read from the storyboard, the small frames YouTube keeps of every video.
		resolver.Sources.Static = stillpicture.NewChecker(func(videoID string) (stillpicture.Storyboard, error) {
			probe, err := probes.of(videoID)
			sb := probe.Storyboard
			board := stillpicture.Storyboard{Width: sb.Width, Height: sb.Height, Rows: sb.Rows, Columns: sb.Columns}
			for _, fragment := range sb.Fragments {
				board.Fragments = append(board.Fragments, stillpicture.StoryboardFragment{URL: fragment.URL, DurationSec: fragment.DurationSec})
			}
			return board, err
		}).IsStatic
	}
	if in.CacheFile != "" {
		resolver.Sources.Cache = trackfind.NewResolutionCache(in.CacheFile, 90*24*time.Hour)
	}

	return resolver, searcher
}
