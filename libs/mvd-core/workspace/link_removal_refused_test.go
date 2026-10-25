package workspace

import (
	"testing"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
	downloadpage "youtube-downloader/libs/mvd-core/run/download"
)

// refusingRun is a fake run that will not take a link out.
type refusingRun struct{ *twoLinkRun }

func (refusingRun) RemoveSource(int) bool { return false }

func TestALinkThatTheRunWillNotTakeOutSaysSoAndStaysInTheList(t *testing.T) {
	run := refusingRun{&twoLinkRun{addingRun: &addingRun{fakeRun: newFakeRun()}}}
	list := downloadlist.New([]string{"https://youtu.be/a", "https://youtu.be/b"})
	m := sized(New(autoLoadHost(t, run), config.Config{OutputDir: t.TempDir()}, list))
	m = m.runStartedFor(t, run)

	next, _ := m.Update(downloadpage.ListEditMsg{Op: "remove", Index: 1})
	m = next.(Model)
	if m.notice == "" {
		t.Error("a link the run would not take out was dropped without a word")
	}
	if got := addressesOf(m); len(got) != 2 {
		t.Errorf("the list lost a link the run kept: %v", got)
	}
}
