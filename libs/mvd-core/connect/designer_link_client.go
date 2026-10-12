package connect

import (
	"bytes"
	"encoding/json"
	"os"
	"sync"

	"github.com/meta-ui-labs/cuppa/libs/connector/message"
	"github.com/meta-ui-labs/cuppa/libs/connector/serve"
)

// appName is how the designer shows this app.
const appName = "MVD"

// Event is what a click or key on the designer's drawing raised: the page, the name of
// the event in the design and the cell inside the component that was clicked (0, 0 for a
// key).
type Event struct {
	Page string
	Name string
	X, Y int
}

// Link is the app's end of the connection. A nil *Link is a connection that was not asked
// for: every method of it does nothing, so a screen needs no check before using it.
type Link struct {
	server *serve.Server

	mu   sync.Mutex
	last map[string][]byte // the inputs last published, by page, to skip what did not change
}

var (
	once    sync.Once
	process *Link
)

// FromEnvironment is the process's link: it listens on the address of EnvVar the first
// time it is called and returns nil when the connection was not asked for or the address
// cannot be used (a connection is a development aid, never a reason to stop the app).
func FromEnvironment(pages ...string) *Link {
	once.Do(func() {
		addr, ok := AddressFrom(os.Getenv(EnvVar))
		if !ok {
			return
		}
		if l, err := Listen(addr, pages...); err == nil {
			process = l
		}
	})
	return process
}

// Listen starts listening on addr, which must be on this computer.
func Listen(addr string, pages ...string) (*Link, error) {
	s, err := serve.Listen(addr, serve.Options{App: appName, Pages: pages})
	if err != nil {
		return nil, err
	}
	return &Link{server: s, last: map[string][]byte{}}, nil
}

// Addr is the address the link listens on.
func (l *Link) Addr() string {
	if l == nil {
		return ""
	}
	return l.server.Addr()
}

// Publish tells every designer the inputs of a page, by the input names of its design,
// when they changed since the last time, and says whether the page is the one showing.
func (l *Link) Publish(page string, inputs map[string]any, current bool) {
	if l == nil {
		return
	}
	raw, err := json.Marshal(inputs)
	if err != nil {
		return
	}
	// What changed is the inputs and whether the page is current; both are in the key.
	if current {
		raw = append(raw, 1)
	}
	l.mu.Lock()
	same := bytes.Equal(l.last[page], raw)
	l.last[page] = raw
	l.mu.Unlock()
	if same {
		return
	}
	_ = l.server.Publish(page, inputs, current)
}

// Next waits for the next event the designer sends; false when the link is closed.
func (l *Link) Next() (Event, bool) {
	if l == nil {
		return Event{}, false
	}
	for {
		select {
		case m := <-l.server.Events():
			if m.Type != message.Event {
				continue
			}
			return Event{Page: m.Page, Name: m.Name, X: m.X, Y: m.Y}, true
		case <-l.server.Done():
			return Event{}, false
		}
	}
}

// Close stops listening.
func (l *Link) Close() {
	if l != nil {
		_ = l.server.Close()
	}
}
