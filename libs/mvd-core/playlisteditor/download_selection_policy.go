package playlisteditor

import (
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlistfile"
)

// Target is the video a row would download and what kind of video it is. ok is false when
// the row downloads nothing: it was skipped, or it has no video.
func Target(e playlistfile.Entry) (id string, kind engine.PlanKind, ok bool) {
	switch e.Decision {
	case playlistfile.DecisionSkipped:
		return "", "", false
	case playlistfile.DecisionReplaced:
		if e.ChosenID == "" {
			return "", "", false
		}
		return e.ChosenID, engine.KindChosen, true
	case playlistfile.DecisionOriginal:
		if e.Upload.ID == "" {
			return "", "", false
		}
		return e.Upload.ID, engine.KindOriginal, true
	}
	if e.TargetID == "" || engine.PlanKind(e.Kind) == engine.KindNone {
		return "", "", false
	}
	return e.TargetID, engine.PlanKind(e.Kind), true
}

// Revised reports whether the person has dealt with the row: confirmed it, replaced it,
// taken the original or skipped it.
func Revised(e playlistfile.Entry) bool { return e.Decision != playlistfile.DecisionNone }

// Pending reports whether a row still has to be downloaded if the person gets to it: it
// has not been downloaded and was not skipped.
func Pending(e playlistfile.Entry) bool {
	return !e.Downloaded && e.Decision != playlistfile.DecisionSkipped
}

// Unreviewed counts the rows nobody has dealt with that would still download something.
func Unreviewed(rows []playlistfile.Entry) int {
	n := 0
	for _, e := range rows {
		if _, _, ok := Target(e); ok && Pending(e) && !Revised(e) {
			n++
		}
	}
	return n
}

// Selection is what a download takes from the rows: the songs the person revised, and when
// includeUnreviewed is true the ones they did not look at as well, as they were proposed.
// Songs already downloaded are never taken again. keys[i] is the Key of rows' song that
// planned[i] came from, so a caller can tell which one a download finished.
func Selection(rows []playlistfile.Entry, includeUnreviewed bool) (planned []engine.PlannedEntry, keys []string) {
	for _, e := range rows {
		if !Pending(e) {
			continue
		}
		if !Revised(e) && !includeUnreviewed {
			continue
		}
		id, kind, ok := Target(e)
		if !ok {
			continue
		}
		reason := e.Reason
		if kind == engine.KindChosen {
			reason = "chosen by the person"
		}
		planned = append(planned, engine.PlannedEntry{
			Playlist:    e.Playlist,
			PlaylistURL: e.PlaylistURL,
			Upload:      e.Upload,
			TargetID:    id,
			Kind:        kind,
			Reason:      reason,
			Artist:      e.Artist,
			Title:       e.Title,
		})
		keys = append(keys, e.Key())
	}
	return planned, keys
}

// MarkDownloaded sets Downloaded on the rows with the given keys, and says how many it set.
func MarkDownloaded(rows []playlistfile.Entry, keys ...string) int {
	want := make(map[string]bool, len(keys))
	for _, k := range keys {
		want[k] = true
	}
	n := 0
	for i := range rows {
		if want[rows[i].Key()] && !rows[i].Downloaded {
			rows[i].Downloaded = true
			n++
		}
	}
	return n
}
