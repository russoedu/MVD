package deps

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestADownloadIsWrittenToDiskAndItsChecksumIsReturned(t *testing.T) {
	body := strings.Repeat("program bytes ", 1000)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(body)) }))
	defer server.Close()
	dest := filepath.Join(t.TempDir(), "file")

	sum, err := downloadFile(server.URL, dest)

	if err != nil {
		t.Fatal(err)
	}
	if read(t, dest) != body {
		t.Error("the file does not hold what was served")
	}
	if sum != sumOf([]byte(body)) {
		t.Errorf("checksum = %s, want %s", sum, sumOf([]byte(body)))
	}
}

func TestAnErrorStatusIsAFailureThatLeavesNoFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { http.NotFound(w, nil) }))
	defer server.Close()
	dest := filepath.Join(t.TempDir(), "file")

	_, err := downloadFile(server.URL, dest)

	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("err = %v", err)
	}
	if _, statErr := os.Stat(dest); statErr == nil {
		t.Error("a file was left behind")
	}
}

func TestAConnectionThatDropsHalfWayLeavesNoPartialFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "100000")
		_, _ = w.Write([]byte("only a little"))
		// Returning now ends the response short of the length it promised.
	}))
	defer server.Close()
	dest := filepath.Join(t.TempDir(), "file")

	_, err := downloadFile(server.URL, dest)

	if err == nil {
		t.Fatal("a short download should be an error")
	}
	if _, statErr := os.Stat(dest); statErr == nil {
		t.Error("a partial file was left behind")
	}
}
