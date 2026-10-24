package download

import (
	"strings"
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
)

func namesOf(cols []tableColumn) string {
	var names []string
	for _, c := range cols {
		names = append(names, c.name)
	}

	return strings.Join(names, ",")
}

func TestTheTableKeepsTheMostImportantColumnsWhenThePaneIsNarrow(t *testing.T) {
	m := sampleModel(t)
	cases := map[int]string{
		90: "mark,idx,song,artist,progress,status",
		60: "mark,idx,song,artist,progress,status",
		45: "mark,idx,song,progress,status",
		30: "mark,idx,song,status",
		18: "mark,idx,song",
	}
	for w, want := range cases {
		if got := namesOf(m.tableColumns(w)); got != want {
			t.Errorf("a pane of %d cells has the columns %s, want %s", w, got, want)
		}
	}
}

func TestEveryColumnFitsInThePaneAndTheSongTakesWhatIsLeft(t *testing.T) {
	m := sampleModel(t)
	for _, w := range []int{30, 45, 60, 90, 140} {
		total := 0
		cols := m.tableColumns(w)
		for _, c := range cols {
			total += c.width
		}
		if total+len(cols)-1 > w && w >= 30 {
			t.Errorf("at %d cells the columns need %d", w, total+len(cols)-1)
		}
	}
}

func TestTheVideoColumnAppearsOnlyWhenTheSongIsCalledSomethingElse(t *testing.T) {
	m := sampleModel(t)
	if strings.Contains(namesOf(m.tableColumns(200)), columnVideo) {
		t.Error("a video column with nothing to say is shown")
	}
	en := m.state.Entries[0]
	en.NamedTitle = "You Don't Know Me"
	if !strings.Contains(namesOf(m.tableColumns(200)), columnVideo) {
		t.Error("the name of the video is missing when it differs from the song")
	}
	texts, _ := m.songCells(en, m.tableColumns(200), false)
	if !strings.Contains(strings.Join(texts, "|"), "You Don't Know Me (feat. Duane Harden)") || texts[4] != "You Don't Know Me" {
		t.Errorf("the cells are %q", texts)
	}
}

func TestTheTableHasAHeadingForEachColumnAndTheProgressOfEachSong(t *testing.T) {
	m := sampleModel(t)
	table := m.songsTable(100)
	for _, want := range []string{"#", "Song", "Artist", "Progress", "Status"} {
		if !strings.Contains(strings.Join(table.Header, "|"), want) {
			t.Errorf("the headings %q lack %q", table.Header, want)
		}
	}
	if len(table.Header) != len(table.HeaderStyles) {
		t.Error("a heading has no style")
	}
	en := m.state.Entries[4] // downloading, 34%
	if text, _ := progressCell(en); !strings.Contains(text, "34%") || !strings.Contains(text, "███░░░░░") {
		t.Errorf("progress %q", text)
	}
	en.State = engine.StateQueued
	if text, _ := progressCell(en); text != "" {
		t.Errorf("a song that has not begun shows progress %q", text)
	}
}

func TestWithNoSongsThePaneKeepsItsText(t *testing.T) {
	m := New(&fakeController{events: make(chan interface{})})
	if table := m.songsTable(80); table.Columns != "" {
		t.Errorf("a table with no songs: %q", table.Columns)
	}
}

func TestTheAlbumAndTheQualityAppearOnceTheyAreKnown(t *testing.T) {
	m := sampleModel(t)
	if names := namesOf(m.tableColumns(200)); strings.Contains(names, columnAlbum) || strings.Contains(names, columnQuality) {
		t.Fatalf("columns with nothing to say are shown: %s", names)
	}
	en := m.state.Entries[0]
	en.NamedAlbum, en.Height = "Whenever You Need Somebody", 1080
	names := namesOf(m.tableColumns(200))
	if !strings.Contains(names, columnAlbum) || !strings.Contains(names, columnQuality) {
		t.Fatalf("the album and the quality are missing: %s", names)
	}
	texts := strings.Join(firstOf(m.songCells(en, m.tableColumns(200), false)), "|")
	if !strings.Contains(texts, "Whenever You Need Somebody") || !strings.Contains(texts, "1080p") {
		t.Errorf("the cells are %q", texts)
	}
	en.Height, en.AudioKbps = 0, 160
	if got := qualityText(en); got != "160k" {
		t.Errorf("a sound with no picture says %q", got)
	}
	// In a narrow pane the quality is kept before the artist and the album.
	if got := namesOf(m.tableColumns(60)); !strings.Contains(got, columnQuality) {
		t.Errorf("a pane of 60 cells lost the quality: %s", got)
	}
}

func firstOf(texts, _ []string) []string { return texts }
