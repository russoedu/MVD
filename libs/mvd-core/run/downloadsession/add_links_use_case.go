package downloadsession

import (
	"fmt"
	"strings"

	"youtube-downloader/libs/mvd-core/wording"
)

// CanAdd reports whether the run behind the screen takes more links.
func (s Session) CanAdd() bool {
	_, ok := s.Ctrl.(SourceAdder)

	return ok
}

// BeginAdding opens the box for pasting links, when the run can take them. A hosted screen leaves
// adding to the workspace.
func (s *Session) BeginAdding() {
	if s.Hosted {
		return
	}
	if !s.CanAdd() {
		s.Say("this run cannot take more links")

		return
	}
	s.Adding, s.LinksText = true, ""
}

// SetLinksText says what is in the box for links.
func (s *Session) SetLinksText(text string) { s.LinksText = text }

// CancelAdding closes the box without adding anything.
func (s *Session) CancelAdding() { s.Adding = false }

// CommitLinks queues every line of the box on the run and closes the box.
func (s *Session) CommitLinks() {
	adder, ok := s.Ctrl.(SourceAdder)
	if !ok {
		s.Adding = false

		return
	}
	added, lines := 0, 0
	for _, line := range strings.Split(s.LinksText, "\n") {
		url := strings.TrimSpace(line)
		if url == "" {
			continue
		}
		lines++
		if _, queued := adder.AddSource(url); queued {
			added++
		}
	}
	s.Adding = false
	switch {
	case lines == 0:
		s.Say("no links to add")
	case added < lines:
		s.Say(fmt.Sprintf("added %d of %d links; the run has ended", added, lines))
	default:
		s.Say(fmt.Sprintf("added %d link%s", added, wording.Plural(added)))
		s.Follow = true
	}
}
