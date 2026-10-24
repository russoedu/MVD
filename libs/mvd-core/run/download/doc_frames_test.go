package download

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/look"
)

// docFrameW and docFrameH are the size of the frames kept for the documentation, in cells.
const docFrameW, docFrameH = 120, 30

// docModel is a download in progress, as a person sees it: two playlists, songs in most states,
// with their album and quality.
func docModel(t *testing.T) Model {
	t.Helper()
	c := &fakeController{events: make(chan interface{}), sources: []engine.PlaylistSource{{Index: 0, URL: "https://youtube.com/playlist?list=a"}, {Index: 1, URL: "https://youtube.com/playlist?list=b"}}}
	m := New(c)
	m.loaded = true
	m.state.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "90s Dance Hits", Entries: []engine.EntryInfo{
		{ID: 0, Playlist: 0, Index: 1, VideoID: "a1", Title: "You Don't Know Me (feat. Duane Harden) (Radio Edit)", Channel: "Armand Van Helden - Topic"},
		{ID: 1, Playlist: 0, Index: 2, VideoID: "a2", Title: "Don't Call Me Baby", Channel: "Madison Avenue - Topic"},
		{ID: 2, Playlist: 0, Index: 3, VideoID: "a3", Title: "Modjo - Lady (Hear Me Tonight) [Official Video]", Channel: "Modjo"},
		{ID: 3, Playlist: 0, Index: 4, VideoID: "a4", Title: "Bob Sinclar - I Feel For You", Channel: "Bob Sinclar"},
		{ID: 4, Playlist: 0, Index: 5, VideoID: "a5", Title: "Stardust - Music Sounds Better With You", Channel: "Stardust"},
		{ID: 5, Playlist: 0, Index: 6, VideoID: "a6", Title: "The Bucketheads - The Bomb! (These Sounds Fall Into My Mind)", Channel: "The Bucketheads - Topic"},
		{ID: 6, Playlist: 0, Index: 7, VideoID: "a7", Title: "Haddaway - What Is Love", Channel: "Haddaway"},
	}})
	m.state.Apply(engine.EvPlaylistListed{Playlist: 1, Title: "Road trip", Entries: []engine.EntryInfo{
		{ID: 7, Playlist: 1, Index: 1, VideoID: "b1", Title: "Daft Punk - One More Time", Channel: "Daft Punk"},
		{ID: 8, Playlist: 1, Index: 2, VideoID: "b2", Title: "Cassius - Toop Toop", Channel: "Cassius"},
	}})
	done := func(id int, artist, title, album string, height, kbps int, official bool) {
		m.state.Apply(engine.EvEntryState{Entry: id, State: engine.StateDownloading, TargetID: "t", Artist: artist, Title: title, Album: album})
		m.state.Apply(engine.EvQuality{Entry: id, Height: height, AudioKbps: kbps})
		m.state.Apply(engine.EvEntryState{Entry: id, State: engine.StateDone, TargetID: "t", Official: official, Artist: artist, Title: title, Album: album})
	}
	done(0, "Armand Van Helden", "You Don't Know Me", "Killing Puppets", 1080, 0, true)
	m.state.Apply(engine.EvEntryState{Entry: 1, State: engine.StateDuplicate, TargetID: "t"})
	m.state.Apply(engine.EvEntryState{Entry: 2, State: engine.StateFailed, TargetID: "t", Err: "Video unavailable"})
	done(3, "Bob Sinclar", "I Feel For You", "Western Dream", 720, 0, false)
	m.state.Apply(engine.EvEntryState{Entry: 4, State: engine.StateDownloading, TargetID: "t", Artist: "Stardust", Title: "Music Sounds Better With You", Album: "The Ride"})
	m.state.Apply(engine.EvQuality{Entry: 4, Height: 1080})
	m.state.Apply(engine.EvProgress{Entry: 4, Percent: 62, Downloaded: 40_000_000, Total: 64_000_000, Speed: 8.1 * 1024 * 1024, ETA: 3})
	m.state.Apply(engine.EvEntryState{Entry: 5, State: engine.StateResolving, TargetID: "t"})
	m.state.Apply(engine.EvEntryState{Entry: 8, State: engine.StateNotFound, TargetID: "t"})
	done(7, "Daft Punk", "One More Time", "Discovery", 1080, 0, true)

	m.width, m.height = docFrameW, docFrameH
	m.focus = paneEntries
	m.selEntry = 4

	return m
}

// TestDocFrames writes the download screen in each theme of the app as text with its colours, to
// the folder in MVD_DOC_FRAMES, for the screenshots of the documentation (docs/screenshots). It
// does nothing without it.
func TestDocFrames(t *testing.T) {
	dir := os.Getenv("MVD_DOC_FRAMES")
	if dir == "" {
		t.Skip("MVD_DOC_FRAMES is not set")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { look.Apply(config.DefaultThemeColors()) })

	for _, preset := range config.ThemePresets {
		look.Apply(preset.Colors)
		name := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(preset.Name, "+", "-plus"), " ", "-"))
		m := docModel(t)
		write := func(file string, m Model) {
			if err := os.WriteFile(filepath.Join(dir, file+".ansi"), []byte(m.View().Content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		write(name, m)
		if preset.Name == config.DefaultThemeName {
			m.technical = true
			write(name+"-technical", m)
		}
	}
}
