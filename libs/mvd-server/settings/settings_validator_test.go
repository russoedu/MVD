package settings

import (
	"strings"
	"testing"
)

func valid(t *testing.T) Settings {
	t.Helper()
	return Settings{
		OutputDir: t.TempDir(), VideoQuality: "best", AudioQuality: "best", MergeOutputFormat: "mp4",
		OutputTemplate: "%(title)s.%(ext)s", MaxConcurrentDownloads: 4, ConcurrentFragments: 4,
		AutoRetry: true, Cookies: "all", CreateLogFile: true, LogDir: t.TempDir(),
	}
}

func TestAnOrdinaryConfigurationIsValid(t *testing.T) {
	if bad := Validate(valid(t)); bad != nil {
		t.Errorf("rejected: %v", bad)
	}
}

func TestEachFieldIsCheckedAndNamedByItsJSONKey(t *testing.T) {
	cases := map[string]func(*Settings){
		"outputDir":               func(s *Settings) { s.OutputDir = "" },
		"outputDir ":              func(s *Settings) { s.OutputDir = "relative/path" },
		"outputDir  ":             func(s *Settings) { s.OutputDir = "/a\nb" },
		"logDir":                  func(s *Settings) { s.LogDir = "logs" },
		"videoQuality":            func(s *Settings) { s.VideoQuality = "8k" },
		"audioQuality":            func(s *Settings) { s.AudioQuality = "loud" },
		"mergeOutputFormat":       func(s *Settings) { s.MergeOutputFormat = "avi" },
		"outputTemplate":          func(s *Settings) { s.OutputTemplate = "  " },
		"outputTemplate ":         func(s *Settings) { s.OutputTemplate = "a\nb" },
		"maxConcurrentDownloads":  func(s *Settings) { s.MaxConcurrentDownloads = 0 },
		"maxConcurrentDownloads ": func(s *Settings) { s.MaxConcurrentDownloads = 33 },
		"concurrentFragments":     func(s *Settings) { s.ConcurrentFragments = -1 },
		"concurrentFragments ":    func(s *Settings) { s.ConcurrentFragments = 33 },
		"cookies":                 func(s *Settings) { s.Cookies = "" },
		"cookies ":                func(s *Settings) { s.Cookies = "--exec" },
		"cookies  ":               func(s *Settings) { s.Cookies = "Firefox" },
		"cookies   ":              func(s *Settings) { s.Cookies = "firefox:a\nb" },
	}
	for name, mutate := range cases {
		s := valid(t)
		mutate(&s)

		bad := Validate(s)

		field := strings.TrimSpace(name)
		if _, ok := bad[field]; !ok {
			t.Errorf("%q: %q was not reported (got %v)", name, field, bad)
		}
		if len(bad) != 1 {
			t.Errorf("%q: reported more than the one field: %v", name, bad)
		}
	}
}

func TestBrowserSpecsWithAProfileAreAccepted(t *testing.T) {
	for _, cookies := range []string{"firefox", "chrome", "firefox:default", "chrome:Profile 1", "all", "off"} {
		s := valid(t)
		s.Cookies = cookies
		if bad := Validate(s); bad != nil {
			t.Errorf("%q rejected: %v", cookies, bad)
		}
	}
}

func TestTheLogFolderIsOnlyCheckedWhenALogIsKept(t *testing.T) {
	s := valid(t)
	s.CreateLogFile = false
	s.LogDir = ""
	if bad := Validate(s); bad != nil {
		t.Errorf("rejected: %v", bad)
	}
}

func TestEveryProblemIsReportedAtOnce(t *testing.T) {
	s := valid(t)
	s.OutputDir = ""
	s.VideoQuality = "8k"
	s.Cookies = "-x"

	if bad := Validate(s); len(bad) != 3 {
		t.Errorf("reported %d of 3: %v", len(bad), bad)
	}
}
