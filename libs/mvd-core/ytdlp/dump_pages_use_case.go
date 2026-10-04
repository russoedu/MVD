package ytdlp

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"

	"youtube-downloader/libs/mvd-core/procwindow"
)

// Page is one response yt-dlp fetched while extracting a URL.
type Page struct {
	URL  string
	Body []byte
}

// DumpPages makes yt-dlp extract a URL without downloading and returns
// every page and API response it fetched on the way, in order. It lets
// other code read a YouTube page with yt-dlp's cookies, client
// impersonation and bot-check workarounds.
func DumpPages(ctx context.Context, bin, url string, extraArgs []string) ([]Page, error) {
	args := []string{"--dump-pages", "--simulate", "--skip-download", "--no-playlist", "--no-warnings"}
	args = append(args, extraArgs...)
	args = append(args, url)

	cmd := exec.CommandContext(ctx, bin, args...)
	procwindow.Hide(cmd)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	pages := ParseDumpedPages(string(out))
	if err != nil && len(pages) == 0 {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("yt-dlp --dump-pages failed: %s", lastLine(msg))
	}
	return pages, nil
}

// ParseDumpedPages decodes the "[ie] Dumping request to URL" lines, each
// followed by one base64 line, that --dump-pages prints.
func ParseDumpedPages(out string) []Page {
	const marker = "] Dumping request to "
	var pages []Page
	scanner := bufio.NewScanner(strings.NewReader(out))
	scanner.Buffer(make([]byte, 0, 1024*1024), 64*1024*1024)
	pendingURL := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if pendingURL != "" {
			if body, err := base64.StdEncoding.DecodeString(line); err == nil {
				pages = append(pages, Page{URL: pendingURL, Body: body})
			}
			pendingURL = ""
			continue
		}
		if i := strings.Index(line, marker); i >= 0 && strings.HasPrefix(line, "[") {
			pendingURL = strings.TrimSpace(line[i+len(marker):])
		}
	}
	return pages
}
