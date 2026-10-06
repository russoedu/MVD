package localserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"youtube-downloader/apps/mvd/uninstall"
)

type fakeUninstaller struct {
	mu      sync.Mutex
	asked   []*bool
	err     error
	removed chan struct{}
	opened  chan struct{}
	release chan struct{}
}

func (f *fakeUninstaller) Confirm(deletePreferences *bool) (func(), error) {
	f.mu.Lock()
	f.asked = append(f.asked, deletePreferences)
	f.mu.Unlock()
	if f.opened != nil {
		f.opened <- struct{}{}
		<-f.release
	}
	if f.err != nil {
		return nil, f.err
	}

	return func() { close(f.removed) }, nil
}

func newFakeUninstaller() *fakeUninstaller {
	return &fakeUninstaller{removed: make(chan struct{})}
}

func waitRemoved(t *testing.T, f *fakeUninstaller) {
	t.Helper()
	select {
	case <-f.removed:
	case <-time.After(2 * time.Second):
		t.Fatal("the removal never ran")
	}
}

func post(handler http.Handler, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	r.Host = "127.0.0.1:8421"
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func TestAConfirmedRemovalIsAcceptedAndThenRunsWithTheChoiceThatWasMade(t *testing.T) {
	for body, want := range map[string]bool{`{"deletePreferences":true}`: true, `{"deletePreferences":false}`: false} {
		fake := newFakeUninstaller()

		w := post(NewHandler(fake, nil), "/api/uninstall", body)

		if w.Code != http.StatusAccepted || !strings.Contains(w.Body.String(), `"status":"removing"`) {
			t.Errorf("%s: %d %s", body, w.Code, w.Body)
		}
		waitRemoved(t, fake)
		if len(fake.asked) != 1 || fake.asked[0] == nil || *fake.asked[0] != want {
			t.Errorf("%s: asked = %v", body, fake.asked)
		}
	}
}

func TestWithoutAChoiceThePersonIsAskedTheQuestionToo(t *testing.T) {
	for _, body := range []string{`{}`, ``} {
		fake := newFakeUninstaller()

		w := post(NewHandler(fake, nil), "/api/uninstall", body)

		if w.Code != http.StatusAccepted {
			t.Errorf("%q: %d %s", body, w.Code, w.Body)
		}
		waitRemoved(t, fake)
		if len(fake.asked) != 1 || fake.asked[0] != nil {
			t.Errorf("%q: asked = %v", body, fake.asked)
		}
	}
}

func TestDecliningOnTheMachinesOwnScreenRemovesNothingAndSaysSo(t *testing.T) {
	fake := newFakeUninstaller()
	fake.err = uninstall.ErrDeclined

	w := post(NewHandler(fake, nil), "/api/uninstall", `{"deletePreferences":true}`)

	if w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "nothing was removed") {
		t.Errorf("%d %s", w.Code, w.Body)
	}
	select {
	case <-fake.removed:
		t.Error("the removal ran although the person said no")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestAMachineThatCannotAskRemovesNothingAndSaysDistinctlyFromAFailure(t *testing.T) {
	fake := newFakeUninstaller()
	fake.err = uninstall.ErrNoDialog
	w := post(NewHandler(fake, nil), "/api/uninstall", `{}`)
	if w.Code != http.StatusNotImplemented || !strings.Contains(w.Body.String(), "-uninstall") {
		t.Errorf("no dialog: %d %s", w.Code, w.Body)
	}

	fake = newFakeUninstaller()
	fake.err = errors.New("boom")
	w = post(NewHandler(fake, nil), "/api/uninstall", `{}`)
	if w.Code != http.StatusBadGateway || !strings.Contains(w.Body.String(), "boom") {
		t.Errorf("failure: %d %s", w.Code, w.Body)
	}
}

func TestOnlyOneConfirmationIsOpenAtATime(t *testing.T) {
	fake := newFakeUninstaller()
	fake.opened = make(chan struct{}, 2)
	fake.release = make(chan struct{})
	handler := NewHandler(fake, nil)

	first := make(chan int)
	go func() { first <- post(handler, "/api/uninstall", `{}`).Code }()
	<-fake.opened

	if w := post(handler, "/api/uninstall", `{}`); w.Code != http.StatusConflict {
		t.Errorf("second request while one is open: %d", w.Code)
	}
	if len(fake.asked) != 1 {
		t.Errorf("the second request was put to the person: %d questions", len(fake.asked))
	}

	fake.release <- struct{}{}
	select {
	case code := <-first:
		if code != http.StatusAccepted {
			t.Errorf("first request: %d", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the first request never finished")
	}
	waitRemoved(t, fake)
}

func TestMalformedUninstallBodiesAreRefusedWithoutAskingAnything(t *testing.T) {
	for name, body := range map[string]string{
		"not json":          `x`,
		"a misspelling":     `{"deletePrefs":true}`,
		"a string for flag": `{"deletePreferences":"yes"}`,
	} {
		fake := newFakeUninstaller()

		w := post(NewHandler(fake, nil), "/api/uninstall", body)

		if w.Code != http.StatusBadRequest || len(fake.asked) != 0 {
			t.Errorf("%s: %d, asked %d", name, w.Code, len(fake.asked))
		}
	}
}

func TestUninstallingIsGuardedAndAbsentWithoutAnUninstaller(t *testing.T) {
	fake := newFakeUninstaller()
	for name, mutate := range map[string]func(r *http.Request){
		"a foreign host":    func(r *http.Request) { r.Host = "evil.example" },
		"a foreign origin":  func(r *http.Request) { r.Header.Set("Origin", "https://evil.example") },
		"a cross-site page": func(r *http.Request) { r.Header.Set("Sec-Fetch-Site", "cross-site") },
		"a form post":       func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") },
	} {
		r := httptest.NewRequest(http.MethodPost, "/api/uninstall", strings.NewReader(`{"deletePreferences":true}`))
		r.Host = "127.0.0.1:8421"
		r.Header.Set("Content-Type", "application/json")
		mutate(r)
		w := httptest.NewRecorder()

		NewHandler(fake, nil).ServeHTTP(w, r)

		if w.Code != http.StatusForbidden {
			t.Errorf("%s: %d", name, w.Code)
		}
	}
	if len(fake.asked) != 0 {
		t.Errorf("a refused request was put to the person: %d", len(fake.asked))
	}

	if w := post(NewHandler(nil, nil), "/api/uninstall", `{}`); w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Errorf("no uninstaller configured: %d", w.Code)
	}
}

func TestOnlyPostReachesTheUninstallRoute(t *testing.T) {
	w := fetch(t, NewHandler(newFakeUninstaller(), nil), http.MethodGet, "/api/uninstall", "127.0.0.1:8421")

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET: %d", w.Code)
	}
}
