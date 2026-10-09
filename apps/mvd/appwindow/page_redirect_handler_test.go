package appwindow

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPageHandlerSendsTheWindowToTheAppAddress(t *testing.T) {
	rec := httptest.NewRecorder()
	pageHandler("http://127.0.0.1:8421").ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	body := rec.Body.String()
	if !strings.Contains(body, `location.replace("http://127.0.0.1:8421")`) {
		t.Errorf("the page should redirect to the app address, got %s", body)
	}
	if rec.Header().Get("Content-Type") != "text/html; charset=utf-8" || rec.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("unexpected headers %v", rec.Header())
	}
}

func TestPageHandlerEscapesTheAddress(t *testing.T) {
	rec := httptest.NewRecorder()
	pageHandler(`http://x/"</script><script>alert(1)`).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if strings.Contains(rec.Body.String(), "</script><script>") {
		t.Errorf("an address must not be able to close the script: %s", rec.Body.String())
	}
}
