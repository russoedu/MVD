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
