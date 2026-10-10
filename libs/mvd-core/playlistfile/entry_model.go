package playlistfile

import (
	"strconv"

	"youtube-downloader/libs/mvd-core/ytdlp"
)

// Decisions a person can make about a song.
const (
	// DecisionNone is a song nobody has looked at.
	DecisionNone = ""
	// DecisionConfirmed accepts what was proposed.
	DecisionConfirmed = "confirmed"
	// DecisionReplaced takes another video, ChosenID, in place of the proposal.
	DecisionReplaced = "replaced"
	// DecisionOriginal takes the upload that is in the playlist.
	DecisionOriginal = "original"
	// DecisionSkipped leaves the song out.
	DecisionSkipped = "skipped"
)

// Entry is one song of a plan file.
type Entry struct {
	Playlist    string              `json:"playlist"`
	PlaylistURL string              `json:"playlist_url"`
	Upload      ytdlp.PlaylistEntry `json:"upload"`

	// What was proposed: the video, its kind (official, own-official, better, original,
	// none), why, and the song as a music database named it.
	TargetID string `json:"target"`
	Kind     string `json:"kind"`
	Reason   string `json:"reason,omitempty"`
	Artist   string `json:"artist,omitempty"`
	Title    string `json:"title,omitempty"`

	// What the person decided: see the Decision constants. ChosenID is the video taken
	// when the proposal was replaced, and ChosenTitle what it is called.
	Decision    string `json:"decision,omitempty"`
	ChosenID    string `json:"chosen,omitempty"`
	ChosenTitle string `json:"chosen_title,omitempty"`

	// Downloaded is true once the song has been downloaded, so the next time it is skipped.
	Downloaded bool `json:"downloaded,omitempty"`
}

// Key identifies a song across files and runs: the playlist and the upload in it.
func (e Entry) Key() string {
	id := e.Upload.ID
	if id == "" {
		// A song of another service has no upload: its place in the playlist and its name do.
		id = strconv.Itoa(e.Upload.PlaylistIndex) + ":" + e.Upload.Title + ":" + e.Upload.Channel
	}
	return e.PlaylistURL + "|" + id
}
