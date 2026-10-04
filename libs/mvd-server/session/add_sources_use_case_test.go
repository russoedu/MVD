package session

import (
	"reflect"
	"sync"
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
)

func TestTheFirstURLsBuildTheEngineAndAreItsSources(t *testing.T) {
	s, fakes := newSession(t)

	result, err := s.Add([]string{"https://a.example/1", "https://b.example/2"})

	if err != nil {
		t.Fatal(err)
	}
	want := []string{"https://a.example/1", "https://b.example/2"}
	if !reflect.DeepEqual(result.Added, want) {
		t.Errorf("added %v, want %v", result.Added, want)
	}
	if !reflect.DeepEqual(fakes.calls, [][]string{want}) {
		t.Errorf("the factory was called with %v, want one call with %v", fakes.calls, want)
	}
	if got := len(s.Snapshot().Playlists); got != 2 {
		t.Errorf("snapshot has %d playlists, want 2", got)
	}
}

func TestLaterURLsJoinTheRunningEngineInsteadOfBuildingAnother(t *testing.T) {
	s, fakes := newSession(t)
	if _, err := s.Add([]string{"https://a.example/1"}); err != nil {
		t.Fatal(err)
	}

	result, err := s.Add([]string{"https://b.example/2"})

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Added, []string{"https://b.example/2"}) {
		t.Errorf("added %v", result.Added)
	}
	if len(fakes.calls) != 1 {
		t.Errorf("the factory was called %d times, want 1", len(fakes.calls))
	}
	if got := fakes.engine(t).sourceURLs(); !reflect.DeepEqual(got, []string{"https://a.example/1", "https://b.example/2"}) {
		t.Errorf("engine sources %v", got)
	}
	waitFor(t, "the added playlist in the snapshot", func() bool { return len(s.Snapshot().Playlists) == 2 })
}

func TestAddSaysWhatItDidWithEachLine(t *testing.T) {
	s, fakes := newSession(t)
	if _, err := s.Add([]string{"https://a.example/1"}); err != nil {
		t.Fatal(err)
	}

	result, err := s.Add([]string{"https://a.example/1", "not a url", "https://c.example/3"})

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Duplicates, []string{"https://a.example/1"}) {
		t.Errorf("duplicates %v", result.Duplicates)
	}
	if !reflect.DeepEqual(result.Rejected, []string{"not a url"}) {
		t.Errorf("rejected %v", result.Rejected)
	}
	if !reflect.DeepEqual(result.Added, []string{"https://c.example/3"}) {
		t.Errorf("added %v", result.Added)
	}
	if got := fakes.engine(t).sourceURLs(); !reflect.DeepEqual(got, []string{"https://a.example/1", "https://c.example/3"}) {
		t.Errorf("the duplicate reached the engine: %v", got)
	}
}

func TestAddOfNothingUsableDoesNotBuildAnEngine(t *testing.T) {
	s, fakes := newSession(t)

	result, err := s.Add([]string{"", "# a comment", "hello"})

	if err != nil || len(result.Added) != 0 || !reflect.DeepEqual(result.Rejected, []string{"hello"}) {
		t.Errorf("result %+v err %v", result, err)
	}
	if len(fakes.calls) != 0 {
		t.Error("an engine was built for nothing")
	}
}

func TestAddResultListsEncodeAsArraysEvenWhenEmpty(t *testing.T) {
	s, _ := newSession(t)

	result, _ := s.Add(nil)

	if result.Added == nil || result.Duplicates == nil || result.Rejected == nil {
		t.Errorf("a nil list would encode as null: %+v", result)
	}
}

func TestAPlaylistThatFailedToListCanBeAddedAgain(t *testing.T) {
	s, fakes := newSession(t)
	if _, err := s.Add([]string{"https://a.example/1"}); err != nil {
		t.Fatal(err)
	}
	fakes.engine(t).emit(engine.EvPlaylistFailed{Playlist: 0, Err: "network down"})
	waitFor(t, "the failure", func() bool { return s.Snapshot().Playlists[0].Err != "" })

	result, err := s.Add([]string{"https://a.example/1"})

	if err != nil || !reflect.DeepEqual(result.Added, []string{"https://a.example/1"}) || len(result.Duplicates) != 0 {
		t.Errorf("a failed playlist should be addable again: %+v err %v", result, err)
	}
}

func TestAFailedFactoryLeavesTheSessionUnstartedSoTheNextAddTriesAgain(t *testing.T) {
	s, fakes := newSession(t)
	fakes.fail = errBuild

	if _, err := s.Add([]string{"https://a.example/1"}); err != errBuild {
		t.Fatalf("Add = %v, want the factory's error", err)
	}
	if got := len(s.Snapshot().Playlists); got != 0 {
		t.Errorf("a failed start left %d playlists", got)
	}

	result, err := s.Add([]string{"https://a.example/1"})

	if err != nil || len(result.Added) != 1 {
		t.Errorf("the retry should have worked: %+v err %v", result, err)
	}
	if len(fakes.calls) != 2 {
		t.Errorf("the factory was called %d times, want 2", len(fakes.calls))
	}
}

func TestConcurrentFirstPastesBuildOneEngineAndLoseNoURL(t *testing.T) {
	s, fakes := newSession(t)
	const n = 12

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			url := "https://a.example/" + string(rune('a'+i))
			if _, err := s.Add([]string{url}); err != nil {
				t.Errorf("Add(%s): %v", url, err)
			}
		}(i)
	}
	wg.Wait()

	if len(fakes.calls) != 1 {
		t.Errorf("the factory was called %d times, want 1", len(fakes.calls))
	}
	if got := len(fakes.engine(t).Sources()); got != n {
		t.Errorf("the engine has %d sources, want %d", got, n)
	}
}

func TestAddFailsWhenTheEngineHasAlreadyFinished(t *testing.T) {
	s, fakes := newSession(t)
	if _, err := s.Add([]string{"https://a.example/1"}); err != nil {
		t.Fatal(err)
	}
	fakes.engine(t).refuse = true

	if _, err := s.Add([]string{"https://b.example/2"}); err != ErrClosed {
		t.Errorf("Add = %v, want ErrClosed", err)
	}
}

func TestRetryIsRefusedBeforeAnyEngineAndDelegatedAfter(t *testing.T) {
	s, fakes := newSession(t)
	if s.Retry(1) || s.RetryPlaylist(0) != 0 {
		t.Error("retry did something before any engine existed")
	}

	if _, err := s.Add([]string{"https://a.example/1"}); err != nil {
		t.Fatal(err)
	}
	fake := fakes.engine(t)

	if !s.Retry(1) || s.Retry(2) {
		t.Error("Retry did not pass the engine's answer through")
	}
	if got := s.RetryPlaylist(0); got != 2 {
		t.Errorf("RetryPlaylist = %d, want the engine's 2", got)
	}
	if !reflect.DeepEqual(fake.retried, []int{1, 2}) || !reflect.DeepEqual(fake.playlists, []int{0}) {
		t.Errorf("engine saw retries %v and playlist retries %v", fake.retried, fake.playlists)
	}
}
