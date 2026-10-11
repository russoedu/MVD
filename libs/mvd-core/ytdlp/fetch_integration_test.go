package ytdlp_test

import (
	"context"
	"strings"
	"testing"

	"youtube-downloader/libs/mvd-core/ytdlp"
	"youtube-downloader/libs/mvd-core/ytdlp/fetch"
)

func TestDownload(t *testing.T) {
	bin := ytdlp.UseStub(t)
	var lines []string
	err := fetch.Download(context.Background(), bin, []string{"https://www.youtube.com/watch?v=aaaaaaaaaa1"}, func(l string) { lines = append(lines, l) })
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 3 || !strings.HasPrefix(lines[1], "MVD|") || !fetch.IsPostProcessLine(lines[2]) {
		t.Errorf("unexpected lines %q", lines)
	}

	err = fetch.Download(context.Background(), bin, []string{"https://www.youtube.com/watch?v=failfailfai"}, func(string) {})
	if err == nil || !strings.Contains(err.Error(), "Video unavailable") {
		t.Errorf("expected yt-dlp's error line, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := fetch.Download(ctx, bin, []string{"https://www.youtube.com/watch?v=aaaaaaaaaa1"}, func(string) {}); err == nil {
		t.Error("cancelled context should fail")
	}
}
