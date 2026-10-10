package playlisteditor

import (
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlistfile"
)

// FromPlan turns what a plan-only run proposes into rows nobody has decided on yet.
func FromPlan(plan []engine.PlannedEntry) []playlistfile.Entry {
	rows := make([]playlistfile.Entry, 0, len(plan))
	for _, p := range plan {
		rows = append(rows, playlistfile.Entry{
			Playlist:    p.Playlist,
			PlaylistURL: p.PlaylistURL,
			Upload:      p.Upload,
			TargetID:    p.TargetID,
			Kind:        string(p.Kind),
			Reason:      p.Reason,
			Artist:      p.Artist,
			Title:       p.Title,
		})
	}
	return rows
}

// Merge brings what was decided and downloaded in saved onto the rows of a fresh plan, for
// the songs that are in both. Songs only in the fresh plan stay as proposed; songs only in
// saved are kept after them, so a file that is reopened loses nothing.
func Merge(fresh, saved []playlistfile.Entry) []playlistfile.Entry {
	byKey := make(map[string]playlistfile.Entry, len(saved))
	for _, s := range saved {
		byKey[s.Key()] = s
	}
	used := make(map[string]bool, len(fresh))
	out := make([]playlistfile.Entry, 0, len(fresh)+len(saved))
	for _, f := range fresh {
		key := f.Key()
		used[key] = true
		if s, ok := byKey[key]; ok {
			f.Decision, f.ChosenID, f.ChosenTitle, f.Downloaded = s.Decision, s.ChosenID, s.ChosenTitle, s.Downloaded
		}
		out = append(out, f)
	}
	for _, s := range saved {
		if !used[s.Key()] {
			out = append(out, s)
		}
	}
	return out
}
