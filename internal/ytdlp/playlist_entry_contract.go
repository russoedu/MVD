// Package ytdlp runs the yt-dlp executable: it lists playlists, downloads
// single videos with captured output and decodes what yt-dlp prints.
package ytdlp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// PlaylistEntry is one line of `yt-dlp --flat-playlist -j <playlist>`.
type PlaylistEntry struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Channel       string `json:"channel"`
	Uploader      string `json:"uploader"`
	URL           string `json:"url"`
	Playlist      string `json:"playlist"`
	PlaylistTitle string `json:"playlist_title"`
	PlaylistID    string `json:"playlist_id"`
	PlaylistIndex int    `json:"playlist_index"`
	PlaylistCount int    `json:"playlist_count"`
}

// ParsePlaylistEntries decodes the JSON lines of a flat playlist listing.
func ParsePlaylistEntries(r io.Reader) ([]PlaylistEntry, error) {
	var entries []PlaylistEntry
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 16*1024*1024)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var e PlaylistEntry
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("cannot parse yt-dlp playlist entry: %w", err)
		}
		if e.ID == "" {
			continue
		}
		entries = append(entries, e)
	}
	return entries, scanner.Err()
}
