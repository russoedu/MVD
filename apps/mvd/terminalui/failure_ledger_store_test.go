package terminalui

import (
	"sort"
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
)

type fakeRetrier struct {
	retried []int
	added   []string
	ended   bool
}

func (f *fakeRetrier) Retry(id int) bool {
	if f.ended {
		return false
	}
	f.retried = append(f.retried, id)
	return true
}

func (f *fakeRetrier) AddSource(url string) (engine.PlaylistSource, bool) {
	if f.ended {
		return engine.PlaylistSource{}, false
	}
	f.added = append(f.added, url)
	return engine.PlaylistSource{Index: 9, URL: url}, true
}

func ledgerWithLinks() *failureLedger {
	links := []string{"https://a", "https://b"}
	return newFailureLedger(func(i int) string { return links[i] })
}

func TestTheLedgerReportsFailuresAndRetriesThem(t *testing.T) {
	ledger := ledgerWithLinks()

	if ledger.apply(engine.EvEntryState{Entry: 3, State: engine.StateDownloading}) {
		t.Error("a download starting is not a failure")
	}
	if !ledger.apply(engine.EvEntryState{Entry: 3, State: engine.StateFailed}) ||
		!ledger.apply(engine.EvEntryState{Entry: 5, State: engine.StateFailed}) ||
		!ledger.apply(engine.EvPlaylistFailed{Playlist: 1, Err: "unable to recognize"}) {
		t.Fatal("failed entries and links that could not be looked up are failures")
	}

	retrier := &fakeRetrier{}
	if got := ledger.retryAll(retrier); got != 3 {
		t.Errorf("queued %d, want 3", got)
	}
	sort.Ints(retrier.retried)
	if len(retrier.retried) != 2 || retrier.retried[0] != 3 || retrier.retried[1] != 5 {
		t.Errorf("retried entries %v", retrier.retried)
	}
	if len(retrier.added) != 1 || retrier.added[0] != "https://b" {
		t.Errorf("links added again %v", retrier.added)
	}
}

func TestAnEntryThatRecoveredOrALinkThatWasListedIsNotRetried(t *testing.T) {
	ledger := ledgerWithLinks()
	ledger.apply(engine.EvEntryState{Entry: 3, State: engine.StateFailed})
	ledger.apply(engine.EvEntryState{Entry: 3, State: engine.StateQueued}) // retried by the user
	ledger.apply(engine.EvPlaylistFailed{Playlist: 0, Err: "x"})
	ledger.apply(engine.EvPlaylistListed{Playlist: 0, Title: "A"}) // worked the second time

	retrier := &fakeRetrier{}
	if got := ledger.retryAll(retrier); got != 0 || len(retrier.retried)+len(retrier.added) != 0 {
		t.Errorf("queued %d (%v %v), want nothing", got, retrier.retried, retrier.added)
	}
}

func TestALinkIsAddedAgainOnlyOnce(t *testing.T) {
	ledger := ledgerWithLinks()
	ledger.apply(engine.EvPlaylistFailed{Playlist: 0, Err: "x"})

	retrier := &fakeRetrier{}
	ledger.retryAll(retrier)
	if got := ledger.retryAll(retrier); got != 0 || len(retrier.added) != 1 {
		t.Errorf("second retry queued %d, links added %v", got, retrier.added)
	}
}

func TestNothingIsCountedWhenTheRunHasEnded(t *testing.T) {
	ledger := ledgerWithLinks()
	ledger.apply(engine.EvEntryState{Entry: 1, State: engine.StateFailed})
	ledger.apply(engine.EvPlaylistFailed{Playlist: 0, Err: "x"})

	if got := ledger.retryAll(&fakeRetrier{ended: true}); got != 0 {
		t.Errorf("queued %d on an ended run", got)
	}
}
