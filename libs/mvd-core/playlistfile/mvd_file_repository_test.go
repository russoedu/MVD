package playlistfile

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"youtube-downloader/libs/mvd-core/ytdlp"
)

func sample() File {
	return File{
		Created: time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC),
		Items: []Item{{Source: "https://youtube.com/playlist?list=M", Songs: []Entry{
			{
				Playlist: "Mine", PlaylistURL: "https://youtube.com/playlist?list=M",
				Upload:   ytdlp.PlaylistEntry{ID: "aaaaaaaaaa1", Title: "Café ☕ (Live)", Channel: "Artist - Topic", PlaylistIndex: 1},
				TargetID: "OFFICIAL001", Kind: "official", Reason: "linked from the description", Artist: "Artist", Title: "Café",
				Decision: DecisionReplaced, ChosenID: "CHOSEN00001", ChosenTitle: "Café (official)", Downloaded: true,
			},
			{Playlist: "Mine", PlaylistURL: "https://youtube.com/playlist?list=M", Upload: ytdlp.PlaylistEntry{ID: "bbbbbbbbbb1", Title: "Two", PlaylistIndex: 2}, TargetID: "bbbbbbbbbb1", Kind: "original", State: StateQueued, Error: "slow"},
		}}},
	}
}

func TestAFileReadsBackAsItWasWritten(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, sample()); err != nil {
		t.Fatal(err)
	}
	got, err := Read(&buf)
	if err != nil {
		t.Fatal(err)
	}
	want := sample()
	if !got.Created.Equal(want.Created) || len(got.Items) != 1 || got.Items[0].Source != want.Items[0].Source || len(got.Songs()) != 2 {
		t.Fatalf("got %+v", got)
	}
	if got.Songs()[0] != want.Songs()[0] || got.Songs()[1] != want.Songs()[1] {
		t.Errorf("songs differ:\n got %+v\nwant %+v", got.Songs(), want.Songs())
	}
}

func TestAFileIsCompressedJSONLines(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, sample()); err != nil {
		t.Fatal(err)
	}
	if b := buf.Bytes(); len(b) < 2 || b[0] != 0x1f || b[1] != 0x8b {
		t.Fatalf("not gzip: % x", b[:min(4, len(b))])
	}
	zr, err := gzip.NewReader(&buf)
	if err != nil {
		t.Fatal(err)
	}
	var plain bytes.Buffer
	_, _ = plain.ReadFrom(zr)
	lines := strings.Split(strings.TrimSpace(plain.String()), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], `{"mvd":2`) {
		t.Errorf("want a header and one item as JSON lines, got %q", lines)
	}
}

func TestSaveAndLoadGoThroughAFileAndReplaceAnOldOne(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "list.mvd")
	if err := Save(path, sample()); err != nil {
		t.Fatal(err)
	}
	f := sample()
	f.Items[0].Songs = f.Items[0].Songs[:1]
	if err := Save(path, f); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil || len(got.Songs()) != 1 {
		t.Fatalf("%+v, %v", got, err)
	}
	if left, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".mvd-*")); len(left) != 0 {
		t.Errorf("temporary files left behind: %v", left)
	}
}

func gzipped(content string) []byte {
	var b bytes.Buffer
	zw := gzip.NewWriter(&b)
	_, _ = zw.Write([]byte(content))
	_ = zw.Close()
	return b.Bytes()
}

func TestFilesThatAreNotMVDFilesAreRefusedWithAReason(t *testing.T) {
	for name, c := range map[string]struct {
		data []byte
		want string
	}{
		"text":      {[]byte("just some text"), "not an MVD file"},
		"gzip":      {gzipped("{\"hello\":1}\n"), "not an MVD file"},
		"newer":     {gzipped("{\"mvd\":99}\n"), "another version"},
		"older":     {gzipped("{\"mvd\":1}\n"), "another version"},
		"empty":     {nil, "not an MVD file"},
		"empty-zip": {gzipped(""), "not an MVD file"},
	} {
		_, err := Read(bytes.NewReader(c.data))
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: error %v, want one saying %q", name, err, c.want)
		}
	}
}

func TestLoadOfAMissingFileSaysSo(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "nothing.mvd")); !os.IsNotExist(err) {
		t.Errorf("error %v, want a not-exist error", err)
	}
}

func TestTheKeyOfASongIsThePlaylistAndTheUpload(t *testing.T) {
	a := Entry{PlaylistURL: "u", Upload: ytdlp.PlaylistEntry{ID: "x"}}
	b := Entry{PlaylistURL: "u", Upload: ytdlp.PlaylistEntry{ID: "x"}, Decision: DecisionConfirmed}
	c := Entry{PlaylistURL: "v", Upload: ytdlp.PlaylistEntry{ID: "x"}}
	if a.Key() != b.Key() || a.Key() == c.Key() {
		t.Errorf("keys %q %q %q", a.Key(), b.Key(), c.Key())
	}
	t1 := Entry{PlaylistURL: "u", Upload: ytdlp.PlaylistEntry{Title: "Song", Channel: "Artist", PlaylistIndex: 3}}
	t2 := Entry{PlaylistURL: "u", Upload: ytdlp.PlaylistEntry{Title: "Other", Channel: "Artist", PlaylistIndex: 4}}
	if t1.Key() == t2.Key() {
		t.Error("two songs of another service share a key")
	}
}

func TestFromSongsMakesOneItemPerPlaylistInOrder(t *testing.T) {
	songs := []Entry{
		{PlaylistURL: "a", Upload: ytdlp.PlaylistEntry{ID: "1"}},
		{PlaylistURL: "b", Upload: ytdlp.PlaylistEntry{ID: "2"}},
		{PlaylistURL: "a", Upload: ytdlp.PlaylistEntry{ID: "3"}},
	}
	f := FromSongs(songs)
	if got := f.Sources(); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("sources %v", got)
	}
	if len(f.Items[0].Songs) != 2 || f.Songs()[1].Upload.ID != "3" {
		t.Errorf("songs %+v", f.Songs())
	}
}
