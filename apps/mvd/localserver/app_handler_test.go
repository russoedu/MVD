package localserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func fetch(t *testing.T, handler http.Handler, method, path, host string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(""))
	r.Host = host
	if method == http.MethodPost {
		r.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func TestTheAppAnswersWhoItIsAndTheOldStateRouteForOlderVersions(t *testing.T) {
	handler := NewHandler(nil, nil)

	ping := fetch(t, handler, http.MethodGet, "/api/ping", "127.0.0.1:8421")
	if ping.Code != http.StatusOK || !strings.Contains(ping.Body.String(), `"app":"mvd"`) {
		t.Errorf("/api/ping: %d %s", ping.Code, ping.Body)
	}

	state := fetch(t, handler, http.MethodGet, "/api/state", "127.0.0.1:8421")
	if state.Code != http.StatusOK || !strings.Contains(state.Body.String(), `"version":1`) || !strings.Contains(state.Body.String(), `"tally":{}`) {
		t.Errorf("/api/state: %d %s", state.Code, state.Body)
	}
}

func TestThePageIsServedForEveryOtherPath(t *testing.T) {
	handler := NewHandler(nil, nil)

	for _, path := range []string{"/", "/terminal", "/anything/at/all"} {
		page := fetch(t, handler, http.MethodGet, path, "127.0.0.1:8421")
		if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "<html") {
			t.Errorf("%s: %d", path, page.Code)
		}
	}
}

func TestAnUnknownAPIPathIsNotFoundRatherThanThePage(t *testing.T) {
	w := fetch(t, NewHandler(nil, nil), http.MethodGet, "/api/nope", "127.0.0.1:8421")

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "<html") {
		t.Error("answered an API path with the page")
	}
}

func TestTheAPIRefusesAForeignHost(t *testing.T) {
	if w := fetch(t, NewHandler(nil, nil), http.MethodGet, "/api/ping", "evil.example"); w.Code != http.StatusForbidden {
		t.Errorf("status = %d", w.Code)
	}
}

func TestTheTerminalInterfaceIsMountedOnlyWhenThereIsOne(t *testing.T) {
	terminal := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("terminal")) })

	if w := fetch(t, NewHandler(nil, terminal), http.MethodGet, TerminalPath, "127.0.0.1:8421"); w.Body.String() != "terminal" {
		t.Errorf("with a terminal interface %s should reach it, got %q", TerminalPath, w.Body)
	}
	if w := fetch(t, NewHandler(nil, nil), http.MethodGet, TerminalPath, "127.0.0.1:8421"); strings.Contains(w.Body.String(), "terminal") {
		t.Errorf("without one %s should fall through to the page", TerminalPath)
	}
}
