package download

import (
	"fmt"
	"strings"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/runstate"
)

// detailRows are the lines of the details pane: for the song that is selected, or for the link
// when that is what has the focus, what happened and what to do, in everyday words and not as
// the tool said it. Each is wrapped to w cells. The technical log is the other way to see it.
func (m Model) detailRows(w int) []OutputRow {
	var rows []OutputRow
	say := func(text string, style CellStyle) {
		for _, line := range wrapText(text, max(w-2, 8)) {
			rows = append(rows, OutputRow{Line: cells.CellText(line), LineStyle: style})
		}
	}
	blank := func() { rows = append(rows, OutputRow{Line: " "}) }
	c := look.Colours
	plain, bold, dim := CellStyle{Fg: c.Text}, CellStyle{Fg: c.Text, Bold: true}, CellStyle{Fg: c.Dim}

	if m.selPlaylist >= len(m.state.Playlists) {
		return emptyOutputRows(m.emptyOutputText(), w)
	}
	pl := m.state.Playlists[m.selPlaylist]
	en := m.selectedEntry()
	if m.focus == paneLists || en == nil {
		say(m.playlistName(pl), bold)
		if address := shortAddress(pl.URL); address != m.playlistName(pl) {
			say(address, dim)
		}
		blank()
		switch {
		case pl.Err != "":
			say("✗ "+cells.Display(runstate.PlainError(pl.Err)), CellStyle{Fg: c.Error})
			blank()
			say("Check the link, or remove it and add it again.", dim)
		case !pl.Listed:
			say("Reading the songs of this link…", dim)
		default:
			say(m.linkSummary(pl), plain)
			if m.waitingIn(pl) > 0 && m.anyWaiting() {
				blank()
				say("Accept them all with the button below, or tab to the songs to choose one by one.", CellStyle{Fg: c.Highlight})
			}
		}
		return rows
	}

	say(entryTitle(en), bold)
	if who := m.whoSings(en); who != "" {
		say("by "+who, dim)
	}
	blank()
	sentence, style := m.songSentence(en)
	say(sentence, style)
	if waiting(en) && en.TargetID != "" {
		say(shortAddress("https://www.youtube.com/watch?v="+en.TargetID), dim)
	}
	return rows
}

// whoSings is the artist of a song, or the channel it came from when no artist is known.
func (m Model) whoSings(en *runstate.Entry) string {
	switch {
	case en.NamedArtist != "":
		return en.NamedArtist
	case en.Channel != "":
		// A channel of the music service is named after the artist, with this added.
		return strings.TrimSuffix(en.Channel, " - Topic")
	}
	return ""
}

// linkSummary says how the songs of a link are doing, counting only what is not zero.
func (m Model) linkSummary(pl *runstate.Playlist) string {
	var saved, failed, waiting, working, had, toChoose, cancelled, accepted int
	for _, id := range pl.Entries {
		switch m.state.Entries[id].State {
		case engine.StateDone:
			saved++
		case engine.StateFailed:
			failed++
		case engine.StateQueued:
			waiting++
		case engine.StateDuplicate:
			had++
		case engine.StatePlanned:
			toChoose++
		case engine.StateAccepted:
			accepted++
		case engine.StateSkipped:
			cancelled++
		case engine.StateResolving, engine.StateDownloading, engine.StateMerging:
			working++
		}
	}
	parts := []string{counted(len(pl.Entries), "song")}
	for _, p := range []struct {
		n    int
		text string
	}{{saved, "saved"}, {working, "being downloaded"}, {toChoose, "waiting for you to choose"}, {accepted, "accepted"}, {waiting, "waiting"}, {had, "already in your folder"}, {cancelled, "rejected"}, {failed, "failed"}} {
		if p.n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", p.n, p.text))
		}
	}
	return strings.Join(parts, " · ")
}

// songSentence says what happened to a song, and what to do when it went wrong.
func (m Model) songSentence(en *runstate.Entry) (string, CellStyle) {
	c := look.Colours
	switch en.State {
	case engine.StateQueued:
		return "Waiting for its turn.", CellStyle{Fg: c.Dim}
	case engine.StateResolving:
		return "Looking for the best version of this song…", CellStyle{Fg: c.Focus}
	case engine.StateDownloading:
		return "Downloading: " + progressText(en), CellStyle{Fg: c.Highlight}
	case engine.StateMerging:
		return "Putting the picture and the sound together…", CellStyle{Fg: c.Highlight}
	case engine.StateDone:
		switch {
		case en.Official:
			return "✓ Saved in your videos folder. It was replaced with the official music video. Not the right one? Swap the video.", CellStyle{Fg: c.Success}
		case en.Better:
			return "✓ Saved in your videos folder. A better quality version was found and used.", CellStyle{Fg: c.Success}
		}
		return "✓ Saved in your videos folder. Not the right video? Swap it.", CellStyle{Fg: c.Success}
	case engine.StateDuplicate:
		return "Skipped: you already have this song.", CellStyle{Fg: c.Dim}
	case engine.StateNotFound:
		return "Skipped: no official video was found for it.", CellStyle{Fg: c.Dim}
	case engine.StateFailed:
		return "✗ " + cells.Display(runstate.PlainError(en.Err)) + " Press r to try again.", CellStyle{Fg: c.Error}
	case engine.StateAccepted:
		if m.paused {
			return "Accepted. It downloads when you press Download.", CellStyle{Fg: c.Focus}
		}

		return "Accepted: it downloads when its turn comes.", CellStyle{Fg: c.Focus}
	case engine.StatePlanned:
		what, style := "No official video was found: the video from your playlist is what you would get.", CellStyle{Fg: c.Highlight}
		switch {
		case en.Official:
			what, style = "Found the official music video.", CellStyle{Fg: c.Success}
		case en.OwnOfficial:
			what, style = "The video in your playlist is the official one.", CellStyle{Fg: c.Success}
		case en.Better:
			what, style = "Found a better quality version of it.", CellStyle{Fg: c.Focus}
		}
		return what + " Accept it to put it in the downloads, or reject it.", style
	case engine.StateSkipped:
		return "Rejected: this song will not be downloaded. Accept it to bring it back.", CellStyle{Fg: c.Dim}
	}
	return "", CellStyle{Fg: c.Text}
}

// progressText is how far a download is, said briefly: "42% of 12.1MiB · 8.1MiB/s · 0:09 left".
func progressText(en *runstate.Entry) string {
	if en.State == engine.StateMerging {
		return "joining the picture and the sound"
	}
	var parts []string
	switch {
	case en.Total > 0:
		parts = append(parts, fmt.Sprintf("%.0f%% of %s", en.Percent, runstate.HumanBytes(en.Total)))
	default:
		parts = append(parts, "starting")
	}
	if en.Speed > 0 {
		parts = append(parts, runstate.HumanSpeed(en.Speed))
	}
	if en.ETA > 0 {
		parts = append(parts, runstate.HumanETA(en.ETA)+" left")
	}
	return strings.Join(parts, " · ")
}

// playlistName is what a link is called: its title when it has one; for a single video the
// title of the song; and until it is read, its address.
func (m Model) playlistName(pl *runstate.Playlist) string {
	if pl.Title != "" && pl.Title != pl.URL {
		return pl.Title
	}
	if pl.Listed && len(pl.Entries) == 1 {
		if name := entryTitle(m.state.Entries[pl.Entries[0]]); name != "" {
			return name
		}
	}
	return shortAddress(pl.URL)
}

// shortAddress is a web address without what a person does not read: the scheme and "www.".
func shortAddress(address string) string {
	for _, prefix := range []string{"https://", "http://"} {
		address = strings.TrimPrefix(address, prefix)
	}
	return strings.TrimPrefix(address, "www.")
}
