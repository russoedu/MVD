package official

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// LogFunc receives progress messages, printf style.
type LogFunc func(format string, a ...interface{})

// DumpedPage is one response fetched by an external tool for a video.
type DumpedPage struct {
	URL  string
	Body []byte
}

// PageDumper fetches the pages behind a video through another tool (the
// app wires yt-dlp here), as a last resort when direct requests come back
// stripped down.
type PageDumper func(videoID string) ([]DumpedPage, error)

// Resolver crawls YouTube to find the official music video of an art track.
type Resolver struct {
	Client    *http.Client
	WatchBase string // e.g. https://www.youtube.com/watch?v=
	NextURL   string // innertube "next" endpoint
	OEmbedURL string // e.g. https://www.youtube.com/oembed?format=json&url=
	UserAgent string
	Attempts  int
	Log       LogFunc
	Dumper    PageDumper // optional
	Searcher  Searcher   // optional
	// TrackInfos, when set, tells art tracks from real videos whatever the channel is
	// called (YouTube Music's own tag) and names the song's real artist; without it
	// every upload is looked up and the artist is the channel's name.
	TrackInfos TrackDescriber
	// OfficialOnly leaves out the choice of the best quality upload when there is no
	// official video: the caller only wants official videos.
	OfficialOnly bool
	// Sources are the other places the video of a song is looked for (Wikidata, YouTube
	// Music's video search) and the cache of earlier runs; each is optional.
	Sources Sources
}

// NewResolver returns a resolver pointed at the real YouTube endpoints.
func NewResolver(log LogFunc) *Resolver {
	return &Resolver{
		Client:    &http.Client{Timeout: 45 * time.Second},
		WatchBase: "https://www.youtube.com/watch?v=",
		NextURL:   "https://www.youtube.com/youtubei/v1/next?prettyPrint=false",
		OEmbedURL: "https://www.youtube.com/oembed?format=json&url=",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		Attempts:  3,
		Log:       log,
	}
}

// Wanted reports whether an upload is worth a lookup: every upload, except one
// whose title already says it is the official video. Whether it is an art track
// is decided by ResolveLog, which can ask YouTube; the channel name cannot tell,
// because many art tracks carry the artist's own name instead of "- Topic".
func (r *Resolver) Wanted(title, _, _ string) bool {
	return !LooksLikeOfficialVideo(title)
}

// Version is the video ResolveVersion chose to download in place of an upload.
type Version struct {
	// ID is the video, "" when the upload stays.
	ID string
	// Official is true for the official video, false for the upload of the song with
	// the best quality, taken when there is no official one.
	Official bool
	// OwnOfficial is true when the upload itself is the official video (YouTube Music
	// tags it as one the artist uploaded), so ID is empty and nothing better is needed.
	OwnOfficial bool
	// Reason is a short explanation for the log.
	Reason string
}

// Resolve returns the video id of the official music video linked from the
// given art track, or "" when none could be found. The returned reason is a
// short human readable explanation for logging.
func (r *Resolver) Resolve(videoID string) (string, string) {
	return r.ResolveLog(videoID, "", "", 0, r.Log)
}

// ResolveLog is Resolve with a per-call log function, so concurrent
// callers can route messages to their own entry. The title and channel of
// the art track let it search YouTube for the video when the description
// links none. Only an official video is an answer: the best quality upload of
// ResolveVersion changes with every view, so it is not one to track.
func (r *Resolver) ResolveLog(videoID, title, channel string, durationSec int, logFn func(format string, a ...interface{})) (string, string) {
	version := r.ResolveVersion(videoID, title, channel, durationSec, logFn)
	if !version.Official {
		return "", version.Reason
	}
	return version.ID, version.Reason
}

// Identified is what the first half of a lookup learned about an upload: whether the
// answer is already known, and otherwise the song to search for.
type Identified struct {
	// Done is true when the answer is known (the description linked the official video,
	// or the upload is a plain video no database names), so Pick has nothing to do.
	Done bool
	// Version is the answer when Done.
	Version Version

	song   Song
	reason string
}

// Named returns the artist and the title a music database gave the song, when it did.
func (i Identified) Named() (artist, title string, ok bool) {
	if i.song.Title == "" || len(i.song.Artists) == 0 {
		return "", "", false
	}
	return i.song.Artists[0], i.song.Title, true
}

// ResolveVersion looks for the video to download in place of an upload: the
// official music video, or failing that the upload of the song with the best
// picture and sound. It is Identify followed by Pick, which an engine runs
// separately so that each can have its own workers.
func (r *Resolver) ResolveVersion(videoID, title, channel string, durationSec int, logFn func(format string, a ...interface{})) Version {
	identified := r.Identify(videoID, title, channel, durationSec, logFn)
	if identified.Done {
		return identified.Version
	}
	return r.Pick(identified, logFn)
}

// Identify is the first half of ResolveVersion: it tells art tracks from videos, follows
// the link of the description to the official video, and names the song through the
// music databases.
func (r *Resolver) Identify(videoID, title, channel string, durationSec int, logFn func(format string, a ...interface{})) Identified {
	logf := logWith(logFn)
	artTrack, info, why := r.isArtTrack(videoID, channel, logf)
	// YouTube Music names the artist better than a channel does ("Kate Bush", not
	// "KateBushMusic") and knows the length of the song.
	if info.Artist != "" {
		channel = info.Artist
	}
	if durationSec == 0 {
		durationSec = info.DurationSec
	}

	// A video YouTube Music says the artist uploaded as an official video is one already.
	if !artTrack && info.Type == "OMV" {
		return Identified{Done: true, Version: Version{OwnOfficial: true, Reason: why}}
	}

	reason := why
	if artTrack {
		var id string
		if id, reason = r.fromDescription(videoID, logf); id != "" {
			return Identified{Done: true, Version: Version{ID: id, Official: true, Reason: reason}}
		}
	}

	// A plain video is only searched for when a music database says what song it is.
	song, named := r.songOfUpload(videoID, title, channel, durationSec, artTrack, logf)
	if !named {
		return Identified{Done: true, Version: Version{Reason: reason}}
	}
	return Identified{song: song, reason: reason}
}

// Pick is the second half of ResolveVersion: it searches for the official video of the
// song under its right name and artist and, failing that, takes the upload of the song
// with the best picture and sound.
func (r *Resolver) Pick(identified Identified, logFn func(format string, a ...interface{})) Version {
	if identified.Done {
		return identified.Version
	}
	logf := logWith(logFn)
	found, foundWhy, results := r.fromSearch(identified.song, logf)
	if found != "" {
		return Version{ID: found, Official: true, Reason: foundWhy}
	}
	// Only an official video is wanted: the best quality upload is not looked for.
	if r.OfficialOnly {
		return Version{Reason: identified.reason}
	}
	if best, bestWhy := r.fromBestQuality(identified.song, results, logf); best != "" {
		return Version{ID: best, Reason: bestWhy}
	}
	return Version{Reason: identified.reason}
}

// logWith makes a log function that may be nil, and ends every message with a newline.
func logWith(logFn func(format string, a ...interface{})) func(format string, a ...interface{}) {
	return func(format string, a ...interface{}) {
		if logFn != nil {
			logFn(format+"\n", a...)
		}
	}
}

// fromDescription follows the video linked from the art track's page.
func (r *Resolver) fromDescription(videoID string, logf func(format string, a ...interface{})) (string, string) {
	var candidates []string

	// 1. Crawl the watch page like a browser would.
	html, err := r.get(r.WatchBase + videoID)
	if err != nil {
		logf("[official] %s: cannot fetch watch page: %v", videoID, err)
	} else {
		candidates = append(candidates, candidatesFromWatchPage(html, videoID)...)
	}

	// 2. The watch page sometimes ships a stripped down description panel.
	//    The innertube "next" call is what the page itself uses to fill it.
	if len(candidates) == 0 {
		nextJSON, err := r.fetchNext(videoID, html)
		if err != nil {
			logf("[official] %s: innertube next request failed: %v", videoID, err)
		} else {
			candidates = append(candidates, candidatesFromInitialData(nextJSON, videoID)...)
		}
	}

	// 3. Still nothing: let yt-dlp fetch the page with its cookies and
	//    bot-check workarounds, and read whatever it got.
	if len(candidates) == 0 && r.Dumper != nil {
		pages, err := r.Dumper(videoID)
		if err != nil {
			logf("[official] %s: yt-dlp page dump failed: %v", videoID, err)
		}
		for _, p := range pages {
			switch {
			case strings.Contains(p.URL, "/watch?"):
				candidates = append(candidates, candidatesFromWatchPage(p.Body, videoID)...)
			case strings.Contains(p.URL, "/youtubei/v1/next"):
				candidates = append(candidates, candidatesFromInitialData(p.Body, videoID)...)
			}
		}
		if len(candidates) > 0 {
			logf("[official] %s: found the link through yt-dlp", videoID)
		}
	}

	if len(candidates) == 0 {
		return "", "no official video link found in description"
	}

	for _, cand := range uniqueStrings(candidates) {
		author, ok, err := r.lookupAuthor(cand)
		switch {
		case err != nil:
			// oEmbed can refuse (401) for videos that disallow embedding.
			// The link came from YouTube itself, so accept it.
			logf("[official] %s: oembed check for %s inconclusive (%v), accepting", videoID, cand, err)
			return cand, "linked from description (unverified channel)"
		case !ok:
			logf("[official] %s: candidate %s does not exist, skipping", videoID, cand)
		case IsTopicChannel(author):
			logf("[official] %s: candidate %s is another auto-generated track (%s), skipping", videoID, cand, author)
		default:
			return cand, fmt.Sprintf("official video by %q", author)
		}
	}

	return "", "linked videos were not official uploads"
}
