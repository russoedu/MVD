package musicfixer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain lets the test binary double as the music-fixer app.
func TestMain(m *testing.M) {
	if os.Getenv("MVD_STUB_FIXER") == "1" {
		os.Exit(stubApp(os.Args[1:]))
	}
	os.Exit(m.Run())
}

func stubApp(args []string) int {
	out := ""
	for i, a := range args {
		if a == "--out" {
			out = args[i+1]
		}
	}
	switch os.Getenv("MVD_STUB_MODE") {
	case "fail":
		fmt.Println(`{"event":"video.finished","ok":false,"error":"no music found"}`)
		return 1
	case "wait": // runs until told to cancel
		fmt.Println(`{"event":"action.started","action":"Where"}`)
		var line string
		_, _ = fmt.Scanln(&line)
		return 1
	}
	fmt.Println(`{"event":"action.started","action":"Where"}`)
	_ = os.WriteFile(out, []byte(`[{"better":"v-better.mp4","offsetMs":1902,"score":535.5,"speed":1.001131}]`), 0o600)
	fmt.Println(`{"event":"run.finished","ok":true}`)
	return 0
}

func job(t *testing.T, mode string) Job {
	t.Helper()
	t.Setenv("MVD_STUB_FIXER", "1")
	t.Setenv("MVD_STUB_MODE", mode)
	return Job{Exe: os.Args[0], Video: "v.mkv", Music: "m.m4a", OutDir: filepath.Join(t.TempDir(), "out")}
}

func TestFixReturnsWhatTheAppDid(t *testing.T) {
	j := job(t, "")
	var actions []string

	got, err := Fix(context.Background(), j, func(a string) { actions = append(actions, a) })

	if err != nil || got.OffsetMs != 1902 || got.Speed != 1.001131 || got.Better != filepath.Join(j.OutDir, "v-better.mp4") {
		t.Fatalf("got %+v, %v", got, err)
	}
	if strings.Join(actions, ",") != "Where" {
		t.Errorf("actions: %v", actions)
	}
}

func TestFixReportsAFailedRun(t *testing.T) {
	if _, err := Fix(context.Background(), job(t, "fail"), nil); err == nil || !strings.Contains(err.Error(), "no music found") {
		t.Errorf("err: %v", err)
	}
}

func TestCancellingStopsTheApp(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	go func() { <-started; cancel() }()

	_, err := Fix(ctx, job(t, "wait"), func(string) { close(started) })

	if err != context.Canceled {
		t.Errorf("err: %v", err)
	}
}
