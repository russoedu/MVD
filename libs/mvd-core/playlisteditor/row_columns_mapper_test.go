package playlisteditor

import (
	"testing"

	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/ytdlp"
)

func entry(title, channel string) playlistfile.Entry {
	return playlistfile.Entry{Upload: ytdlp.PlaylistEntry{ID: "id1", Title: title, Channel: channel}}
}

func TestColumnsKeepWhatADatabaseNamed(t *testing.T) {
	r := entry("Another Night (Official Audio)", "Real McCoy - Topic")
	r.Artist, r.Title = "Real McCoy", "Another Night"
	got := ColumnsOf(r)
	if got.Artist != "Real McCoy" || got.Song != "Another Night" || got.Video != "Another Night (Official Audio)" {
		t.Errorf("got %+v", got)
	}
}

func TestATitleWithAnArtistIsSplit(t *testing.T) {
	got := ColumnsOf(entry("Double You - Part-Time Lover", "Some Channel"))
	if got.Artist != "Double You" || got.Song != "Part-Time Lover" {
		t.Errorf("got %+v", got)
	}
}

func TestATitleWithoutAnArtistTakesTheChannelWithoutTopic(t *testing.T) {
	got := ColumnsOf(entry("Mr. Vain (Original Radio Edit)", "Culture Beat - Topic"))
	if got.Artist != "Culture Beat" || got.Song != "Mr. Vain (Original Radio Edit)" {
		t.Errorf("got %+v", got)
	}
}

func TestAReplacedSongShowsTheTitleOfTheChosenVideo(t *testing.T) {
	r := entry("Le Click - Tonight Is the Night", "x")
	r.Decision, r.ChosenID, r.ChosenTitle = playlistfile.DecisionReplaced, "abc", "Le Click - Tonight (Ritsatv Remastered)"
	if got := ColumnsOf(r); got.Replaced != "Le Click - Tonight (Ritsatv Remastered)" {
		t.Errorf("got %+v", got)
	}
	r.ChosenTitle = ""
	if got := ColumnsOf(r); got.Replaced != "(abc)" {
		t.Errorf("got %+v", got)
	}
}
