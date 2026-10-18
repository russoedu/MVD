package fetch

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestMain lets the test binary stand in for yt-dlp: it does nothing for a long time.
func TestMain(m *testing.M) {
	if os.Getenv("MVD_STUB_SLEEP") == "1" {
		time.Sleep(time.Minute)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestDownloadGivesUpWhenItNeverStarts(t *testing.T) {
	if _, err := exec.LookPath(os.Args[0]); err != nil && !strings.Contains(os.Args[0], "fetch") {
		t.Skip("cannot run the test binary")
	}
	t.Setenv("MVD_STUB_SLEEP", "1")
	began := time.Now()
	err := DownloadWithStartTimeout(context.Background(), os.Args[0], nil, 300*time.Millisecond, func(string) {})
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("want a timed out error, got %v", err)
	}
	if time.Since(began) > 10*time.Second {
		t.Fatalf("took %s to give up", time.Since(began))
	}
}
