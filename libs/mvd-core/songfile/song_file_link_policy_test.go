package songfile

import "testing"

func TestFilePath(t *testing.T) {
	for link, want := range map[string]string{
		`C:\Users\me\songs.txt`:      `C:\Users\me\songs.txt`,
		`"C:\Users\me\My Songs.CSV"`: `C:\Users\me\My Songs.CSV`,
		`/home/me/list.tsv`:          `/home/me/list.tsv`,
		`  songs.txt  `:              `songs.txt`,
	} {
		if got, ok := FilePath(link); !ok || got != want {
			t.Errorf("FilePath(%q) = %q, %v; want %q", link, got, ok, want)
		}
	}
	for _, link := range []string{"", "https://example.com/songs.txt", "https://youtube.com/playlist?list=abc", `C:\music\song.mp3`, "songs"} {
		if got, ok := FilePath(link); ok {
			t.Errorf("FilePath(%q) = %q; not a song file", link, got)
		}
	}
}
