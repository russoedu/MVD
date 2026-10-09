package main

import (
	"strings"
	"testing"
)

func TestTheReportSaysWhatChangedSongBySong(t *testing.T) {
	verdict := Verdict{
		Known: 50, Failed: true, Reason: "more of the songs than allowed now give a different answer",
		Different: []Outcome{{ID: "abc", Title: "A | B", Expected: "v1", Found: "", Why: "no link\nin the page"}},
	}

	report := reportOf("https://playlist", verdict, 0.1)

	for _, want := range []string{"CHANGED", "https://playlist", "| A / B (`abc`) | `v1` | none | no link in the page |", "up to 10%"} {
		if !strings.Contains(report, want) {
			t.Errorf("the report should contain %q:\n%s", want, report)
		}
	}
}

func TestAPassingReportSaysTheAnswersStillMatch(t *testing.T) {
	if report := reportOf("p", Verdict{Known: 50}, 0.1); !strings.Contains(report, "still match") {
		t.Errorf("report: %s", report)
	}
}
