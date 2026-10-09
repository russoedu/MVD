package runner

import (
	"context"
	"sync"

	"youtube-downloader/libs/mvd-core/ytdlp"
)

// videoProbes asks yt-dlp about each video once: its quality and its storyboard come
// from the same run, and several songs of a playlist may ask about the same upload.
type videoProbes struct {
	ctx       context.Context
	ytDlp     string
	extraArgs []string

	mu     sync.Mutex
	probes map[string]*probeEntry
}

type probeEntry struct {
	once  sync.Once
	probe ytdlp.VideoProbe
	err   error
}

func newVideoProbes(ctx context.Context, ytDlp string, extraArgs []string) *videoProbes {
	return &videoProbes{ctx: ctx, ytDlp: ytDlp, extraArgs: extraArgs, probes: map[string]*probeEntry{}}
}

// of returns what yt-dlp says of a video, asking only the first time.
func (p *videoProbes) of(videoID string) (ytdlp.VideoProbe, error) {
	p.mu.Lock()
	entry, ok := p.probes[videoID]
	if !ok {
		entry = &probeEntry{}
		p.probes[videoID] = entry
	}
	p.mu.Unlock()

	entry.once.Do(func() {
		entry.probe, entry.err = ytdlp.ProbeVideo(p.ctx, p.ytDlp, videoID, p.extraArgs)
	})
	return entry.probe, entry.err
}
