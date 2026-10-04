package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"youtube-downloader/libs/mvd-server/settings"
)

type fakeStore struct {
	doc     settings.Document
	loadErr error
	saveErr error
	saved   []settings.Settings
}

func (f *fakeStore) Load() (settings.Document, error) { return f.doc, f.loadErr }

func (f *fakeStore) Save(s settings.Settings) error {
	f.saved = append(f.saved, s)
	if f.saveErr == nil {
		f.doc.Settings = s
	}

	return f.saveErr
}

func put(handler http.Handler, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPut, path, strings.NewReader(body))
	r.Host = "127.0.0.1:8080"
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	return w
}

func TestGettingTheSettingsReturnsTheValuesAndTheChoices(t *testing.T) {
	store := &fakeStore{doc: settings.Document{
		Settings: settings.Settings{OutputDir: "/music", VideoQuality: "720p"},
		Options:  settings.Options{VideoQualities: []string{"best", "720p"}},
	}}

	w := get(New(newFake(), Options{Settings: store}), "/api/settings")

	var got settings.Document
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
		t.Fatalf("%d %s (%v)", w.Code, w.Body, err)
	}
	if got.Settings.OutputDir != "/music" || len(got.Options.VideoQualities) != 2 {
		t.Errorf("got %+v", got)
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Error("settings must not be cached")
	}
}

func TestSavingPassesTheValuesOnAndReturnsWhatIsNowStored(t *testing.T) {
	store := &fakeStore{}

	w := put(New(newFake(), Options{Settings: store}), "/api/settings", `{"outputDir":"/new","maxConcurrentDownloads":3}`)

	if w.Code != http.StatusOK || len(store.saved) != 1 {
		t.Fatalf("%d %s saved=%d", w.Code, w.Body, len(store.saved))
	}
	if store.saved[0].OutputDir != "/new" || store.saved[0].MaxConcurrentDownloads != 3 {
		t.Errorf("saved %+v", store.saved[0])
	}
	if !strings.Contains(w.Body.String(), `"outputDir":"/new"`) {
		t.Errorf("the response is not the stored settings: %s", w.Body)
	}
}

func TestInvalidSettingsAreAnErrorThatNamesEachField(t *testing.T) {
	store := &fakeStore{saveErr: settings.Invalid{"outputDir": "choose a folder", "cookies": "no"}}

	w := put(New(newFake(), Options{Settings: store}), "/api/settings", `{}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", w.Code)
	}
	var body errorBody
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Fields["outputDir"] != "choose a folder" || body.Fields["cookies"] != "no" {
		t.Errorf("fields = %v", body.Fields)
	}
}

func TestAStorageFailureIsOurErrorNotTheCallers(t *testing.T) {
	store := &fakeStore{saveErr: errors.New("disk full")}

	w := put(New(newFake(), Options{Settings: store}), "/api/settings", `{}`)

	if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), "disk full") {
		t.Errorf("%d %s", w.Code, w.Body)
	}

	store = &fakeStore{loadErr: errors.New("unreadable")}
	if w := get(New(newFake(), Options{Settings: store}), "/api/settings"); w.Code != http.StatusInternalServerError {
		t.Errorf("load failure: %d", w.Code)
	}
}

func TestMalformedSettingsBodiesAreRefusedWithoutSaving(t *testing.T) {
	for name, body := range map[string]string{
		"not json":      `outputDir=/x`,
		"empty":         ``,
		"a misspelling": `{"outputFolder":"/x"}`,
		"wrong type":    `{"maxConcurrentDownloads":"4"}`,
	} {
		store := &fakeStore{}
		w := put(New(newFake(), Options{Settings: store}), "/api/settings", body)
		if w.Code != http.StatusBadRequest || len(store.saved) != 0 {
			t.Errorf("%s: status %d, saved %d", name, w.Code, len(store.saved))
		}
	}
}

func TestSettingsAreGuardedLikeEverythingElse(t *testing.T) {
	store := &fakeStore{}
	handler := New(newFake(), Options{Settings: store})

	r := httptest.NewRequest(http.MethodPut, "/api/settings", strings.NewReader(`{}`))
	r.Host = "127.0.0.1:8080"
	r.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden || len(store.saved) != 0 {
		t.Errorf("status %d, saved %d", w.Code, len(store.saved))
	}
}

func TestWithoutAStoreThereAreNoSettingsRoutes(t *testing.T) {
	handler := New(newFake(), Options{})
	if w := get(handler, "/api/settings"); w.Code != http.StatusNotFound {
		t.Errorf("GET: %d", w.Code)
	}
	if w := put(handler, "/api/settings", `{}`); w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Errorf("PUT: %d", w.Code)
	}
}
