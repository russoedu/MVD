package connect

import (
	"bufio"
	"encoding/json"
	"net"
	"testing"
	"time"
)

// designer is a fake designer: it dials the link, reads lines and sends events.
type designer struct {
	t     *testing.T
	conn  net.Conn
	lines *bufio.Scanner
}

func dialDesigner(t *testing.T, addr string) *designer {
	t.Helper()
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	d := &designer{t: t, conn: conn, lines: bufio.NewScanner(conn)}
	if _, err := conn.Write([]byte(`{"type":"hello","protocol":1,"design":"mvd"}` + "\n")); err != nil {
		t.Fatal(err)
	}
	return d
}

func (d *designer) read(typ string) map[string]any {
	d.t.Helper()
	_ = d.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	for d.lines.Scan() {
		var m map[string]any
		if err := json.Unmarshal(d.lines.Bytes(), &m); err != nil {
			d.t.Fatal(err)
		}
		if m["type"] == typ {
			return m
		}
	}
	d.t.Fatalf("no %q line: %v", typ, d.lines.Err())
	return nil
}

func TestAPublishedPageReachesTheDesignerAndAnEventComesBack(t *testing.T) {
	l, err := Listen("127.0.0.1:0", "Download")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	d := dialDesigner(t, l.Addr())
	if hello := d.read("hello"); hello["app"] != "MVD" {
		t.Errorf("hello: %v", hello)
	}

	l.Publish("Download", map[string]any{"Counters": "Queue 3"}, true)
	st := d.read("state")
	if st["page"] != "Download" || st["current"] != true || st["inputs"].(map[string]any)["Counters"] != "Queue 3" {
		t.Errorf("state: %v", st)
	}

	if _, err := d.conn.Write([]byte(`{"type":"event","page":"Download","name":"PickEntry","x":2,"y":4}` + "\n")); err != nil {
		t.Fatal(err)
	}
	ev, ok := l.Next()
	if !ok || ev != (Event{Page: "Download", Name: "PickEntry", X: 2, Y: 4}) {
		t.Errorf("event: %+v %v", ev, ok)
	}
}

func TestInputsThatDidNotChangeAreNotSentAgain(t *testing.T) {
	l, err := Listen("127.0.0.1:0", "Download")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	d := dialDesigner(t, l.Addr())
	d.read("hello")

	l.Publish("Download", map[string]any{"N": 1}, true)
	l.Publish("Download", map[string]any{"N": 1}, true)
	l.Publish("Download", map[string]any{"N": 2}, true)
	first := d.read("state")["inputs"].(map[string]any)["N"]
	second := d.read("state")["inputs"].(map[string]any)["N"]
	if first != float64(1) || second != float64(2) {
		t.Errorf("states: %v then %v, want 1 then 2", first, second)
	}
}

func TestAnAddressOffThisComputerIsRefused(t *testing.T) {
	if l, err := Listen("0.0.0.0:0", "Download"); err == nil {
		l.Close()
		t.Fatal("a non-loopback address must be refused")
	}
}

func TestANilLinkDoesNothing(t *testing.T) {
	var l *Link
	l.Publish("Download", map[string]any{"N": 1}, true)
	if _, ok := l.Next(); ok {
		t.Error("a nil link has no events")
	}
	l.Close()
}
