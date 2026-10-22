// Package loading draws the screen shown between pressing start and the moment the
// playlists are known: what is being prepared, how many playlists are listed, and which.
package loading

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/look"
)

// barWidth is the width of the progress bar, in cells.
const barWidth = 40

// Playlist is one playlist being loaded.
type Playlist struct {
	Name   string
	Songs  int
	Listed bool
	Failed bool
}

// Progress is what the screen says.
type Progress struct {
	// Detail is what is happening now; empty when there is nothing more to say.
	Detail string
	// Playlists are the playlists being listed; empty while the run is being prepared,
	// when the bar is not shown.
	Playlists []Playlist
	// Spin is the frame of the spinner to show.
	Spin int
	// Elapsed is how long the run has been going, "" when it is not known.
	Elapsed string
	// Confirm is a question on the bottom line, "" for none.
	Confirm string
}

// Render draws the screen for a window of w by h cells: a panel with the progress, what has
// been found so far, and every playlist with what became of it.
func Render(p Progress, w, h int) string {
	frame := look.SpinnerFrames[p.Spin%len(look.SpinnerFrames)]
	lines := []string{look.Yellow.Render(frame) + " " + look.Text.Bold(true).Render("Loading playlists")}

	if n := len(p.Playlists); n > 0 {
		listed, failed, songs := 0, 0, 0
		for _, pl := range p.Playlists {
			if pl.Listed {
				listed++
				songs += pl.Songs
			}
			if pl.Failed {
				failed++
			}
		}
		lines = append(lines, "", bar(listed, n)+look.Dim.Render(fmt.Sprintf("  %d of %d  ·  %d%%", listed, n, 100*listed/n)))
		facts := look.Dim.Render("Songs found ") + look.Text.Render(fmt.Sprint(songs))
		if failed > 0 {
			facts += look.Dim.Render("  ·  Could not load ") + look.Red.Render(fmt.Sprint(failed))
		}
		if p.Elapsed != "" {
			facts += look.Dim.Render("  ·  Elapsed ") + look.Text.Render(p.Elapsed)
		}
		lines = append(lines, facts)
	} else if p.Elapsed != "" {
		lines = append(lines, "", look.Dim.Render("Elapsed ")+look.Text.Render(p.Elapsed))
	}
	if p.Detail != "" {
		lines = append(lines, "", look.Dim.Render(p.Detail))
	}
	// Every row the window has room for names a playlist.
	room := max(h-len(lines)-6, 0)
	lines = append(lines, rows(p.Playlists, w, room)...)

	return centred(lines, p.Confirm, w, h)
}

// bar is a progress bar with done of total filled.
func bar(done, total int) string {
	filled := 0
	if total > 0 {
		filled = barWidth * done / total
	}
	return look.Green.Render(strings.Repeat("█", filled)) + look.Dim.Render(strings.Repeat("░", barWidth-filled))
}

// rows names the playlists, the ones still loading first so the person sees what is
// being waited for, as many as there is room for; the rest are counted.
func rows(playlists []Playlist, w, room int) []string {
	if len(playlists) == 0 || room < 2 {
		return nil
	}
	out := []string{""}
	room--
	if len(playlists) > room {
		room-- // the line that counts the rest
	}
	shown := 0
	for _, pass := range []bool{false, true} {
		for _, pl := range playlists {
			if pl.Listed != pass || shown >= max(room, 1) {
				continue
			}
			out = append(out, cells.Fit(cells.Truncate(row(pl), max(10, min(w-4, 90))), max(10, min(w-4, 90))))
			shown++
		}
	}
	if more := len(playlists) - shown; more > 0 {
		out = append(out, look.Dim.Render(fmt.Sprintf("… and %d more", more)))
	}
	return out
}

func row(pl Playlist) string {
	name := cells.Display(pl.Name)
	switch {
	case pl.Failed:
		return look.Red.Render("✗ ") + look.Text.Render(name) + look.Red.Render("  could not be loaded")
	case pl.Listed:
		return look.Green.Render("✓ ") + look.Text.Render(name) + look.Dim.Render(fmt.Sprintf("  %d songs", pl.Songs))
	}
	return look.Dim.Render("· " + name)
}

// centred puts the lines in the middle of the window, each one centred on its own.
func centred(lines []string, confirm string, w, h int) string {
	top := max(0, (h-len(lines))/2)
	out := make([]string, 0, h)
	for len(out) < top {
		out = append(out, "")
	}
	for _, l := range lines {
		out = append(out, strings.Repeat(" ", max(0, (w-ansi.StringWidth(l))/2))+l)
	}
	for len(out) < h-1 {
		out = append(out, "")
	}
	out = append(out, confirm)
	if len(out) > h {
		out = out[:h]
	}
	return strings.Join(out, "\n")
}
