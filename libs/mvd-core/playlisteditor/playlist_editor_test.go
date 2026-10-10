package playlisteditor

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/ytdlp"
)

func row(id, target, kind, decision string) playlistfile.Entry {
	return playlistfile.Entry{
		Playlist: "Mine", PlaylistURL: "https://youtube.com/playlist?list=M",
		Upload:   ytdlp.PlaylistEntry{ID: id, Title: "Song " + id},
		TargetID: target, Kind: kind, Decision: decision,
	}
}

func TestTheTargetOfARowFollowsTheDecision(t *testing.T) {
	cases := []struct {
		name     string
		row      playlistfile.Entry
		wantID   string
		wantKind engine.PlanKind
		wantOK   bool
	}{
		{"proposal", row("up1", "OFF1", "official", ""), "OFF1", engine.KindOfficial, true},
		{"confirmed", row("up1", "OFF1", "official", playlistfile.DecisionConfirmed), "OFF1", engine.KindOfficial, true},
		{"original", row("up1", "OFF1", "official", playlistfile.DecisionOriginal), "up1", engine.KindOriginal, true},
		{"skipped", row("up1", "OFF1", "official", playlistfile.DecisionSkipped), "", "", false},
		{"nothing proposed", row("up1", "up1", "none", ""), "", "", false},
	}
	for _, c := range cases {
		id, kind, ok := Target(c.row)
		if id != c.wantID || kind != c.wantKind || ok != c.wantOK {
			t.Errorf("%s: got %q %q %v, want %q %q %v", c.name, id, kind, ok, c.wantID, c.wantKind, c.wantOK)
		}
	}

	replaced := row("up1", "OFF1", "official", playlistfile.DecisionReplaced)
	replaced.ChosenID = "MINE0000001"
	if id, kind, ok := Target(replaced); id != "MINE0000001" || kind != engine.KindChosen || !ok {
		t.Errorf("replaced: %q %q %v", id, kind, ok)
	}
	// A replacement with no video named has nothing to download.
	replaced.ChosenID = ""
	if _, _, ok := Target(replaced); ok {
		t.Error("a replacement with no video is taken")
	}

	// A person can take the original of a song no video was found for.
	none := row("up1", "up1", "none", playlistfile.DecisionOriginal)
	if id, kind, ok := Target(none); id != "up1" || kind != engine.KindOriginal || !ok {
		t.Errorf("original of an unfound song: %q %q %v", id, kind, ok)
	}
}

func rowsToDownload() []playlistfile.Entry {
	replaced := row("up3", "OFF3", "official", playlistfile.DecisionReplaced)
	replaced.ChosenID = "MINE0000001"
	done := row("up5", "OFF5", "official", playlistfile.DecisionConfirmed)
	done.Downloaded = true
	return []playlistfile.Entry{
		row("up1", "OFF1", "official", playlistfile.DecisionConfirmed),
		row("up2", "OFF2", "better", ""), // not looked at
		replaced,
		row("up4", "OFF4", "official", playlistfile.DecisionSkipped),
		done,
		row("up6", "up6", "none", ""), // nothing found
	}
}

func TestADownloadTakesTheRevisedSongsAndLeavesOutTheOthers(t *testing.T) {
	planned, keys := Selection(rowsToDownload(), false)
	if len(planned) != 2 || len(keys) != 2 {
		t.Fatalf("planned %+v", planned)
	}
	if planned[0].TargetID != "OFF1" || planned[0].Kind != engine.KindOfficial {
		t.Errorf("the confirmed song: %+v", planned[0])
	}
	if planned[1].TargetID != "MINE0000001" || planned[1].Kind != engine.KindChosen || planned[1].Upload.ID != "up3" {
		t.Errorf("the replaced song: %+v", planned[1])
	}
	if planned[0].Playlist != "Mine" || planned[0].PlaylistURL == "" {
		t.Errorf("the playlist is lost: %+v", planned[0])
	}
}

func TestADownloadCanTakeTheUnreviewedSongsAsProposedToo(t *testing.T) {
	planned, _ := Selection(rowsToDownload(), true)
	if len(planned) != 3 || planned[1].Upload.ID != "up2" || planned[1].Kind != engine.KindBetter {
		t.Fatalf("planned %+v", planned)
	}
}

func TestOnlyTheSongsThatWouldDownloadAndWereNotLookedAtAreCountedAsUnreviewed(t *testing.T) {
	if n := Unreviewed(rowsToDownload()); n != 1 {
		t.Errorf("unreviewed = %d, want 1 (up2): a skipped, a done and an unfound song do not count", n)
	}
}

func TestDownloadedSongsAreMarkedByKeyAndSkippedTheNextTime(t *testing.T) {
	rows := rowsToDownload()
	_, keys := Selection(rows, false)
	if n := MarkDownloaded(rows, keys...); n != 2 {
		t.Fatalf("marked %d, want 2", n)
	}
	if planned, _ := Selection(rows, true); len(planned) != 1 || planned[0].Upload.ID != "up2" {
		t.Errorf("after the download only the song nobody looked at is left: %+v", planned)
	}
}

func TestReopeningAPlanKeepsTheDecisionsAndWhatWasDownloaded(t *testing.T) {
	saved := []playlistfile.Entry{row("up1", "OFF1", "official", playlistfile.DecisionConfirmed), row("gone", "G", "official", playlistfile.DecisionSkipped)}
	saved[0].Downloaded = true
	fresh := []playlistfile.Entry{row("up1", "OFF9", "better", ""), row("up2", "OFF2", "official", "")}

	got := Merge(fresh, saved)
	if len(got) != 3 {
		t.Fatalf("got %+v", got)
	}
	// The new proposal is what the lookup says now; the person's decision is theirs.
	if got[0].TargetID != "OFF9" || got[0].Decision != playlistfile.DecisionConfirmed || !got[0].Downloaded {
		t.Errorf("merged row: %+v", got[0])
	}
	if got[1].Upload.ID != "up2" || got[1].Decision != "" {
		t.Errorf("new row: %+v", got[1])
	}
	if got[2].Upload.ID != "gone" || got[2].Decision != playlistfile.DecisionSkipped {
		t.Errorf("a song only the file had is dropped: %+v", got[2])
	}
}

func TestFromPlanMakesRowsNobodyHasDecidedOn(t *testing.T) {
	rows := FromPlan([]engine.PlannedEntry{{Playlist: "P", PlaylistURL: "u", Upload: ytdlp.PlaylistEntry{ID: "x"}, TargetID: "Y", Kind: engine.KindOfficial, Reason: "r", Artist: "A", Title: "T"}})
	if len(rows) != 1 || rows[0].TargetID != "Y" || rows[0].Kind != "official" || rows[0].Decision != "" || rows[0].Artist != "A" {
		t.Errorf("rows %+v", rows)
	}
}

func TestVideoAddresses(t *testing.T) {
	const id = "dQw4w9WgXcQ"
	for _, in := range []string{
		id,
		"https://www.youtube.com/watch?v=" + id,
		"https://www.youtube.com/watch?v=" + id + "&list=PLabc&t=10s",
		"youtube.com/watch?v=" + id,
		"https://youtu.be/" + id + "?si=abc",
		"https://music.youtube.com/watch?v=" + id + "&si=xyz",
		"https://www.youtube.com/shorts/" + id,
		"https://www.youtube.com/embed/" + id,
		"  https://m.youtube.com/watch?v=" + id + "  ",
	} {
		if got, ok := VideoID(in); !ok || got != id {
			t.Errorf("VideoID(%q) = %q, %v", in, got, ok)
		}
	}
	for _, in := range []string{"", "hello", "https://example.com/watch?v=" + id, "https://www.youtube.com/playlist?list=PLabc", "https://www.youtube.com/watch?v=short"} {
		if got, ok := VideoID(in); ok {
			t.Errorf("VideoID(%q) = %q, want no video", in, got)
		}
	}
	if VideoAddress(id) != "https://www.youtube.com/watch?v="+id {
		t.Error("address")
	}
}

func TestADecisionIsLoggedAsOneJSONLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "decisions.jsonl")
	r := row("up1", "OFF1", "official", playlistfile.DecisionReplaced)
	r.ChosenID, r.Reason = "MINE0000001", "linked from the description"
	for _, note := range []string{"the official one is a live version", ""} {
		if err := AppendDecision(path, RecordFor(r, note)); err != nil {
			t.Fatal(err)
		}
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	var recs []DecisionRecord
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		var rec DecisionRecord
		if err := json.Unmarshal(sc.Bytes(), &rec); err != nil {
			t.Fatal(err)
		}
		recs = append(recs, rec)
	}
	if len(recs) != 2 || recs[0].Note == "" || recs[0].Proposed != "OFF1" || recs[0].Chosen != "MINE0000001" || recs[0].Decision != "replaced" || recs[0].Reason == "" {
		t.Errorf("records %+v", recs)
	}
}
