package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakePicker struct {
	mu      sync.Mutex
	starts  []string
	path    string
	chosen  bool
	err     error
	opened  chan struct{}
	release chan struct{}
}

func (f *fakePicker) Pick(_ context.Context, start string) (string, bool, error) {
	f.mu.Lock()
	f.starts = append(f.starts, start)
	f.mu.Unlock()
	if f.opened != nil {
		f.opened <- struct{}{}
		<-f.release
	}

	return f.path, f.chosen, f.err
}

func TestAChosenFolderIsReturnedAndTheStartingPointIsPassedOn(t *testing.T) {
	picker := &fakePicker{path: "/music", chosen: true}

	w := post(New(newFake(), Options{Folders: picker}), "/api/folders/pick", `{"start":"/old"}`)

	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"path":"/music"`) || !strings.Contains(w.Body.String(), `"cancelled":false`) {
		t.Errorf("%d %s", w.Code, w.Body)
	}
	if len(picker.starts) != 1 || picker.starts[0] != "/old" {
		t.Errorf("starts = %v", picker.starts)
	}
}

func TestCancellingIsNotAnError(t *testing.T) {
	w := post(New(newFake(), Options{Folders: &fakePicker{}}), "/api/folders/pick", `{}`)

	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"cancelled":true`) || !strings.Contains(w.Body.String(), `"path":""`) {
		t.Errorf("%d %s", w.Code, w.Body)
	}
}

func TestAMachineWithoutADialogSaysSoDistinctlyFromAFailure(t *testing.T) {
	w := post(New(newFake(), Options{Folders: &fakePicker{err: ErrNoDialog}}), "/api/folders/pick", `{}`)
	if w.Code != http.StatusNotImplemented {
		t.Errorf("no dialog: %d", w.Code)
	}

	w = post(New(newFake(), Options{Folders: &fakePicker{err: errors.New("exit 1")}}), "/api/folders/pick", `{}`)
	if w.Code != http.StatusBadGateway || !strings.Contains(w.Body.String(), "exit 1") {
		t.Errorf("failure: %d %s", w.Code, w.Body)
	}
}

func TestOnlyOneChooserIsOpenAtATime(t *testing.T) {
	picker := &fakePicker{path: "/a", chosen: true, opened: make(chan struct{}, 2), release: make(chan struct{})}
	handler := New(newFake(), Options{Folders: picker})

	first := make(chan int)
	go func() { first <- post(handler, "/api/folders/pick", `{}`).Code }()
	<-picker.opened

	if w := post(handler, "/api/folders/pick", `{}`); w.Code != http.StatusConflict {
		t.Errorf("second request while one is open: %d", w.Code)
	}

	picker.release <- struct{}{}
	select {
	case code := <-first:
		if code != http.StatusOK {
			t.Errorf("first request: %d", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the first request never finished")
	}

	picker.opened = nil
	if w := post(handler, "/api/folders/pick", `{}`); w.Code != http.StatusOK {
		t.Errorf("after it closed: %d", w.Code)
	}
}

func TestMalformedPickBodiesAreRefusedWithoutOpeningAnything(t *testing.T) {
	for name, body := range map[string]string{"not json": `x`, "a misspelling": `{"begin":"/x"}`} {
		picker := &fakePicker{}
		w := post(New(newFake(), Options{Folders: picker}), "/api/folders/pick", body)
		if w.Code != http.StatusBadRequest || len(picker.starts) != 0 {
			t.Errorf("%s: %d, opened %d", name, w.Code, len(picker.starts))
		}
	}
}

func TestPickingIsGuardedAndAbsentWithoutAPicker(t *testing.T) {
	picker := &fakePicker{}
	r := httptest.NewRequest(http.MethodPost, "/api/folders/pick", strings.NewReader(`{}`))
	r.Host = "evil.example"
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	New(newFake(), Options{Folders: picker}).ServeHTTP(w, r)
	if w.Code != http.StatusForbidden || len(picker.starts) != 0 {
		t.Errorf("foreign host: %d, opened %d", w.Code, len(picker.starts))
	}

	if w := post(New(newFake(), Options{}), "/api/folders/pick", `{}`); w.Code != http.StatusNotFound {
		t.Errorf("no picker configured: %d", w.Code)
	}
}
