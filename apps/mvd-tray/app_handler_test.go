package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"youtube-downloader/libs/mvd-server/session"
	"youtube-downloader/libs/mvd-server/snapshot"
)

type stubSessions struct{}

func (stubSessions) Snapshot() snapshot.Snapshot { return snapshot.Snapshot{Version: 5} }
func (stubSessions) WaitSince(context.Context, int64) (snapshot.Snapshot, bool) {
	return snapshot.Snapshot{}, false
}
func (stubSessions) Add([]string) (session.AddResult, error) { return session.AddResult{}, nil }
func (stubSessions) Retry(int) bool                          { return false }
func (stubSessions) RetryPlaylist(int) int                   { return 0 }

func fetch(t *testing.T, path, host string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, path, nil)
	r.Host = host
	w := httptest.NewRecorder()
	newAppHandler(stubSessions{}).ServeHTTP(w, r)
	return w
}

func TestTheAPIAndTheFrontendShareOneSite(t *testing.T) {
	api := fetch(t, "/api/state", "127.0.0.1:8421")
	if api.Code != http.StatusOK || !strings.Contains(api.Body.String(), `"version":5`) {
		t.Errorf("/api/state: %d %s", api.Code, api.Body)
	}

	page := fetch(t, "/", "127.0.0.1:8421")
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "<html") {
		t.Errorf("/: %d", page.Code)
	}
}

func TestAnUnknownAPIPathIsNotFoundRatherThanTheFrontendPage(t *testing.T) {
	w := fetch(t, "/api/nope", "127.0.0.1:8421")
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "<html") {
		t.Error("answered an API path with the frontend page")
	}
}

func TestTheAPIStillRefusesAForeignHostThroughTheMountedSite(t *testing.T) {
	if w := fetch(t, "/api/state", "evil.example"); w.Code != http.StatusForbidden {
		t.Errorf("status = %d", w.Code)
	}
}
