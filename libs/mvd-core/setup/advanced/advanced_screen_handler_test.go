package advanced

import (
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

	// Save returns the edited config.
	_, _, out, cfg := m.Update(keys.Press("s"))
	if out != Save || cfg.AutoRetry == before {
		t.Errorf("s should save the edited config")
	}
}
