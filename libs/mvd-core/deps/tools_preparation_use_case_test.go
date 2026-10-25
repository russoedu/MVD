package deps

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWhenNothingIsMissingThePreparationIsReadyAtOnceAndAsksNothing(t *testing.T) {
	w := newWorld(t)
	i, _ := newTestInstaller(t, t.TempDir(), w, map[string]bool{"deno": true})
	first := prepare(i, nil)
	if err := first.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	// With the tools installed, a second preparation has nothing to fetch.
	second, _ := newTestInstaller(t, i.binDir, w, map[string]bool{"deno": true})
	fetchedBefore := len(w.fetched)
	again := prepare(second, nil)
	if !again.Ready() || again.Status() != "" {
		t.Errorf("ready %v, status %q", again.Ready(), again.Status())
	}
	if len(w.fetched) != fetchedBefore {
		t.Errorf("an installed set of tools still fetched %v", w.fetched[fetchedBefore:])
	}
}

func TestOnTheFirstStartTheToolsAreFetchedInTheBackgroundWithAStatus(t *testing.T) {
	w := newWorld(t)
	release := make(chan struct{})
	i, _ := newTestInstaller(t, t.TempDir(), w, nil)
	slow := i.fetch
	i.fetch = func(url, dest string) (string, error) {
		<-release // the download waits until the test lets it go
		return slow(url, dest)
	}
	p := prepare(i, nil)
	if p.Ready() {
		t.Fatal("ready before anything was fetched")
	}
	// The status says what is being got: the tools as a whole until the first one begins, then it.
	if status := p.Status(); !strings.Contains(status, "Getting") {
		t.Errorf("status %q", status)
	}
	close(release)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := p.Wait(ctx); err != nil {
		t.Fatalf("never finished: %v", err)
	}
	if !p.Ready() || p.Status() != "" || p.Problem() != "" {
		t.Errorf("ready %v, status %q, problem %q", p.Ready(), p.Status(), p.Problem())
	}
	if !strings.HasSuffix(ToolPath(i.binDir, "yt-dlp"), exe("yt-dlp")) {
		t.Errorf("ToolPath says %q", ToolPath(i.binDir, "yt-dlp"))
	}
	if got := read(t, filepath.Join(i.binDir, exe("yt-dlp"))); got == "" {
		t.Error("yt-dlp was not installed")
	}
}

func TestAToolThatCannotBeGottenLeavesThePreparationNotReadyWithTheReason(t *testing.T) {
	w := newWorld(t)
	w.fetchErr[w.releases["yt-dlp"].URL] = errors.New("no connection")
	i, _ := newTestInstaller(t, t.TempDir(), w, map[string]bool{"deno": true})
	p := prepare(i, nil)
	if err := p.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if p.Ready() {
		t.Error("it is ready without yt-dlp")
	}
	if got := p.Problem(); !strings.Contains(got, "Could not get yt-dlp") || !strings.Contains(got, "no connection") {
		t.Errorf("the problem is %q", got)
	}
}

func TestWaitStopsWhenTheContextEnds(t *testing.T) {
	w := newWorld(t)
	block := make(chan struct{})
	defer close(block)
	i, _ := newTestInstaller(t, t.TempDir(), w, nil)
	i.fetch = func(string, string) (string, error) { <-block; return "", errors.New("stopped") }
	p := prepare(i, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := p.Wait(ctx); err == nil {
		t.Error("Wait returned though the tools were not ready")
	}
}
