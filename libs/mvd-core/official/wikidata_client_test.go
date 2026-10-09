package official

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const wikidataAnswer = `{"results":{"bindings":[
 {"vid":{"value":"hCuMWrfXG4E"},"perfLabel":{"value":"Billy Joel"}},
 {"vid":{"value":"0HTexqxo1og"},"perfLabel":{"value":"Westlife"}},
 {"vid":{"value":"hCuMWrfXG4E"},"perfLabel":{"value":"Billy Joel"}},
 {"vid":{"value":"short"},"perfLabel":{"value":"Billy Joel"}}
]}}`

func wikidataServer(t *testing.T, status int, answer string, seen *string) *WikidataClient {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if seen != nil {
			*seen = r.URL.Query().Get("query")
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("Wikidata asks for a User-Agent")
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(answer))
	}))
	t.Cleanup(server.Close)

	client := NewWikidataClient()
	client.Endpoint = server.URL
	client.MinInterval = 0
	return client
}

func TestWikidataVideoIDsKeepsTheSongsOfTheArtist(t *testing.T) {
	var query string
	client := wikidataServer(t, http.StatusOK, wikidataAnswer, &query)

	got, err := client.VideoIDs(`Uptown Girl (Remastered)`, []string{"Billy Joel"})
	if err != nil || len(got) != 1 || got[0] != "hCuMWrfXG4E" {
		t.Fatalf("got %v, %v; want only Billy Joel's video, once (a cover and a malformed id are dropped)", got, err)
	}
	for _, want := range []string{`"Uptown Girl (Remastered)"@en`, `"Uptown Girl"@en`, "wdt:P1651", "wdt:P175"} {
		if !strings.Contains(query, want) {
			t.Errorf("the query should contain %s, got %s", want, query)
		}
	}
}

func TestWikidataVideoIDsEscapesTheTitle(t *testing.T) {
	var query string
	client := wikidataServer(t, http.StatusOK, `{"results":{"bindings":[]}}`, &query)

	if _, err := client.VideoIDs(`Say "Hello" \ Goodbye`, []string{"Band"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(query, `"Say \"Hello\" \\ Goodbye"@en`) {
		t.Errorf("quotes and backslashes must be escaped, got %s", query)
	}
}

func TestWikidataVideoIDsErrorsAndEmptyInput(t *testing.T) {
	client := wikidataServer(t, http.StatusTooManyRequests, "slow down", nil)
	if _, err := client.VideoIDs("Song", []string{"Band"}); err == nil || !strings.Contains(err.Error(), "429") {
		t.Errorf("want the server's failure, got %v", err)
	}
	bad := wikidataServer(t, http.StatusOK, "not json", nil)
	if _, err := bad.VideoIDs("Song", []string{"Band"}); err == nil {
		t.Error("an answer that is not JSON is an error")
	}
	if ids, err := client.VideoIDs("", []string{"Band"}); err != nil || ids != nil {
		t.Errorf("no title, no request: got %v, %v", ids, err)
	}
	if ids, err := client.VideoIDs("Song", nil); err != nil || ids != nil {
		t.Errorf("no artist, no request: got %v, %v", ids, err)
	}
}

func TestPerformedBy(t *testing.T) {
	if !performedBy("The Weeknd", []string{"Weeknd"}) || !performedBy("Weeknd", []string{"The Weeknd"}) {
		t.Error("the leading The does not matter")
	}
	if !performedBy("ADÉLA", []string{"Adela"}) {
		t.Error("accents do not matter")
	}
	if performedBy("Westlife", []string{"Billy Joel"}) || performedBy("", []string{"Billy Joel"}) {
		t.Error("another performer is not the artist")
	}
}
