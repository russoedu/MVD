package ytdlp

import (
	"encoding/base64"
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
	isFlat, dump := false, false
	cookiesFile, browser := "", ""
	for i, a := range args {
		switch a {
		case "--flat-playlist":
			isFlat = true
		case "--dump-pages":
			dump = true
		case "--cookies":
			cookiesFile = args[i+1]
		case "--cookies-from-browser":
			browser = args[i+1]
		}
	}
	url := args[len(args)-1]

	if browser != "" {
		if browser == "nope" {
			fmt.Fprintln(os.Stderr, "ERROR: could not find nope cookies database")
			return 1
		}
		os.WriteFile(cookiesFile, []byte("# Netscape HTTP Cookie File\n.youtube.com\tTRUE\t/\tTRUE\t0\tSID\tsecret\n"), 0600)
		return 0
	}

	if dump {
		id := url[strings.LastIndex(url, "=")+1:]
		page := `<script>var ytInitialData = {"engagementPanels":[{"engagementPanelSectionListRenderer":{"panelIdentifier":"engagement-panel-structured-description","content":{"videoDescriptionMusicSectionRenderer":{"carouselLockups":[{"carouselLockupRenderer":{"videoLockup":{"compactVideoRenderer":{"videoId":"DUMPEDOFF01"}}}}]}}}}]};</script>`
		fmt.Println("[youtube] " + id + ": Downloading webpage")
		fmt.Println("[youtube] Dumping request to https://www.youtube.com/watch?v=" + id)
		fmt.Println(base64.StdEncoding.EncodeToString([]byte(page)))
		fmt.Println("[youtube] Dumping request to https://www.youtube.com/youtubei/v1/player?prettyPrint=false")
		fmt.Println(base64.StdEncoding.EncodeToString([]byte(`{"videoDetails":{}}`)))
		return 0
	}

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
