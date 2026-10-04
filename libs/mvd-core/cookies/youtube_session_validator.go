package cookies

import (
	"os"
	"strings"
)

// sessionCookieNames are cookies that indicate a logged-in Google/YouTube
// session. Any one of them, on a google.com or youtube.com domain with a
// value, means the export carries a usable login.
var sessionCookieNames = map[string]bool{
	"SID":            true,
	"SSID":           true,
	"SAPISID":        true,
	"APISID":         true,
	"__Secure-1PSID": true,
	"__Secure-3PSID": true,
	"LOGIN_INFO":     true,
}

// hasYouTubeSession reports whether a Netscape cookie file contains a live
// YouTube/Google login cookie.
func hasYouTubeSession(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		line = strings.TrimPrefix(line, "#HttpOnly_")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			continue
		}
		domain := strings.ToLower(fields[0])
		if !strings.Contains(domain, "youtube.com") && !strings.Contains(domain, "google.com") {
			continue
		}
		if sessionCookieNames[fields[5]] && fields[6] != "" {
			return true
		}
	}
	return false
}
