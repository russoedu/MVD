package urllist

import (
	"strconv"
	"strings"

	"youtube-downloader/libs/mvd-core/wording"
)

// QuitQuestion is what the screen asks before it throws the list away.
const QuitQuestion = "Discard the list and quit?"

// Heading is the name of the screen.
const Heading = "MVD · Download list"

// List is the download list being edited: its text, one link or file per line, and whether the
// person is being asked to discard it.
type List struct {
	// Text is the whole text of the list.
	Text string
	// ConfirmQuit is true while the question to discard the list waits for an answer.
	ConfirmQuit bool
}

// New is a list holding the given lines.
func New(urls []string) List { return List{Text: strings.Join(urls, "\n")} }

// SetText is the text as the renderer's text box holds it.
func (l *List) SetText(text string) { l.Text = text }

// URLs are the trimmed, non-empty lines.
func (l List) URLs() []string {
	var out []string
	for _, line := range strings.Split(l.Text, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			out = append(out, t)
		}
	}

	return out
}

// Count tells how many lines there are, as "(3 items)", and nothing for an empty list.
func (l List) Count() string {
	n := len(l.URLs())
	if n == 0 {
		return ""
	}

	return "(" + strconv.Itoa(n) + " item" + wording.Plural(n) + ")"
}

// AddURL adds a line to the list unless it is already there.
func (l *List) AddURL(url string) {
	for _, existing := range l.URLs() {
		if existing == url {
			return
		}
	}
	value := strings.TrimRight(l.Text, "\n")
	if value != "" {
		value += "\n"
	}
	l.Text = value + url
}
