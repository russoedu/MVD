package ytdlp

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestMain lets the test binary double as a yt-dlp stub, so tests run the
// same on every platform without shell scripts.
func TestMain(m *testing.M) {
	if os.Getenv("MVD_STUB_YTDLP") == "1" {
		os.Exit(stubYtDlp(os.Args[1:]))
	}
	os.Exit(m.Run())
}

// stubYtDlp mimics the two yt-dlp invocations the app makes.
func stubYtDlp(args []string) int {
	isFlat := false
	for _, a := range args {
		if a == "--flat-playlist" {
			isFlat = true
		}
	}
	url := args[len(args)-1]

	if isFlat {
		if !strings.Contains(url, "list=A") {
			fmt.Fprintln(os.Stderr, "ERROR: [youtube:tab] Unable to recognize playlist")
			return 1
		}
		for i, e := range []PlaylistEntry{
			{ID: "aaaaaaaaaa1", Title: "Track One", Channel: "Artist - Topic"},
			{ID: "failfailfai", Title: "Broken", Channel: "Someone"},
		} {
			e.Playlist, e.PlaylistTitle, e.PlaylistID, e.PlaylistIndex, e.PlaylistCount = "Playlist A", "Playlist A", "A", i+1, 2
			b, _ := json.Marshal(e)
			fmt.Println(string(b))
		}
		return 0
	}

	id := url[strings.LastIndex(url, "=")+1:]
	fmt.Printf("[youtube] %s: Downloading webpage\n", id)
	if strings.Contains(id, "fail") {
		fmt.Fprintln(os.Stderr, "ERROR: [youtube] "+id+": Video unavailable")
		return 1
	}
	fmt.Println("MVD|512|1024|NA|2048.0|1")
	fmt.Printf("[Merger] Merging formats into \"%s.mp4\"\n", id)
	return 0
}

func useStub(t *testing.T) string {
	t.Helper()
	os.Setenv("MVD_STUB_YTDLP", "1")
	t.Cleanup(func() { os.Unsetenv("MVD_STUB_YTDLP") })
	return os.Args[0]
}
