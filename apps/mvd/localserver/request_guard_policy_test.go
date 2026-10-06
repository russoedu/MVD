package localserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func request(method, host string, headers map[string]string) *http.Request {
	r := httptest.NewRequest(method, "http://"+host+"/api/state", strings.NewReader(""))
	r.Host = host
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	return r
}

func TestLoopbackNamesAreAnsweredWithOrWithoutAPort(t *testing.T) {
	for _, host := range []string{"127.0.0.1:8080", "127.0.0.1", "localhost:3000", "LOCALHOST:3000", "[::1]:8080", "[::1]"} {
		if got := refusal(request(http.MethodGet, host, nil)); got != "" {
			t.Errorf("%s was refused: %s", host, got)
		}
	}
}

func TestAnyOtherHostIsRefusedWhichIsWhatEndsDNSRebinding(t *testing.T) {
	for _, host := range []string{"evil.example:8080", "127.0.0.1.evil.example", "localhost.evil.example:8080", "192.168.1.5:8080", "0.0.0.0:8080"} {
		if got := refusal(request(http.MethodGet, host, nil)); got == "" {
			t.Errorf("%s was let through", host)
		}
	}
}

func TestAPageOnAnotherSiteIsRefused(t *testing.T) {
	cases := map[string]map[string]string{
		"a foreign Origin":      {"Origin": "https://evil.example"},
		"a loopback look-alike": {"Origin": "https://localhost.evil.example"},
		"cross-site fetch":      {"Sec-Fetch-Site": "cross-site"},
		"a non-http Origin":     {"Origin": "file://"},
		"an unparseable Origin": {"Origin": "://"},
		"null origin (sandbox)": {"Origin": "null"},
	}
	for name, headers := range cases {
		if got := refusal(request(http.MethodGet, "127.0.0.1:8080", headers)); got == "" {
			t.Errorf("%s was let through", name)
		}
	}
}

func TestOurOwnPagesAreLetThrough(t *testing.T) {
	cases := []map[string]string{
		{"Origin": "http://127.0.0.1:8080"},
		{"Origin": "http://localhost:5173"},
		{"Origin": "http://[::1]:8080"},
		{"Sec-Fetch-Site": "same-origin"},
		{"Sec-Fetch-Site": "same-site"},
		{"Sec-Fetch-Site": "none"},
	}
	for _, headers := range cases {
		if got := refusal(request(http.MethodGet, "127.0.0.1:8080", headers)); got != "" {
			t.Errorf("%v was refused: %s", headers, got)
		}
	}
}

func TestAWriteMustBeJSON(t *testing.T) {
	for _, contentType := range []string{"", "text/plain", "application/x-www-form-urlencoded", "multipart/form-data; boundary=x", "application/jsonp"} {
		if got := refusal(request(http.MethodPost, "127.0.0.1:8080", map[string]string{"Content-Type": contentType})); got == "" {
			t.Errorf("a POST as %q was let through", contentType)
		}
	}
	for _, contentType := range []string{"application/json", "application/json; charset=utf-8", "Application/JSON"} {
		if got := refusal(request(http.MethodPost, "127.0.0.1:8080", map[string]string{"Content-Type": contentType})); got != "" {
			t.Errorf("a POST as %q was refused: %s", contentType, got)
		}
	}
}

func TestAReadNeedsNoContentType(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		if got := refusal(request(method, "127.0.0.1:8080", nil)); got != "" {
			t.Errorf("%s was refused: %s", method, got)
		}
	}
}

func TestEveryOtherMethodThatWritesIsHeldToTheSameRule(t *testing.T) {
	for _, method := range []string{http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodPost} {
		if got := refusal(request(method, "127.0.0.1:8080", nil)); got == "" {
			t.Errorf("%s without JSON was let through", method)
		}
	}
}
