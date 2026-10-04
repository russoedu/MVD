package deps

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func githubReturning(t *testing.T, status int, body string) (*httptest.Server, *http.Client) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/o/tool/releases/latest" {
			t.Errorf("asked for %s", r.URL.Path)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return server, server.Client()
}

const releaseJSON = `{
  "tag_name": "2026.08.19",
  "assets": [
    {"id": 1, "name": "other.zip", "browser_download_url": "https://x/other.zip", "digest": "sha256:aaaa", "updated_at": "2026-08-19T23:48:37Z"},
    {"id": 2, "name": "tool.exe", "browser_download_url": "https://x/tool.exe", "digest": "sha256:ABCDEF0123", "updated_at": "2026-08-19T23:48:38Z"}
  ]
}`

func TestTheLatestBuildOfTheNamedAssetIsFoundWithItsChecksum(t *testing.T) {
	server, client := githubReturning(t, 200, releaseJSON)

	info, err := lookupLatestFrom(server.URL, client, Source{Repo: "o/tool", Asset: "tool.exe"})

	if err != nil {
		t.Fatal(err)
	}
	if info.ID != 2 || info.Tag != "2026.08.19" || info.URL != "https://x/tool.exe" || info.SHA256 != "abcdef0123" {
		t.Errorf("info = %+v", info)
	}
	if info.UpdatedAt.Format("2006-01-02") != "2026-08-19" {
		t.Errorf("updated at = %v", info.UpdatedAt)
	}
}

func TestADigestInAnotherAlgorithmIsNotMistakenForASHA256(t *testing.T) {
	body := strings.Replace(releaseJSON, "sha256:ABCDEF0123", "sha512:ABCDEF0123", 1)
	server, client := githubReturning(t, 200, body)

	info, err := lookupLatestFrom(server.URL, client, Source{Repo: "o/tool", Asset: "tool.exe"})

	if err != nil || info.SHA256 != "" {
		t.Errorf("info = %+v, err = %v; a sha512 digest must not be compared as a sha256", info, err)
	}
}

func TestAReleaseWithoutThatAssetIsAnError(t *testing.T) {
	server, client := githubReturning(t, 200, releaseJSON)

	_, err := lookupLatestFrom(server.URL, client, Source{Repo: "o/tool", Asset: "missing.exe"})

	if err == nil || !strings.Contains(err.Error(), "missing.exe") {
		t.Errorf("err = %v", err)
	}
}

func TestRateLimitingIsExplainedInPlainWords(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusTooManyRequests} {
		server, client := githubReturning(t, status, `{"message":"API rate limit exceeded"}`)

		_, err := lookupLatestFrom(server.URL, client, Source{Repo: "o/tool", Asset: "tool.exe"})

		if err == nil || !strings.Contains(err.Error(), "limiting requests") {
			t.Errorf("HTTP %d: err = %v", status, err)
		}
	}
}

func TestOtherErrorsAndUnreadableAnswersAreErrors(t *testing.T) {
	server, client := githubReturning(t, 404, `{}`)
	if _, err := lookupLatestFrom(server.URL, client, Source{Repo: "o/tool", Asset: "tool.exe"}); err == nil {
		t.Error("a 404 should be an error")
	}

	server, client = githubReturning(t, 200, `not json`)
	if _, err := lookupLatestFrom(server.URL, client, Source{Repo: "o/tool", Asset: "tool.exe"}); err == nil {
		t.Error("an unreadable answer should be an error")
	}
}
