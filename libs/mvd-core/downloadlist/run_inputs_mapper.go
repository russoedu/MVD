package downloadlist

import (
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlisteditor"
)

// Inputs is what a run is given to carry the list on: the addresses nobody has looked at
// yet, to be listed as usual, and the songs of the others that are not downloaded and not
// skipped, each with the video to download for it (as proposed, or as the person
// decided). Songs already downloaded are not given again.
func (s *Session) Inputs() (sources []string, plan []engine.PlannedEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.file.Items {
		if len(item.Songs) == 0 {
			sources = append(sources, item.Source)
			continue
		}
		planned, _ := playlisteditor.Selection(item.Songs, true)
		plan = append(plan, planned...)
	}
	return sources, plan
}
