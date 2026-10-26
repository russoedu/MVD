package workspace

import (
	"net/url"
	"regexp"
	"strings"

	"youtube-downloader/libs/mvd-core/musicsearch"
	"youtube-downloader/libs/mvd-core/songfile"
)

// looksLikeAddress is a web address written without its scheme, such as
// "www.youtube.com/playlist?list=PL1" or "open.spotify.com/playlist/37i9": a host with a dot, and
// then nothing or a path.
var looksLikeAddress = regexp.MustCompile(`(?i)^[a-z0-9][a-z0-9-]*(\.[a-z0-9-]+)+\.?([/?#].*)?$`)

// parseLinks reads what was written or pasted: one or several sources, separated by line breaks or
// by spaces, as a paste of several links comes (a line editor turns breaks into spaces). It gives
// the sources, or why the text is not made of them. A source is a web address (http or https),
// written with or without its scheme, or the path of a song file (.txt, .csv or .tsv); the path may
// be in quotes and have spaces. Nothing is looked up here: a link that does not exist is for the
// run to find.
func parseLinks(text string) (links []string, reason string) {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if _, ok := songfile.FilePath(line); ok {
			links = append(links, line)
			continue
		}
		parts := strings.Fields(line)
		for _, part := range parts {
			if reason := notALinkWord(part, len(parts) > 1); reason != "" {
				return nil, reason
			}
			links = append(links, part)
		}
	}

	return links, ""
}

// notALink says why text cannot be made of sources, or "" when it can.
func notALink(text string) string {
	_, reason := parseLinks(text)

	return reason
}

// notALinkWord says why one word is not a source, or "" when it is.
func notALinkWord(word string, among bool) string {
	if musicsearch.Is(word) {
		if _, ok := musicsearch.Parse(word); !ok {
			return "That search has no artist"
		}

		return ""
	}
	if strings.Contains(word, "://") {
		u, err := url.Parse(word)
		switch {
		case err != nil || u.Host == "":
			return "That is not a complete web address. Paste the whole link, such as https://www.youtube.com/playlist?list=..."
		case u.Scheme != "http" && u.Scheme != "https":
			return "Only web addresses (http or https) can be used, not " + u.Scheme + "://"
		}

		return ""
	}
	if looksLikeAddress.MatchString(word) {
		return ""
	}
	if among {
		return "That has words that are not links in it. Paste only links, one or several."
	}

	return "That does not look like a link. Paste a YouTube, Spotify or Apple Music link."
}
