package download

import (
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
)

func footerPercent(t *testing.T, m Model) float64 {
	t.Helper()
	f := m.outputFooter()
	if f.bar == nil {
		t.Fatal("no progress bar")
	}

	return f.bar.percent
}

func TestTheBarIsEmptyWhileTheSizeIsNotKnown(t *testing.T) {
	m := sampleModel(t)
	m.focus = paneEntries
	m.selEntry = 4
	en := m.selectedEntry()
	en.State, en.Total, en.Percent = engine.StateDownloading, 0, 0
	if got := footerPercent(t, m); got != 0 {
		t.Errorf("a download of unknown size shows %v%%, want an empty bar", got)
	}
}

func TestTheBarFollowsTheDownloadOnceTheSizeIsKnown(t *testing.T) {
	m := sampleModel(t)
	m.focus = paneEntries
	m.selEntry = 4
	en := m.selectedEntry()
	en.State, en.Total, en.Percent = engine.StateDownloading, 1000, 34.2
	if got := footerPercent(t, m); got != 34.2 {
		t.Errorf("the bar shows %v%%, want 34.2", got)
	}
}

func TestTheBarIsFullWhileTheFilesAreBeingJoined(t *testing.T) {
	m := sampleModel(t)
	m.focus = paneEntries
	m.selEntry = 5
	en := m.selectedEntry()
	en.State, en.Total = engine.StateMerging, 0
	if got := footerPercent(t, m); got != 100 {
		t.Errorf("the merge shows %v%%, want a full bar", got)
	}
}
