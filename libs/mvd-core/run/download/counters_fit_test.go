package download

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestInANarrowWindowTheCountersThatAreZeroGoBeforeFailedIsCut(t *testing.T) {
	m := sampleModel(t)
	wide, _ := m.headerTexts(200)
	for _, want := range []string{"Official", "Better", "Dup", "Failed"} {
		if !strings.Contains(ansi.Strip(wide), want) {
			t.Errorf("a wide window lacks %q: %q", want, ansi.Strip(wide))
		}
	}
	narrow, _ := m.headerTexts(ansi.StringWidth(wide) - 12)
	plain := ansi.Strip(narrow)
	if !strings.Contains(plain, "Failed") {
		t.Errorf("Failed was lost: %q", plain)
	}
	if strings.Contains(plain, "Better") {
		t.Errorf("a counter at zero should have gone first: %q", plain)
	}
}

func TestCountersThatSayingSomethingAreKeptEvenWhenItIsTight(t *testing.T) {
	m := sampleModel(t) // one song is official, one is a duplicate
	narrow, _ := m.headerTexts(60)
	plain := ansi.Strip(narrow)
	if !strings.Contains(plain, "Official 1") || !strings.Contains(plain, "Dup 1") {
		t.Errorf("a counter with something to say went: %q", plain)
	}
}
