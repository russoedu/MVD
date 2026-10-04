package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// occupy holds a loopback port with the given handler and returns its address.
func occupy(t *testing.T, handler http.Handler) string {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return strings.TrimPrefix(server.URL, "http://")
}

func TestAFreeAddressIsListenedOnDirectly(t *testing.T) {
	listener, existing, err := listen("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	if existing != "" {
		t.Errorf("existing = %q", existing)
	}
}

func TestAnotherMVDOnTheAddressIsReportedSoItCanBeOpenedInstead(t *testing.T) {
	address := occupy(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"version":12,"idle":false,"tally":{"total":0}}`))
	}))

	listener, existing, err := listen(address)
	if err != nil {
		t.Fatal(err)
	}
	if listener != nil {
		_ = listener.Close()
		t.Fatal("started a second server next to the running one")
	}
	if existing != address {
		t.Errorf("existing = %q, want %q", existing, address)
	}
}

func TestSomethingElseOnTheAddressIsLeftAloneAndAnotherPortIsTaken(t *testing.T) {
	cases := map[string]http.Handler{
		"a page":        http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("<html>")) }),
		"an error":      http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "no", http.StatusInternalServerError) }),
		"other json":    http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"ok":true}`)) }),
		"a bare number": http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`12`)) }),
	}
	for name, handler := range cases {
		address := occupy(t, handler)

		listener, existing, err := listen(address)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if existing != "" || listener == nil {
			t.Fatalf("%s: existing = %q, listener = %v", name, existing, listener)
		}
		if listener.Addr().String() == address {
			t.Errorf("%s: took the occupied address", name)
		}
		host, _, _ := net.SplitHostPort(listener.Addr().String())
		if host != "127.0.0.1" {
			t.Errorf("%s: moved off loopback to %s", name, host)
		}
		_ = listener.Close()
	}
}

func TestAnAddressThatCannotBeListenedOnIsAnError(t *testing.T) {
	if _, _, err := listen("not an address"); err == nil {
		t.Fatal("expected an error")
	}
}
