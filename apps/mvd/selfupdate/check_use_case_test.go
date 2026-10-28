package selfupdate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"youtube-downloader/apps/mvd/notification"
	"youtube-downloader/apps/mvd/question"
)

const releasesBody = `[
 {"tag_name":"mvd-tui@0.9.0","html_url":"https://example.test/tui","draft":false,"prerelease":false,"assets":[]},
 {"tag_name":"mvd@0.7.0","html_url":"https://example.test/mvd-0.7.0","draft":false,"prerelease":false,
  "assets":[{"name":"mvd_0.7.0_windows_amd64.zip","browser_download_url":"https://example.test/a.zip","digest":"sha256:abcd"}]},
 {"tag_name":"mvd@0.6.1","html_url":"https://example.test/mvd-0.6.1","draft":false,"prerelease":false,"assets":[]}
]`

type harness struct {
	checker  *Checker
	notices  []notification.Notice
	asked    int
	opened   []string
	requests int
}

func newHarness(t *testing.T, version string, answer question.Answer, body string, status int) *harness {
	t.Helper()
	h := &harness{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		h.requests++
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	h.checker = &Checker{
		Dir: t.TempDir(), Version: version,
		Enabled: func() bool { return true },
		Notify:  func(n notification.Notice) { h.notices = append(h.notices, n) },
		Ask: func(string, string, []string) question.Answer {
			h.asked++

			return answer
		},
		Open: func(url string) error { h.opened = append(h.opened, url); return nil },
		Logf: func(string, ...interface{}) {},
		URL:  server.URL, Now: func() time.Time { return now },
	}
	return h
}

func TestAtStartTellsOnceAboutANewerVersion(t *testing.T) {
	h := newHarness(t, "0.6.1", question.AnswerLeave, releasesBody, http.StatusOK)
	h.checker.AtStart(context.Background())
	if len(h.notices) != 1 || !strings.Contains(h.notices[0].Title, "0.7.0") {
		t.Fatalf("notices = %+v, want one about 0.7.0 (the terminal app's 0.9.0 is not MVD)", h.notices)
	}
	if got := h.checker.Note(); !strings.Contains(got, "0.7.0") {
		t.Errorf("Note() = %q, want it to name 0.7.0", got)
	}

	// A day later it looks again but does not repeat itself.
	later := h.checker.now().Add(25 * time.Hour)
	h.checker.Now = func() time.Time { return later }
	h.checker.AtStart(context.Background())
	if len(h.notices) != 1 || h.requests != 2 {
		t.Errorf("notices = %d, requests = %d; want 1 and 2", len(h.notices), h.requests)
	}
}

func TestAtStartLooksOnlyOnceADayAndOnlyWhenAllowed(t *testing.T) {
	h := newHarness(t, "0.6.1", question.AnswerLeave, releasesBody, http.StatusOK)
	h.checker.AtStart(context.Background())
	h.checker.AtStart(context.Background())
	if h.requests != 1 {
		t.Errorf("requests = %d, want 1 within a day", h.requests)
	}

	off := newHarness(t, "0.6.1", question.AnswerLeave, releasesBody, http.StatusOK)
	off.checker.Enabled = func() bool { return false }
	off.checker.AtStart(context.Background())
	if off.requests != 0 || len(off.notices) != 0 {
		t.Errorf("switched off, but requests = %d, notices = %d", off.requests, len(off.notices))
	}
}

func TestAtStartIsSilentWhenOffline(t *testing.T) {
	h := newHarness(t, "0.6.1", question.AnswerLeave, "oops", http.StatusServiceUnavailable)
	h.checker.AtStart(context.Background())
	if len(h.notices) != 0 || h.checker.Note() != "" {
		t.Errorf("a failed look said something: %+v / %q", h.notices, h.checker.Note())
	}
	// The failure is not counted as a look, so the next start tries again.
	h.checker.AtStart(context.Background())
	if h.requests != 2 {
		t.Errorf("requests = %d, want 2", h.requests)
	}
}

func TestAtStartIgnoresDevelopmentBuilds(t *testing.T) {
	h := newHarness(t, "dev", question.AnswerLeave, releasesBody, http.StatusOK)
	h.checker.AtStart(context.Background())
	if len(h.notices) != 0 || h.checker.Note() != "" {
		t.Errorf("a development build was offered an update: %+v", h.notices)
	}
}

func TestCheckNowAsksAndOpensThePageOnYes(t *testing.T) {
	h := newHarness(t, "0.6.1", question.AnswerFirst, releasesBody, http.StatusOK)
	got := h.checker.CheckNow(context.Background())
	if h.asked != 1 || len(h.opened) != 1 || h.opened[0] != "https://example.test/mvd-0.7.0" {
		t.Errorf("asked = %d, opened = %v", h.asked, h.opened)
	}
	if !strings.Contains(got, "0.7.0") {
		t.Errorf("line = %q", got)
	}
}

func TestCheckNowNotNowOpensNothingAndStopsNotifying(t *testing.T) {
	h := newHarness(t, "0.6.1", question.AnswerSecond, releasesBody, http.StatusOK)
	h.checker.CheckNow(context.Background())
	if len(h.opened) != 0 {
		t.Errorf("opened %v after Not now", h.opened)
	}
	later := h.checker.now().Add(48 * time.Hour)
	h.checker.Now = func() time.Time { return later }
	h.checker.AtStart(context.Background())
	if len(h.notices) != 0 {
		t.Errorf("told again about a version the person already answered: %+v", h.notices)
	}
}

func TestCheckNowWithoutADialogOpensThePage(t *testing.T) {
	h := newHarness(t, "0.6.1", question.AnswerUnavailable, releasesBody, http.StatusOK)
	h.checker.CheckNow(context.Background())
	if len(h.opened) != 1 {
		t.Errorf("opened = %v, want the release page", h.opened)
	}
}

func TestCheckNowUpToDateAndFailures(t *testing.T) {
	h := newHarness(t, "0.7.0", question.AnswerFirst, releasesBody, http.StatusOK)
	if got := h.checker.CheckNow(context.Background()); !strings.Contains(got, "up to date") || h.asked != 0 {
		t.Errorf("line = %q, asked = %d", got, h.asked)
	}
	bad := newHarness(t, "0.6.1", question.AnswerFirst, "{", http.StatusOK)
	if got := bad.checker.CheckNow(context.Background()); !strings.Contains(got, "Could not look") {
		t.Errorf("line = %q", got)
	}
}
