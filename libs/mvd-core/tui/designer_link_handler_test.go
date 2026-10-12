package tui

import (
	"bufio"
	tea "charm.land/bubbletea/v2"
	"encoding/json"
	"net"
	"testing"
	"time"

	"youtube-downloader/libs/mvd-core/connect"
)

// designerSide is a fake designer: it reads the states the app publishes.
type designerSide struct {
	t    *testing.T
	conn net.Conn
	sc   *bufio.Scanner
}

// state reads lines until the next state of the page.
func (d *designerSide) state(page string) map[string]any {
	d.t.Helper()
	_ = d.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	for d.sc.Scan() {
		var m map[string]any
		if err := json.Unmarshal(d.sc.Bytes(), &m); err != nil {
			d.t.Fatal(err)
		}
		if m["type"] == "state" && m["page"] == page {
			return m["inputs"].(map[string]any)
		}
	}
	d.t.Fatalf("no state of %s: %v", page, d.sc.Err())
	return nil
}

func TestEveryScreenIsPublishedAndTheDesignersClickSelectsASetting(t *testing.T) {
	link, err := connect.Listen("127.0.0.1:0", connect.Pages...)
	if err != nil {
		t.Fatal(err)
	}
	defer link.Close()
	conn, err := net.Dial("tcp", link.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	d := &designerSide{t: t, conn: conn, sc: bufio.NewScanner(conn)}
	d.sc.Buffer(make([]byte, 1<<20), 1<<20)
	if _, err := conn.Write([]byte(`{"type":"hello","protocol":1,"design":"mvd"}` + "\n")); err != nil {
		t.Fatal(err)
	}

	m, _ := appForTest(t, []string{"https://a"}, nil)
	am := m.(appModel)
	am.link = link
	var model tea.Model = am
	model = drive(t, model, tea.WindowSizeMsg{Width: 100, Height: 30})
	if v := d.state("List")["Title"].(string); v == "" {
		t.Error("the list was published without a title")
	}

	model = drive(t, model, key("ctrl+p"))
	items := d.state("Preferences")["Items"].([]any)
	if len(items) < 3 {
		t.Fatalf("preferences published %d items", len(items))
	}

	model = drive(t, model, designerMsg{Page: "Preferences", Name: "Pick", Y: 2})
	items = d.state("Preferences")["Items"].([]any)
	if mark := items[2].(map[string]any)["Mark"]; mark != "▸ " {
		t.Errorf("row 2 mark %q after the designer's click, want the cursor", mark)
	}
	_ = model
}
