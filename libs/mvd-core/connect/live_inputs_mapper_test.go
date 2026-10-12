package connect

import "testing"

func TestPlainStripsColourCodesFromTextsAndRows(t *testing.T) {
	type row struct{ Title string }
	got := Plain(map[string]any{
		"Counters": "\x1b[31mQueue 1\x1b[0m",
		"N":        3,
		"Rows":     []row{{Title: "\x1b[1mSong\x1b[0m"}},
	})
	if got["Counters"] != "Queue 1" {
		t.Errorf("counters %q", got["Counters"])
	}
	rows := got["Rows"].([]any)
	if rows[0].(map[string]any)["Title"] != "Song" {
		t.Errorf("rows %v", rows)
	}
	if got["N"] != float64(3) {
		t.Errorf("n %v", got["N"])
	}
}
