package songpaste

import (
	"strconv"
	"time"

	"youtube-downloader/libs/mvd-core/songfile"
	"youtube-downloader/libs/mvd-core/wording"
)

// Heading is the name of the screen.
const Heading = "MVD · Add songs"

// TextLabel names the text pasted.
const TextLabel = "Songs, one Artist - Title per line"

// Outcome is where the screen leads when the person asks for it: the host acts on it.
type Outcome int

const (
	// None stays on the screen.
	None Outcome = iota
	// Added says the songs were saved as a file the download list can name (see Paste.Saved).
	Added
	// Cancel leaves without saving.
	Cancel
)

// Paste is the screen for pasting songs: one "Artist - Title" per line, or a CSV exported from a
// music service. Saving stores the list as a file that the download list then names, so any
// service can feed the app.
type Paste struct {
	// Text is the whole text pasted.
	Text string
	// Saved is the file the last save wrote.
	Saved string
	// Err is why the last save failed, "" when it did not.
	Err string
	dir string
}

// New is an empty paste that saves its files in dir.
func New(dir string) Paste { return Paste{dir: dir} }

// SetText is the text as the renderer's text box holds it.
func (p *Paste) SetText(text string) { p.Text = text }

// ClearError forgets the failure of the last save, as the first thing typed after it does.
func (p *Paste) ClearError() { p.Err = "" }

// Save writes the songs to a file in the folder of the lists; a failure is left in Err.
func (p *Paste) Save(now time.Time) Outcome {
	path, err := songfile.Save(p.dir, p.Text, now)
	if err != nil {
		p.Err = err.Error()

		return None
	}
	p.Saved, p.Err = path, ""

	return Added
}

// Leave gives up without saving.
func (*Paste) Leave() Outcome { return Cancel }

// Summary tells how the text reads: how many songs, and how many lines no song could be read
// from. An error from the last save replaces it.
func (p Paste) Summary() string {
	if p.Err != "" {
		return p.Err
	}
	list := songfile.Parse(p.Text)
	text := strconv.Itoa(len(list.Songs)) + " song" + wording.Plural(len(list.Songs))
	if list.Skipped > 0 {
		text += ", " + strconv.Itoa(list.Skipped) + " line" + wording.Plural(list.Skipped) + " not understood"
	}

	return text
}
