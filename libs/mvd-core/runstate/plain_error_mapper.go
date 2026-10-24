package runstate

import (
	"regexp"
	"strings"
)

// toolPrefix is how a failed listing starts: the name of the tool and the option it ran with.
var toolPrefix = regexp.MustCompile(`^yt-dlp[^:]*failed:\s*`)

// errorTag is the "[youtube] jNQXAC9IVRw:" or "[generic]" that yt-dlp puts before a message.
var errorTag = regexp.MustCompile(`^(ERROR:\s*)?(\[[^\]]+\]\s*)+([A-Za-z0-9_-]{6,}:\s*)?`)

// plainErrors are the failures people meet most, in the words of the tool and what to say instead.
// The first that matches wins, so the more particular come first.
var plainErrors = []struct {
	has []string
	say string
}{
	{[]string{"is not a valid url", "unsupported url"},
		"That is not a link MVD can use. Paste a YouTube, Spotify or Apple Music link."},
	{[]string{"sign in to confirm your age", "age-restricted", "age restricted", "confirm your age"},
		"This video is age-restricted. In Preferences > Advanced, choose your browser under \"Use my browser's YouTube login\"."},
	{[]string{"not a bot", "confirm you’re not a bot", "confirm you're not a bot"},
		"YouTube wants to be sure it is a person. In Preferences > Advanced, choose your browser under \"Use my browser's YouTube login\"."},
	{[]string{"private video", "this video is private"},
		"This video is private, so it cannot be downloaded."},
	{[]string{"not available in your country", "blocked it in your country", "blocked in your country", "geo restricted", "geo-restricted"},
		"This video is not available in your country."},
	{[]string{"video unavailable", "this video is unavailable", "video is no longer available", "has been removed", "been terminated", "does not exist", "this video has been"},
		"This video is unavailable: it may be private, removed or not exist."},
	{[]string{"http error 429", "too many requests", "rate limit"},
		"YouTube is limiting requests for now. Wait a few minutes, then retry."},
	{[]string{"getaddrinfo", "name resolution", "unable to download webpage", "timed out", "connection reset", "connection refused", "network is unreachable", "no route to host", "temporary failure", "ssl", "unable to connect"},
		"There is no connection to the internet, or the site did not answer. Check the connection, then retry."},
}

// PlainError says in everyday words what a failure was, and what to do, for the failures people
// meet most; for the others it is the message of the tool without its name, its options and the
// "[youtube] id:" in front. The full text stays in the output of the song.
func PlainError(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}
	lower := strings.ToLower(text)
	for _, e := range plainErrors {
		for _, has := range e.has {
			if strings.Contains(lower, has) {
				return e.say
			}
		}
	}
	text = toolPrefix.ReplaceAllString(text, "")
	text = errorTag.ReplaceAllString(text, "")
	text = strings.TrimSpace(text)
	if text == "" {
		return strings.TrimSpace(raw)
	}

	return strings.ToUpper(text[:1]) + text[1:]
}
