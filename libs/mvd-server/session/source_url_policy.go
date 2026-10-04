package session

import (
	"net/url"
	"strings"
)

// NormalizeURLs sorts what a user pasted into the URLs worth handing to yt-dlp
// and the lines that are not.
//
// Blank lines and `#` comments are ignored without a word, since a pasted list
// has them. What is left must be an absolute http or https URL with a host;
// anything else is reported back as rejected, so the UI can say which line it
// did not understand instead of dropping it. A URL repeated in one paste is kept
// once, in the order it first appeared.
func NormalizeURLs(raw []string) (accepted, rejected []string) {
	seen := map[string]bool{}
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parsed, err := url.Parse(line)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			rejected = append(rejected, line)
			continue
		}
		if seen[line] {
			continue
		}
		seen[line] = true
		accepted = append(accepted, line)
	}
	return accepted, rejected
}
