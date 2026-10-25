package downloadsnapshot

import (
	"fmt"
	"sort"
	"strings"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/run/downloadsession"
	"youtube-downloader/libs/mvd-core/runstate"
	"youtube-downloader/libs/mvd-core/songlist"
	"youtube-downloader/libs/mvd-core/wording"
)

// DetailsOf are the lines of the details: for the song that is chosen, or for the link when that is
// what the person is in, what happened and what to do, in everyday words and not as the tool said it.
func DetailsOf(s downloadsession.Session) []Line {
	var lines []Line
	say := func(text string, tone songlist.Tone, strong bool) {
		lines = append(lines, Line{Text: text, Tone: tone, Strong: strong})
	}
	blank := func() { lines = append(lines, Line{Blank: true}) }

	if s.SelPlaylist >= len(s.State.Playlists) {
		return []Line{{Text: EmptyOutputText(s), Tone: songlist.ToneDim}}
	}
	pl := s.State.Playlists[s.SelPlaylist]
	en := s.SelectedEntry()
	picks := len(s.Picked)
	if picks > 1 {
		say(wording.Counted(picks, "song")+" selected", songlist.ToneText, true)
		blank()
		counts := map[string]int{}
		for id := range s.Picked {
			if p := s.State.Entry(id); p != nil {
				word, _ := songlist.StatusWord(p)
				counts[word]++
			}
		}
		words := make([]string, 0, len(counts))
		for word := range counts {
			words = append(words, word)
		}
		sort.Strings(words)
		for _, word := range words {
			say(fmt.Sprintf("%d %s", counts[word], word), songlist.ToneText, false)
		}
		blank()
		say("Accept them or reject them with the buttons below.", songlist.ToneHighlight, false)

		return lines
	}
	if picks == 1 {
		for id := range s.Picked {
			if p := s.State.Entry(id); p != nil {
				en = p // the one song that is ticked is the one the details are about
			}
		}
	}
	if (s.Pane == downloadsession.PaneLists && picks == 0) || en == nil {
		say(PlaylistName(s, pl), songlist.ToneText, true)
		if address := shortAddress(pl.URL); address != PlaylistName(s, pl) {
			say(address, songlist.ToneDim, false)
		}
		blank()
		switch {
		case pl.Err != "":
			say("✗ "+runstate.PlainError(pl.Err), songlist.ToneError, false)
			blank()
			say("Check the link, or remove it and add it again.", songlist.ToneDim, false)
		case !pl.Listed:
			say("Reading the songs of this link…", songlist.ToneDim, false)
		default:
			say(linkSummary(s, pl), songlist.ToneText, false)
			if s.WaitingIn(pl) > 0 && s.AnyWaiting() {
				blank()
				say("Accept them all with the button below, or tab to the songs to choose one by one.", songlist.ToneHighlight, false)
			}
		}

		return lines
	}

	lines = append(lines, Line{Text: songlist.VideoTitle(en), Tone: songlist.ToneText, Strong: true, SongTitle: true})
	if who := songlist.Artist(en); who != "" {
		say("by "+who, songlist.ToneDim, false)
	}
	blank()
	sentence, tone := SongSentence(s, en)
	say(sentence, tone, false)
	if downloadsession.Waiting(en) && en.TargetID != "" {
		say(shortAddress("https://www.youtube.com/watch?v="+en.TargetID), songlist.ToneDim, false)
	}

	return lines
}

// linkSummary says how the songs of a link are doing, counting only what is not zero.
func linkSummary(s downloadsession.Session, pl *runstate.Playlist) string {
	var saved, failed, waiting, working, had, toChoose, cancelled, accepted int
	for _, id := range pl.Entries {
		switch s.State.Entries[id].State {
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
	parts := []string{wording.Counted(len(pl.Entries), "song")}
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

// SongSentence says what happened to a song, and what to do when it went wrong.
func SongSentence(s downloadsession.Session, en *runstate.Entry) (string, songlist.Tone) {
	switch en.State {
	case engine.StateQueued:
		return "Waiting for its turn.", songlist.ToneDim
	case engine.StateResolving:
		return "Looking for the best version of this song…", songlist.ToneFocus
	case engine.StateDownloading:
		return "Downloading: " + progressText(en), songlist.ToneHighlight
	case engine.StateMerging:
		return "Putting the picture and the sound together…", songlist.ToneHighlight
	case engine.StateDone:
		switch {
		case en.Official:
			return "✓ Saved in your videos folder. It was replaced with the official music video. Not the right one? Swap the video.", songlist.ToneSuccess
		case en.Better:
			return "✓ Saved in your videos folder. A better quality version was found and used.", songlist.ToneSuccess
		}

		return "✓ Saved in your videos folder. Not the right video? Swap it.", songlist.ToneSuccess
	case engine.StateDuplicate:
		return "Skipped: you already have this song.", songlist.ToneDim
	case engine.StateNotFound:
		return "Skipped: no official video was found for it.", songlist.ToneDim
	case engine.StateFailed:
		return "✗ " + runstate.PlainError(en.Err) + " Press r to try again.", songlist.ToneError
	case engine.StateAccepted:
		if s.Paused {
			return "Accepted. It downloads when you press Download.", songlist.ToneFocus
		}

		return "Accepted: it downloads when its turn comes.", songlist.ToneFocus
	case engine.StatePlanned:
		what, tone := "No official video was found: the video from your playlist is what you would get.", songlist.ToneHighlight
		switch {
		case en.Official:
			what, tone = "Found the official music video.", songlist.ToneSuccess
		case en.OwnOfficial:
			what, tone = "The video in your playlist is the official one.", songlist.ToneSuccess
		case en.Better:
			what, tone = "Found a better quality version of it.", songlist.ToneFocus
		}

		return what + " Accept it to put it in the downloads, or reject it.", tone
	case engine.StateSkipped:
		return "Rejected: this song will not be downloaded. Accept it to bring it back.", songlist.ToneDim
	}

	return "", songlist.ToneText
}

// progressText is how far a download is, said briefly: "42% of 12.1MiB · 8.1MiB/s · 0:09 left".
func progressText(en *runstate.Entry) string {
	if en.State == engine.StateMerging {
		return "joining the picture and the sound"
	}
	var parts []string
	if en.Total > 0 {
		parts = append(parts, fmt.Sprintf("%.0f%% of %s", en.Percent, runstate.HumanBytes(en.Total)))
	} else {
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

// PlaylistName is what a link is called: its title when it has one; for a single video the title of
// the song; and until it is read, its address.
func PlaylistName(s downloadsession.Session, pl *runstate.Playlist) string {
	if pl.Title != "" && pl.Title != pl.URL {
		return pl.Title
	}
	if pl.Listed && len(pl.Entries) == 1 {
		if name := songlist.VideoTitle(s.State.Entries[pl.Entries[0]]); name != "" {
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
