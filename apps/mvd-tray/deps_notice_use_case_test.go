package main

import (
	"errors"
	"strings"
	"testing"
	"youtube-downloader/apps/mvd-tray/notification"
	"youtube-downloader/libs/mvd-core/deps"
)

func TestDependencyEventsBecomeLogLinesAndOnlyTwoKindsOfNotification(t *testing.T) {
	var notices []notification.Notice
	var logs []string
	report := depsReporter(
		func(format string, a ...interface{}) { logs = append(logs, format) },
		func(n notification.Notice) { notices = append(notices, n) },
	)

	report(deps.Event{Kind: deps.EventMissing, Names: []string{"yt-dlp", "ffmpeg"}, Dir: `C:\Users\me\AppData\Roaming\mvd\bin`})
	report(deps.Event{Kind: deps.EventDownloading, Name: "yt-dlp", URL: "https://example.test/yt-dlp"})
	report(deps.Event{Kind: deps.EventInstalled, Name: "yt-dlp"})
	report(deps.Event{Kind: deps.EventFailed, Name: "ffmpeg", Err: errors.New("download failed: HTTP status 503")})

	if len(logs) != 4 {
		t.Errorf("logged %d lines, want 4", len(logs))
	}
	if len(notices) != 2 {
		t.Fatalf("notified %d times, want 2 (start and failure): %+v", len(notices), notices)
	}
	start, failed := notices[0], notices[1]
	if start.Failure || !strings.Contains(start.Text, "yt-dlp, ffmpeg") || !strings.Contains(start.Text, `AppData\Roaming\mvd\bin`) {
		t.Errorf("the start notice should say what is downloaded and where: %+v", start)
	}
	if !failed.Failure || !strings.Contains(failed.Title, "ffmpeg") || !strings.Contains(failed.Text, "503") {
		t.Errorf("the failure notice should name the tool and the reason: %+v", failed)
	}
}
