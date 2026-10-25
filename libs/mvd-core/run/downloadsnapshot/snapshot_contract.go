package downloadsnapshot

import "youtube-downloader/libs/mvd-core/songlist"

// ConfirmQuitQuestion is asked when the person leaves while the downloads go.
const ConfirmQuitQuestion = "Downloads are still running. Quit?"

// CounterKind is what a counter of the header counts.
type CounterKind int

// The counters of the header, in the order they are shown.
const (
	CounterWaiting CounterKind = iota
	CounterDownloading
	CounterDone
	CounterFailed
	CounterOfficialFound
	CounterOriginalOnly
	CounterAccepted
	CounterOfficialVideo
	CounterBetterQuality
	CounterAlreadyHad
	CounterNoOfficial
	CounterRejected
	CounterTriedAgain
)

// Counter is one count of the header.
type Counter struct {
	Kind  CounterKind
	Label string
	Count int
	Tone  songlist.Tone
	// Moving says that something is going on, so a renderer may animate it.
	Moving bool
}

// Header is the counters and the label on the top edge.
type Header struct {
	Counters []Counter
	// Right is the state of the run and the time it has taken.
	Right string
}

// Hint is a key on offer and what it does.
type Hint struct{ Key, Desc string }

// Line is one line of the details: a sentence with the tone it is said in. Blank is an empty line.
type Line struct {
	Text   string
	Tone   songlist.Tone
	Strong bool
	Blank  bool
	// SongTitle says the line is the title of a song as the listing gave it, which a renderer with
	// narrow room for symbols may want to clean.
	SongTitle bool
}

// GlyphKind is the mark in front of a playlist or a song; a renderer picks the character or the
// icon.
type GlyphKind int

// The marks.
const (
	GlyphNone GlyphKind = iota
	GlyphIdle
	GlyphBusy
	GlyphResolving
	GlyphMerging
	GlyphDone
	GlyphHad
	GlyphMissing
	GlyphFailed
	GlyphChoose
	GlyphAccepted
	GlyphRejected
)

// Glyph is a mark and the tone it has.
type Glyph struct {
	Kind GlyphKind
	Tone songlist.Tone
}

// PlaylistRow is a link of the list.
type PlaylistRow struct {
	Name  string
	Count string
	Glyph Glyph
}

// Progress is how far a song is, when it has begun.
type Progress struct {
	Visible bool
	Percent float64
	Tone    songlist.Tone
}

// SongRow is a song of the table, with what each column says.
type SongRow struct {
	ID       int
	Index    int
	Name     string
	Artist   string
	Video    string
	Album    string
	Quality  string
	Progress Progress
	Status   string
	Tone     songlist.Tone
	Glyph    Glyph
	Ticked   bool
}

// Footer is what is kept under the log of the song: a progress bar with its words, or a note.
type Footer struct {
	// None is true when there is nothing to show.
	None bool
	// Info is the words; Percent is how far a bar would be when HasBar is true.
	Info    string
	Percent float64
	HasBar  bool
	Tone    songlist.Tone
}
