package ytdlp

import (
	"context"
	"os"
	"strings"
	"testing"
)

const flatPlaylistSample = `{"title": "You Don't Know Me (feat. Duane Harden) (Radio Edit)", "channel": "Armand Van Helden - Topic", "uploader": "Armand Van Helden - Topic", "id": "rnlp_avexYQ", "url": "https://www.youtube.com/watch?v=rnlp_avexYQ", "playlist": "90s UK Dance Hits", "playlist_title": "90s UK Dance Hits", "playlist_id": "PLnEEDxeHD6BPvM-iOmSeKx5N6Uug3vZQa", "playlist_index": 1, "playlist_count": 90}

{"title": "Modjo - Lady (Hear Me Tonight)", "channel": "ModjoOfficial", "uploader": "ModjoOfficial", "id": "MR3uP7IYz44", "url": "https://www.youtube.com/watch?v=MR3uP7IYz44", "playlist": "90s UK Dance Hits", "playlist_title": "90s UK Dance Hits", "playlist_id": "PLnEEDxeHD6BPvM-iOmSeKx5N6Uug3vZQa", "playlist_index": 3, "playlist_count": 90}
`

func TestParsePlaylistEntries(t *testing.T) {
	entries, err := ParsePlaylistEntries(strings.NewReader(flatPlaylistSample))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(entries))
	}
	if entries[0].ID != "rnlp_avexYQ" || entries[0].PlaylistIndex != 1 || entries[0].PlaylistTitle != "90s UK Dance Hits" || entries[0].PlaylistCount != 90 {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}
}

func TestListPlaylist(t *testing.T) {
	bin := useStub(t)
	entries, err := ListPlaylist(context.Background(), bin, "https://youtube.com/playlist?list=A", nil)
	if err != nil || len(entries) != 2 || entries[1].PlaylistIndex != 2 {
		t.Fatalf("unexpected listing %v, %v", entries, err)
	}
	if _, err := ListPlaylist(context.Background(), bin, "https://youtube.com/playlist?list=NOPE", nil); err == nil || !strings.Contains(err.Error(), "Unable to recognize playlist") {
		t.Fatalf("expected yt-dlp's error, got %v", err)
	}
}

func TestDownload(t *testing.T) {
	bin := useStub(t)
	var lines []string
	err := Download(context.Background(), bin, []string{"https://www.youtube.com/watch?v=aaaaaaaaaa1"}, func(l string) { lines = append(lines, l) })
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 3 || !strings.HasPrefix(lines[1], "MVD|") || !IsPostProcessLine(lines[2]) {
		t.Errorf("unexpected lines %q", lines)
	}

	err = Download(context.Background(), bin, []string{"https://www.youtube.com/watch?v=failfailfai"}, func(string) {})
	if err == nil || !strings.Contains(err.Error(), "Video unavailable") {
		t.Errorf("expected yt-dlp's error line, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Download(ctx, bin, []string{"https://www.youtube.com/watch?v=aaaaaaaaaa1"}, func(string) {}); err == nil {
		t.Error("cancelled context should fail")
	}
}

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
	e := PlaylistEntry{
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
	if got := ApplyPlaylistFields("%(playlist_title|Singles)s/%(playlist_index)02d.%(ext)s", PlaylistEntry{}); got != "Singles/.%(ext)s" {
		t.Errorf("defaults: got %q", got)
	}
}

func TestAVideoGivenOnItsOwnIsNamedAfterItselfWithNoPlaylistFolderOrNumber(t *testing.T) {
	alone := PlaylistEntry{ID: "jNQXAC9IVRw", Title: "Me at the zoo"}
	cases := map[string]string{
		"%(playlist_title,playlist)s/%(playlist_index)02d - %(title)s.%(ext)s": "%(title)s.%(ext)s",
		"%(playlist_index)02d - %(title)s.%(ext)s":                              "%(title)s.%(ext)s",
		"%(playlist_title)s/%(playlist_index)s - %(title)s [%(id)s].%(ext)s":    "%(title)s [%(id)s].%(ext)s",
		"Music/%(playlist_title,playlist)s/%(title)s.%(ext)s":                   "Music/%(title)s.%(ext)s",
		"%(title)s.%(ext)s":                                                     "%(title)s.%(ext)s",
		"%(playlist_title|Singles)s/%(title)s.%(ext)s":                          "Singles/%(title)s.%(ext)s",
		"%(playlist_index)02d":                                                  "%(title)s.%(ext)s",
	}
	for in, want := range cases {
		if got := ApplyPlaylistFields(in, alone); got != want {
			t.Errorf("template %q:\n want %q\n got  %q", in, want, got)
		}
	}
}

func TestAnEntryOfARealPlaylistKeepsNAForAFieldThatIsMissing(t *testing.T) {
	inPlaylist := PlaylistEntry{ID: "x", PlaylistTitle: "Best of"}

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

func TestExportCookies(t *testing.T) {
	bin := useStub(t)
	file := t.TempDir() + "/cookies.txt"
	if err := ExportCookies(context.Background(), bin, "edge", file, "https://youtube.com/playlist?list=A", nil); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(file); err != nil || !strings.Contains(string(data), "SID") {
		t.Errorf("cookie file not written: %v %q", err, data)
	}
	if err := ExportCookies(context.Background(), bin, "nope", file, "https://youtube.com/playlist?list=A", nil); err == nil || !strings.Contains(err.Error(), "could not find nope") {
		t.Errorf("expected yt-dlp's error, got %v", err)
	}
	if got := CookieArgs(""); got != nil {
		t.Errorf("no file should mean no args, got %v", got)
	}
	if got := strings.Join(CookieArgs("c.txt"), " "); got != "--cookies c.txt" {
		t.Errorf("unexpected cookie args %q", got)
	}
}

func TestDumpPages(t *testing.T) {
	bin := useStub(t)
	pages, err := DumpPages(context.Background(), bin, "https://www.youtube.com/watch?v=aaaaaaaaaa1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 || !strings.Contains(pages[0].URL, "/watch?v=aaaaaaaaaa1") || !strings.Contains(string(pages[0].Body), "DUMPEDOFF01") || !strings.Contains(pages[1].URL, "/youtubei/v1/player") {
		t.Errorf("unexpected pages: %+v", pages)
	}

	parsed := ParseDumpedPages("[youtube:tab] Extracting URL: x\n[youtube:tab] Dumping request to https://a\naGVsbG8=\n[download] done\n[youtube] Dumping request to https://b\nnot base64!!\n")
	if len(parsed) != 1 || parsed[0].URL != "https://a" || string(parsed[0].Body) != "hello" {
		t.Errorf("unexpected parse result %+v", parsed)
	}
}
