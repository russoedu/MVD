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
	m.session.Cursor = 2
	before := m.session.Cfg.AutoRetry
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.session.Cfg.AutoRetry == before {
		t.Error("enter should toggle auto retry")
	}

	// Fragment count (item 1): edit, decrement, commit.
	m.session.Cursor = 1
	start := m.session.Cfg.ConcurrentFragments
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.session.Mode != settings.EditNumber {
		t.Fatalf("enter should edit the fragment count")
	}
	m, _, _, _ = m.Update(keys.Press("down"))
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.session.Cfg.ConcurrentFragments != start-1 {
		t.Errorf("fragments should decrement to %d, got %d", start-1, m.session.Cfg.ConcurrentFragments)
	}

	// Retry cycles (item 3) and the wait (item 4): edit and commit.
	for cursor, want := range map[int]int{3: 12, 4: 7, 5: 45} {
		m.session.Cursor = cursor
		m, _, _, _ = m.Update(keys.Press("enter"))
		m.input.SetValue(strconv.Itoa(want))
		m, _, _, _ = m.Update(keys.Press("enter"))
	}
	if m.session.Cfg.RetryCycles != 12 || m.session.Cfg.RetryWaitSeconds != 7 || m.session.Cfg.StartTimeoutSeconds != 45 {
		t.Errorf("retry settings not committed: %d cycles, %d s", m.session.Cfg.RetryCycles, m.session.Cfg.RetryWaitSeconds)
	}
	m.session.Cursor = 4
	m, _, _, _ = m.Update(keys.Press("enter"))
	m.input.SetValue("0")
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.session.Cfg.RetryWaitSeconds != 7 {
		t.Errorf("a wait of 0 should be refused, got %d", m.session.Cfg.RetryWaitSeconds)
	}
	if got := m.session.Display(4); !strings.Contains(got, "7s, 14s, 21s") {
		t.Errorf("the wait should show how it grows, got %q", got)
	}

	// Save returns the edited config.
	_, _, out, cfg := m.Update(keys.Press("s"))
	if out != Save || cfg.AutoRetry == before {
		t.Errorf("s should save the edited config")
	}
}
