package localserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"youtube-downloader/libs/mvd-server/api"
)

// runningMVD answers like the app does: a snapshot on /api/state, and whatever the test
// wants on /api/uninstall.
func runningMVD(t *testing.T, uninstall http.HandlerFunc) string {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/state", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"version":1,"tally":{}}`))
	})
	mux.HandleFunc("/api/uninstall", uninstall)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return strings.TrimPrefix(server.URL, "http://")
}

func TestWhenNothingIsRunningNothingIsAsked(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	address := strings.TrimPrefix(server.URL, "http://")
	server.Close()

	running, err := RequestUninstall(address)

	if running || err != nil {
		t.Errorf("running = %v, err = %v", running, err)
	}
}

func TestARunningAppThatAcceptsIsLeftToRemoveItself(t *testing.T) {
	var gotBody string
	address := runningMVD(t, func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 100)
		n, _ := r.Body.Read(buf)
		gotBody = string(buf[:n])
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("method %s, content type %q", r.Method, r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusAccepted)
	})

	running, err := RequestUninstall(address)

	if !running || err != nil {
		t.Errorf("running = %v, err = %v", running, err)
	}
	if gotBody != "{}" {
		t.Errorf("body = %q: the person should be asked about the preferences on their screen", gotBody)
	}
}

func TestARunningAppWhosePersonSaidNoIsReportedAsDeclined(t *testing.T) {
	address := runningMVD(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusConflict) })

	running, err := RequestUninstall(address)

	if !running || !errors.Is(err, api.ErrUninstallDeclined) {
		t.Errorf("running = %v, err = %v", running, err)
	}
}

func TestAnyOtherAnswerIsAnErrorWithTheReasonTheAppGave(t *testing.T) {
	address := runningMVD(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		_, _ = w.Write([]byte(`{"error":"nothing to ask with"}`))
	})

	running, err := RequestUninstall(address)

	if !running || err == nil || !strings.Contains(err.Error(), "nothing to ask with") {
		t.Errorf("running = %v, err = %v", running, err)
	}
}
