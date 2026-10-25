package download

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestTheCountersAreInEverydayWordsAndTheQuietOnesOnlyHaveAPlaceWhenTheyHaveAValue(t *testing.T) {
	m := sampleModel(t) // one song is official, one is a duplicate, one failed
	counters, _ := m.headerTexts()
	plain := ansi.Strip(counters)
	for _, want := range []string{"Waiting 1", "Downloading 3", "Done 1", "Failed 2", "Official video 1", "Already had 1"} {
		if !strings.Contains(plain, want) {
			t.Errorf("the counters lack %q: %q", want, plain)
		}
	}
	for _, jargon := range []string{"Queue", "Running", "Dup ", "Better"} {
		if strings.Contains(plain, jargon) {
			t.Errorf("the counters still say %q: %q", jargon, plain)
		}
	}
}

func TestACounterAtZeroLeavesRoomForTheOnesThatMatter(t *testing.T) {
	m := sampleModel(t)
	m.state = New(&fakeController{events: make(chan interface{})}).state
	counters, _ := m.headerTexts()
	plain := ansi.Strip(counters)
	if strings.Contains(plain, "Official") || strings.Contains(plain, "Already had") {
		t.Errorf("a counter at zero is shown: %q", plain)
	}
	if !strings.Contains(plain, "Failed 0") {
		t.Errorf("Failed is not always there: %q", plain)
	}
	if narrow := ansi.Truncate(counters, 50, ""); !strings.Contains(ansi.Strip(narrow), "Failed") {
		t.Errorf("Failed is cut in a narrow window: %q", ansi.Strip(narrow))
	}
}

func TestWhileTheDownloadsAreHeldTheHeaderSaysStopped(t *testing.T) {
	m := sampleModel(t)
	if _, right := m.headerTexts(); strings.Contains(right, "STOPPED") {
		t.Errorf("not stopped, but it says %q", right)
	}
	if _, right := m.WithPaused(true).headerTexts(); !strings.Contains(right, "STOPPED") {
		t.Errorf("stopped, but it says %q", right)
	}
}
