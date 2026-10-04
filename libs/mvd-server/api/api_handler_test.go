package api

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"youtube-downloader/libs/mvd-server/session"
	"youtube-downloader/libs/mvd-server/snapshot"
)

// fakeSessions stands in for the run: a snapshot whose version a test can bump,
// and a record of what the handlers asked of it.
type fakeSessions struct {
	mu       sync.Mutex
	snap     snapshot.Snapshot
	changed  chan struct{}
	added    [][]string
	addErr   error
	addOut   session.AddResult
	retried  []int
	playlist []int
}

func newFake() *fakeSessions {
	return &fakeSessions{changed: make(chan struct{})}
}

func (f *fakeSessions) Snapshot() snapshot.Snapshot {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.snap
}

func (f *fakeSessions) WaitSince(ctx context.Context, version int64) (snapshot.Snapshot, bool) {
	for {
		f.mu.Lock()
		if f.snap.Version > version {
			s := f.snap
			f.mu.Unlock()
			return s, true
		}
		ch := f.changed
		f.mu.Unlock()
		select {
		case <-ch:
		case <-ctx.Done():
			return snapshot.Snapshot{}, false
		}
	}
}

func (f *fakeSessions) bump(idle bool) {
	f.mu.Lock()
	f.snap.Version++
	f.snap.Idle = idle
	close(f.changed)
	f.changed = make(chan struct{})
	f.mu.Unlock()
}

func (f *fakeSessions) Add(raw []string) (session.AddResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.added = append(f.added, raw)
	return f.addOut, f.addErr
}

func (f *fakeSessions) Retry(entry int) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.retried = append(f.retried, entry)
	return entry == 7
}

func (f *fakeSessions) RetryPlaylist(playlist int) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.playlist = append(f.playlist, playlist)
	return 3
}

func post(handler http.Handler, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	r.Host = "127.0.0.1:8080"
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func get(handler http.Handler, path string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, path, nil)
	r.Host = "127.0.0.1:8080"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func TestStateIsTheCurrentSnapshotAndIsNeverCached(t *testing.T) {
	fake := newFake()
	fake.snap.Version = 4
	fake.snap.Tally.Total = 9

	w := get(New(fake, Options{}), "/api/state")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q", got)
	}
	var snap snapshot.Snapshot
	if err := json.Unmarshal(w.Body.Bytes(), &snap); err != nil {
		t.Fatal(err)
	}
	if snap.Version != 4 || snap.Tally.Total != 9 {
		t.Errorf("snapshot = %+v", snap)
	}
}

func TestEveryRouteIsGuardedBeforeItIsReached(t *testing.T) {
	fake := newFake()
	handler := New(fake, Options{})

	r := httptest.NewRequest(http.MethodPost, "/api/sources", strings.NewReader(`{"urls":["https://a.example/x"]}`))
	r.Host = "evil.example"
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d", w.Code)
	}
	if len(fake.added) != 0 {
		t.Errorf("a refused request still reached the session: %v", fake.added)
	}

	r = httptest.NewRequest(http.MethodPost, "/api/sources", strings.NewReader(`urls=x`))
	r.Host = "127.0.0.1:8080"
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden || len(fake.added) != 0 {
		t.Errorf("a form post was let through: %d %v", w.Code, fake.added)
	}
}

func TestAddingURLsPassesListAndTextLinesAndReportsTheOutcome(t *testing.T) {
	fake := newFake()
	fake.addOut = session.AddResult{Added: []string{"https://a.example/1"}, Rejected: []string{"nope"}}

	w := post(New(fake, Options{}), "/api/sources",
		`{"urls":["https://a.example/1"],"text":"https://a.example/2\r\nnope\n"}`)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body)
	}
	if len(fake.added) != 1 {
		t.Fatalf("Add was called %d times", len(fake.added))
	}
	got := fake.added[0]
	want := []string{"https://a.example/1", "https://a.example/2", "nope", ""}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("lines = %q, want %q", got, want)
	}
	var out session.AddResult
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Added) != 1 || len(out.Rejected) != 1 {
		t.Errorf("result = %+v", out)
	}
}

func TestAddingMalformedBodiesIsTheCallersMistake(t *testing.T) {
	for name, body := range map[string]string{
		"not json":      `urls=x`,
		"empty":         ``,
		"a misspelling": `{"url":["https://a.example"]}`,
		"wrong type":    `{"urls":"https://a.example"}`,
	} {
		fake := newFake()
		w := post(New(fake, Options{}), "/api/sources", body)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d", name, w.Code)
		}
		if len(fake.added) != 0 {
			t.Errorf("%s: reached the session", name)
		}
	}
}

func TestAnOversizedBodyIsRefusedRatherThanRead(t *testing.T) {
	fake := newFake()
	body := `{"text":"` + strings.Repeat("a", maxSourcesBody+10) + `"}`

	w := post(New(fake, Options{}), "/api/sources", body)

	if w.Code != http.StatusBadRequest || len(fake.added) != 0 {
		t.Errorf("status = %d, added = %d", w.Code, len(fake.added))
	}
}

func TestAddingWhileShuttingDownIsUnavailableAndOtherFailuresAreOursNotTheirs(t *testing.T) {
	fake := newFake()
	fake.addErr = session.ErrClosed
	if w := post(New(fake, Options{}), "/api/sources", `{"urls":[]}`); w.Code != http.StatusServiceUnavailable {
		t.Errorf("closed: status = %d", w.Code)
	}

	fake = newFake()
	fake.addErr = errors.New("yt-dlp is missing")
	w := post(New(fake, Options{}), "/api/sources", `{"urls":[]}`)
	if w.Code != http.StatusBadGateway || !strings.Contains(w.Body.String(), "yt-dlp is missing") {
		t.Errorf("failure: status = %d body = %s", w.Code, w.Body)
	}
}

func TestRetryRoutesPassTheNumberAndRejectNonNumbers(t *testing.T) {
	fake := newFake()
	handler := New(fake, Options{})

	if w := post(handler, "/api/entries/7/retry", `{}`); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"ok":true`) {
		t.Errorf("entry: %d %s", w.Code, w.Body)
	}
	if w := post(handler, "/api/entries/8/retry", `{}`); !strings.Contains(w.Body.String(), `"ok":false`) {
		t.Errorf("entry that is not retryable: %s", w.Body)
	}
	if w := post(handler, "/api/playlists/2/retry", `{}`); !strings.Contains(w.Body.String(), `"retried":3`) {
		t.Errorf("playlist: %s", w.Body)
	}
	for _, path := range []string{"/api/entries/x/retry", "/api/playlists/1.5/retry"} {
		if w := post(handler, path, `{}`); w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d", path, w.Code)
		}
	}
	if len(fake.retried) != 2 || fake.retried[0] != 7 || len(fake.playlist) != 1 || fake.playlist[0] != 2 {
		t.Errorf("retried = %v, playlists = %v", fake.retried, fake.playlist)
	}
}

func TestRoutesAcceptOnlyTheirOwnMethod(t *testing.T) {
	handler := New(newFake(), Options{})
	r := httptest.NewRequest(http.MethodGet, "/api/sources", nil)
	r.Host = "127.0.0.1:8080"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d", w.Code)
	}
}

// readEvent reads one SSE frame and returns its event name and data.
func readEvent(t *testing.T, scanner *bufio.Scanner) (event, data string) {
	t.Helper()
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "":
			if event != "" {
				return event, data
			}
		case strings.HasPrefix(line, "event: "):
			event = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			data = strings.TrimPrefix(line, "data: ")
		}
	}
	t.Fatalf("the stream ended: %v", scanner.Err())
	return "", ""
}

func openStream(t *testing.T, fake *fakeSessions, options Options) (*bufio.Scanner, *http.Response, func()) {
	t.Helper()
	server := httptest.NewServer(New(fake, options))
	resp, err := http.Get(server.URL + "/api/events")
	if err != nil {
		server.Close()
		t.Fatal(err)
	}
	return bufio.NewScanner(resp.Body), resp, func() {
		_ = resp.Body.Close()
		server.CloseClientConnections()
		server.Close()
	}
}

func TestTheStreamSendsASnapshotAtOnceAndThenOneForEachChange(t *testing.T) {
	fake := newFake()
	fake.snap.Version = 1
	scanner, resp, closeAll := openStream(t, fake, Options{Throttle: time.Millisecond})
	defer closeAll()

	if got := resp.Header.Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("Content-Type = %q", got)
	}

	event, data := readEvent(t, scanner)
	var snap snapshot.Snapshot
	if event != "snapshot" || json.Unmarshal([]byte(data), &snap) != nil || snap.Version != 1 {
		t.Fatalf("first frame = %q %q", event, data)
	}

	fake.bump(true)
	_, data = readEvent(t, scanner)
	if err := json.Unmarshal([]byte(data), &snap); err != nil || snap.Version != 2 || !snap.Idle {
		t.Fatalf("second frame = %q (%v)", data, err)
	}
}

func TestChangesInsideTheThrottleAreSentAsOneSnapshotWithTheLatestState(t *testing.T) {
	fake := newFake()
	fake.snap.Version = 1
	const throttle = 300 * time.Millisecond
	scanner, _, closeAll := openStream(t, fake, Options{Throttle: throttle})
	defer closeAll()

	readEvent(t, scanner)
	fake.bump(false)
	readEvent(t, scanner)
	sentAt := time.Now()

	// The handler is now holding off for the throttle, so these arrive inside it.
	for range 5 {
		fake.bump(false)
	}
	fake.bump(true)

	_, data := readEvent(t, scanner)
	var snap snapshot.Snapshot
	if err := json.Unmarshal([]byte(data), &snap); err != nil {
		t.Fatal(err)
	}
	if snap.Version != 8 || !snap.Idle {
		t.Errorf("version = %d idle = %v, want the last state (8, idle)", snap.Version, snap.Idle)
	}
	if waited := time.Since(sentAt); waited < throttle/2 {
		t.Errorf("the next snapshot came after %v, inside the %v throttle", waited, throttle)
	}
}

func TestAQuietStreamSendsKeepaliveComments(t *testing.T) {
	fake := newFake()
	scanner, _, closeAll := openStream(t, fake, Options{Throttle: time.Millisecond, Keepalive: 20 * time.Millisecond})
	defer closeAll()

	readEvent(t, scanner)
	deadline := time.After(2 * time.Second)
	found := make(chan struct{})
	go func() {
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), ": keepalive") {
				close(found)
				return
			}
		}
	}()
	select {
	case <-found:
	case <-deadline:
		t.Fatal("no keepalive within 2s")
	}
}

func TestADisconnectedClientReleasesItsHandler(t *testing.T) {
	fake := newFake()
	var wg sync.WaitGroup
	wg.Add(1)
	handler := New(fake, Options{Throttle: time.Millisecond})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()
		handler.ServeHTTP(w, r)
	}))
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/events")
	if err != nil {
		t.Fatal(err)
	}
	readEvent(t, bufio.NewScanner(resp.Body))
	_ = resp.Body.Close()

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("the handler is still running after the client left")
	}
}
