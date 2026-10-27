package downloadlist

import (
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/playlisteditor"
)

// Inputs is what a run is given to carry the list on: the addresses nobody has looked at
// yet, to be listed as usual, and the songs of the others that are not downloaded and not
// skipped, each with the video to download for it (as proposed, or as the person
// decided). A song whose video was not found yet, because a run stopped before it got to it,
// is given without one, to be looked up: the songs a source was filled with are kept, and its
// listing or search is not done again. Songs already downloaded are not given again.
func (s *Session) Inputs() (sources []string, plan []engine.PlannedEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.file.Items {
		if len(item.Songs) == 0 {
			sources = append(sources, item.Source)
			continue
		}
		for _, song := range item.Songs {
			if planned, _ := playlisteditor.Selection([]playlistfile.Entry{song}, true); len(planned) == 1 {
				plan = append(plan, planned[0])
			} else if notLookedUp(song) {
				plan = append(plan, engine.PlannedEntry{
					Playlist: song.Playlist, PlaylistURL: song.PlaylistURL, Upload: song.Upload,
					Artist: song.Artist, Title: song.Title,
				})
			}
		}
	}
	return sources, plan
}

// notLookedUp says whether a song is still to be downloaded and nobody has found its video
// yet: no video, no decision, and not found to have none.
func notLookedUp(song playlistfile.Entry) bool {
	return playlisteditor.Pending(song) && !playlisteditor.Revised(song) && song.TargetID == "" &&
		song.State != playlistfile.StateNotFound && engine.PlanKind(song.Kind) != engine.KindNone
}
