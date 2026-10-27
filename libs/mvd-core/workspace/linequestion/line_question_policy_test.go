package linequestion

import "testing"

func TestDecide(t *testing.T) {
	cases := []struct {
		key, value string
		outcome    Outcome
		text       string
	}{
		{"esc", "x", Cancelled, ""},
		{"enter", "  ", Empty, ""},
		{"enter", " a.mvd ", Answered, "a.mvd"},
		{"a", "x", Typing, ""},
	}
	for _, c := range cases {
		if o, text := Decide(c.key, c.value); o != c.outcome || text != c.text {
			t.Errorf("Decide(%q, %q) = %v, %q; want %v, %q", c.key, c.value, o, text, c.outcome, c.text)
		}
	}
}

func TestLinkPrompt(t *testing.T) {
	if LinkPrompt(true) != "Change the link" || LinkPrompt(false) != "Add a playlist or video link" {
		t.Fatal("the prompts changed")
	}
}
