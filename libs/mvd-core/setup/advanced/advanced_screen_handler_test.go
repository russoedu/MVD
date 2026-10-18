package advanced

import (
	"strconv"
	"strings"
	"testing"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/settings"
	"youtube-downloader/libs/mvd-core/testing/keys"
)

func TestAdvancedModel(t *testing.T) {
	m := New(config.Default("/app", "/dl"))

	// Toggle auto retry (item 2).
	m.cursor = 2
	before := m.cfg.AutoRetry
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.cfg.AutoRetry == before {
		t.Error("enter should toggle auto retry")
	}

	// Fragment count (item 1): edit, decrement, commit.
	m.cursor = 1
	start := m.cfg.ConcurrentFragments
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.mode != settings.EditNumber {
		t.Fatalf("enter should edit the fragment count")
	}
	m, _, _, _ = m.Update(keys.Press("down"))
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.cfg.ConcurrentFragments != start-1 {
		t.Errorf("fragments should decrement to %d, got %d", start-1, m.cfg.ConcurrentFragments)
	}

	// Retry cycles (item 3) and the wait (item 4): edit and commit.
	for cursor, want := range map[int]int{3: 12, 4: 7} {
		m.cursor = cursor
		m, _, _, _ = m.Update(keys.Press("enter"))
		m.input.SetValue(strconv.Itoa(want))
		m, _, _, _ = m.Update(keys.Press("enter"))
	}
	if m.cfg.RetryCycles != 12 || m.cfg.RetryWaitSeconds != 7 {
		t.Errorf("retry settings not committed: %d cycles, %d s", m.cfg.RetryCycles, m.cfg.RetryWaitSeconds)
	}
	m.cursor = 4
	m, _, _, _ = m.Update(keys.Press("enter"))
	m.input.SetValue("0")
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.cfg.RetryWaitSeconds != 7 {
		t.Errorf("a wait of 0 should be refused, got %d", m.cfg.RetryWaitSeconds)
	}
	if got := m.display(4); !strings.Contains(got, "7s, 14s, 21s") {
		t.Errorf("the wait should show how it grows, got %q", got)
	}

	// Save returns the edited config.
	_, _, out, cfg := m.Update(keys.Press("s"))
	if out != Save || cfg.AutoRetry == before {
		t.Errorf("s should save the edited config")
	}
}
