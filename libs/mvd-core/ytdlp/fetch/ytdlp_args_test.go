package fetch

import (
	"strings"
	"testing"
	"youtube-downloader/libs/mvd-core/ytdlp"
)

func TestDownloadArgs(t *testing.T) {
	got := DownloadArgs(DownloadOptions{Format: "best", OutputTemplate: "out/%(title)s.%(ext)s", MergeOutputFormat: "mp4", ConcurrentFragments: 4, ExtraArgs: []string{"-4"}}, "--no-playlist", "https://www.youtube.com/watch?v=abc")
	want := "-f best -o out/%(title)s.%(ext)s --merge-output-format mp4 --concurrent-fragments 4 -4 --no-playlist https://www.youtube.com/watch?v=abc"
	if strings.Join(got, " ") != want {
		t.Errorf("want %q, got %q", want, strings.Join(got, " "))
	}

	// ConcurrentFragments <= 0 omits the flag.
	bare := DownloadArgs(DownloadOptions{Format: "best", OutputTemplate: "o", ConcurrentFragments: 0}, "url")
	if strings.Contains(strings.Join(bare, " "), "concurrent-fragments") {
		t.Errorf("zero fragments should omit the flag: %v", bare)
	}
}

func TestApplyPlaylistFields(t *testing.T) {
	e := ytdlp.PlaylistEntry{
		Playlist:      "90s UK Dance Hits: Vol/1",
		PlaylistTitle: "90s UK Dance Hits: Vol/1",
		PlaylistID:    "PLnEEDxeHD6BPvM-iOmSeKx5N6Uug3vZQa",
		PlaylistIndex: 7,
		PlaylistCount: 90,
	}
	cases := map[string]string{
		"%(playlist_title,playlist)s/%(playlist_index)02d - %(title)s.%(ext)s": "90s UK Dance Hits： Vol⧸1/07 - %(title)s.%(ext)s",
		"%(title)s.%(ext)s": "%(title)s.%(ext)s",
		"%(playlist)s/%(playlist_index)s of %(playlist_count)d.%(ext)s": "90s UK Dance Hits： Vol⧸1/7 of 90.%(ext)s",
		"%(playlist_id)s/%(playlist_index)03d.%(ext)s":                  "PLnEEDxeHD6BPvM-iOmSeKx5N6Uug3vZQa/007.%(ext)s",
		"%(playlist_uploader)s/%(title)s.%(ext)s":                       "%(playlist_uploader)s/%(title)s.%(ext)s",
		"%(playlist_title|Singles)s/%(title)s.%(ext)s":                  "90s UK Dance Hits： Vol⧸1/%(title)s.%(ext)s",
	}
	for in, want := range cases {
		if got := ApplyPlaylistFields(in, e); got != want {
			t.Errorf("template %q:\n want %q\n got  %q", in, want, got)
		}
	}
	if got := ApplyPlaylistFields("%(playlist_title|Singles)s/%(playlist_index)02d.%(ext)s", ytdlp.PlaylistEntry{}); got != "Singles/.%(ext)s" {
		t.Errorf("defaults: got %q", got)
	}
}

func TestAVideoGivenOnItsOwnIsNamedAfterItselfWithNoPlaylistFolderOrNumber(t *testing.T) {
	alone := ytdlp.PlaylistEntry{ID: "jNQXAC9IVRw", Title: "Me at the zoo"}
	cases := map[string]string{
		"%(playlist_title,playlist)s/%(playlist_index)02d - %(title)s.%(ext)s": "%(title)s.%(ext)s",
		"%(playlist_index)02d - %(title)s.%(ext)s":                             "%(title)s.%(ext)s",
		"%(playlist_title)s/%(playlist_index)s - %(title)s [%(id)s].%(ext)s":   "%(title)s [%(id)s].%(ext)s",
		"Music/%(playlist_title,playlist)s/%(title)s.%(ext)s":                  "Music/%(title)s.%(ext)s",
		"%(title)s.%(ext)s":                            "%(title)s.%(ext)s",
		"%(playlist_title|Singles)s/%(title)s.%(ext)s": "Singles/%(title)s.%(ext)s",
		"%(playlist_index)02d":                         "%(title)s.%(ext)s",
	}
	for in, want := range cases {
		if got := ApplyPlaylistFields(in, alone); got != want {
			t.Errorf("template %q:\n want %q\n got  %q", in, want, got)
		}
	}
}

func TestAnEntryOfARealPlaylistKeepsNAForAFieldThatIsMissing(t *testing.T) {
	inPlaylist := ytdlp.PlaylistEntry{ID: "x", PlaylistTitle: "Best of"}

	got := ApplyPlaylistFields("%(playlist_title)s/%(playlist_index)02d - %(title)s.%(ext)s", inPlaylist)

	if want := "Best of/NA - %(title)s.%(ext)s"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSanitizeFilename(t *testing.T) {
	cases := map[string]string{
		`AC/DC: "Back" <in> Black?*|\`: "AC⧸DC： ＂Back＂ ＜in＞ Black？＊｜⧹",
		"trailing dots... ":            "trailing dots",
		"":                             "_",
		"line\nbreak":                  "line break",
	}
	for in, want := range cases {
		if got := sanitizeFilename(in); got != want {
			t.Errorf("%q: want %q, got %q", in, want, got)
		}
	}
}

func TestParseProgressLine(t *testing.T) {
	p, ok := ParseProgressLine("MVD|512|NA|1024|2048.5|7")
	if !ok || p.Downloaded != 512 || p.Total != 1024 || p.Percent != 50 || p.Speed != 2048.5 || p.ETA != 7 {
		t.Errorf("unexpected progress: %+v ok=%v", p, ok)
	}
	p, ok = ParseProgressLine("MVD|512|NA|NA|NA|NA")
	if !ok || p.Total != 0 || p.Percent != 0 || p.Speed != 0 || p.ETA != -1 {
		t.Errorf("unknown totals should be zero: %+v", p)
	}
	if _, ok := ParseProgressLine("[download]  34.2% of 112.4MiB"); ok {
		t.Error("normal yt-dlp lines are not progress lines")
	}
	if _, ok := ParseProgressLine("MVD|1|2"); ok {
		t.Error("short lines are not progress lines")
	}
	if !IsPostProcessLine(`[Merger] Merging formats into "x.mp4"`) || IsPostProcessLine("[download] Destination: x") {
		t.Error("post process detection wrong")
	}
}

func TestDownloadArgsKeepUnfinishedFilesInTheTempDir(t *testing.T) {
	got := DownloadArgs(DownloadOptions{Format: "best", OutputTemplate: "%(title)s.%(ext)s", HomeDir: "/music", TempDir: "/tmp/partial"}, "url")
	want := []string{"-f", "best", "-P", "home:/music", "-P", "temp:/tmp/partial", "-o", "%(title)s.%(ext)s", "url"}
	if len(got) != len(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("arg %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestProgressLineWithFragments(t *testing.T) {
	p, ok := ParseProgressLine("MVD|512|1024|NA|2048.5|7|3|120")
	if !ok || p.FragmentIndex != 3 || p.FragmentCount != 120 || p.Speed != 2048.5 {
		t.Errorf("got %+v, %v", p, ok)
	}
	p, ok = ParseProgressLine("MVD|512|1024|NA|2048.5|7|NA|NA")
	if !ok || p.FragmentIndex != 0 || p.FragmentCount != 0 {
		t.Errorf("a part without fragments: %+v, %v", p, ok)
	}
}
