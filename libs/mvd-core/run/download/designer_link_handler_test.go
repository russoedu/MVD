package download

import (
	"bufio"
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/connect"
)

// stateLine reads lines from the designer's side of the connection until a state arrives.
func stateLine(t *testing.T, sc *bufio.Scanner, conn net.Conn) map[string]any {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	for sc.Scan() {
		var m map[string]any
		if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
			t.Fatal(err)
		}
		if m["type"] == "state" {
			return m
		}
	}
	t.Fatalf("no state: %v", sc.Err())
	return nil
}

func TestTheScreenPublishesItsRowsAndAppliesTheDesignersClick(t *testing.T) {
	link, err := connect.Listen("127.0.0.1:0", page)
	if err != nil {
		t.Fatal(err)
	}
	defer link.Close()
	conn, err := net.Dial("tcp", link.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	sc := bufio.NewScanner(conn)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	if _, err := conn.Write([]byte(`{"type":"hello","protocol":1,"design":"mvd"}` + "\n")); err != nil {
		t.Fatal(err)
	}

	m := sampleModel(t)
	m.link = link
	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 36})
	m = next.(Model)

	st := stateLine(t, sc, conn)
	inputs := st["inputs"].(map[string]any)
	if st["page"] != page || st["current"] != true {
		t.Fatalf("state: %v", st)
	}
	if rows := inputs["Playlists"].([]any); len(rows) != 3 {
		t.Errorf("playlists published: %d rows, want 3", len(rows))
	}
	if rows := inputs["Entries"].([]any); len(rows) != 6 {
		t.Errorf("entries published: %d rows, want all 6 of the playlist", len(rows))
	}
	if !strings.Contains(inputs["Counters"].(string), "Running 3") {
		t.Errorf("counters: %v", inputs["Counters"])
	}

	// A click on the second playlist row of the designer's drawing selects it here.
	next, _ = m.Update(designerMsg{Page: page, Name: "PickPlaylist", X: 4, Y: 1})
	m = next.(Model)
	if m.selPlaylist != 1 || m.focus != paneLists {
		t.Errorf("playlist %d focus %d after the designer's click", m.selPlaylist, m.focus)
	}
	// And the new state is what the designer is told next.
	st = stateLine(t, sc, conn)
	if rows := st["inputs"].(map[string]any)["Entries"].([]any); len(rows) != 1 {
		t.Errorf("entries of the second playlist: %d rows, want 1", len(rows))
	}
}
