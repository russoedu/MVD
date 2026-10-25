package downloadsnapshot

import "youtube-downloader/libs/mvd-core/run/downloadsession"

// HintsOf are the keys the download screen offers now.
func HintsOf(s downloadsession.Session) []Hint {
	var hints []Hint
	if s.State.Idle {
		hints = []Hint{{"↑↓", "move"}, {"tab", "pane"}, {"f", "failed"}, {"r", "retry failed"}, {"t", "technical log"}}
	} else {
		hints = []Hint{{"↑↓", "move"}, {"tab", "pane"}, {"⏎", "follow"}, {"f", "failed"}, {"r", "retry"}, {"t", "technical log"}}
	}
	if s.CanChoose() && (s.Pane == downloadsession.PaneEntries || len(s.Picked) > 0) && len(s.TargetIDs()) > 0 {
		hints = append([]Hint{{"space", "select"}, {"d", "accept"}, {"x", "reject"}, {"o", "open video"}, {"D", "accept all"}}, append(hints[:2:2], Hint{"t", "technical log"})...)
	}
	if s.Technical {
		// The log of the tool is not the screen people expect: the way back is always on the bar.
		kept := hints[:0:0]
		for _, h := range hints {
			if h.Key != "t" {
				kept = append(kept, h)
			}
		}
		hints = append(kept, Hint{"t", "back to the details"})
	}
	if s.Hosted {
		return hints
	}
	if s.CanAdd() {
		hints = append(hints, Hint{"a", "add links"})
	}
	if !s.State.Idle {
		hints = append(hints, Hint{"?", "help"})
	}

	return append(hints, Hint{"q", "quit"})
}
