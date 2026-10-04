package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebHandlerServesTheBuiltFrontend(t *testing.T) {
	for _, target := range []string{"/", "/a/client/side/route"} {
		recorder := httptest.NewRecorder()
		webHandler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "<html") {
			t.Fatalf("GET %s: %d %q", target, recorder.Code, recorder.Body.String())
		}
	}
}
