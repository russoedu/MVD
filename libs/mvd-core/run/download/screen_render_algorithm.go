package download

import (
	"youtube-downloader/libs/mvd-core/run/downloadsnapshot"
	"fmt"
	"time"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/run/loading"
	helppage "youtube-downloader/libs/mvd-core/run/help"
	"youtube-downloader/libs/mvd-core/screenview"

	"youtube-downloader/libs/mvd-core/runstate"
)

func (m Model) render() string {
	if m.width == 0 || m.height == 0 {
		return "starting..."
	}
	if m.width < 60 || m.height < 16 {
		return look.Dim.Render(fmt.Sprintf("terminal too small (%dx%d), need at least 60x16", m.width, m.height))
	}

	if !m.session.Loaded && !m.session.Hosted && !m.session.Adding && !m.session.ShowHelp {
		return m.renderLoading()
	}

	view := screenview.Padded(View(m.downloadProps(m.width, m.height), m.width, m.height).View, m.width)
	view = m.withTableControls(m.withScrollbars(m.withChoiceButtons(view)))

	if m.session.Adding {
		return m.renderAddBox()
	}
	if m.session.ShowHelp {
		return m.renderHelp()
	}
	return view
}

// renderLoading is the screen shown until every playlist of the run has been listed.
func (m Model) renderLoading() string {
	p := loading.Progress{Spin: m.spin, Detail: "Reading the songs of each playlist", Elapsed: runstate.HumanDuration(time.Since(m.session.State.Started))}
	for _, pl := range m.session.State.Playlists {
		p.Playlists = append(p.Playlists, loading.Playlist{
			Name: m.playlistName(pl), Songs: len(pl.Entries), Listed: pl.Listed, Failed: pl.Err != "",
		})
	}
	if m.session.ConfirmQuit {
		p.Confirm = look.Red.Render(" Downloads are still running. Quit? ") + look.Key.Render("y") + look.Dim.Render("/") + look.Key.Render("n")
	}
	return loading.Render(p, m.width, m.height)
}

func (m Model) outputTitle() string { return cells.Display(downloadsnapshot.OutputTitle(m.session)) }

// hints lists the keys the download screen offers now. The key bar and the
// outline both read it, so they cannot drift apart.
func (m Model) hints() []keybar.Hint {
	var hints []keybar.Hint
	for _, h := range downloadsnapshot.HintsOf(m.session) {
		hints = append(hints, keybar.Hint{Key: h.Key, Desc: h.Desc})
	}

	return hints
}

func (m Model) renderHelp() string {
	logPath := ""
	if m.session.Ctrl != nil {
		logPath = m.session.Ctrl.LogPath()
	}
	return helppage.Screen(m.width, m.height, helppage.Options{LogPath: logPath, Hosted: m.session.Hosted})
}

// counted writes a count with its noun: "1 file", "4 files".
func counted(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}

// glyphOf is the character and the colour a mark is drawn in.
func (m Model) glyphOf(g downloadsnapshot.Glyph) (glyph, colour string) {
	colour = toneColour(g.Tone)
	switch g.Kind {
	case downloadsnapshot.GlyphBusy:
		return look.SpinnerFrames[m.spin], colour
	case downloadsnapshot.GlyphFailed:
		return "✗", colour
	case downloadsnapshot.GlyphDone:
		return "✓", colour
	case downloadsnapshot.GlyphChoose:
		return "?", colour
	case downloadsnapshot.GlyphIdle:
		return "·", colour
	case downloadsnapshot.GlyphResolving:
		return "⟲", colour
	case downloadsnapshot.GlyphMerging:
		return "⚙", colour
	case downloadsnapshot.GlyphHad:
		return "≡", colour
	case downloadsnapshot.GlyphMissing:
		return "∅", colour
	case downloadsnapshot.GlyphAccepted:
		return "▸", colour
	case downloadsnapshot.GlyphRejected:
		return "–", colour
	}

	return " ", colour
}

// entryGlyph is the mark at the start of an entry row and its colour.
func (m Model) entryGlyph(en *runstate.Entry) (glyph, colour string) {
	return m.glyphOf(downloadsnapshot.SongGlyph(en))
}
