package listhistory

import "youtube-downloader/libs/mvd-core/playlistfile"

// Clone is a copy of items that shares nothing with them.
func Clone(items []playlistfile.Item) []playlistfile.Item {
	out := make([]playlistfile.Item, len(items))
	for i, it := range items {
		out[i] = playlistfile.Item{Source: it.Source, Songs: append([]playlistfile.Entry(nil), it.Songs...)}
	}

	return out
}

// History is the lists as they were before the changes, and the ones taken back, most recent last.
// It is a value: a change gives a new history.
type History struct {
	undone, redone [][]playlistfile.Item
}

// CanUndo says there is a change to take back.
func (h History) CanUndo() bool { return len(h.undone) > 0 }

// CanRedo says there is a change taken back that can be made again.
func (h History) CanRedo() bool { return len(h.redone) > 0 }

// Recorded is the history after a change to current: it can be taken back, and what was taken back
// before cannot be made again.
func (h History) Recorded(current []playlistfile.Item) History {
	return History{undone: append(append([][]playlistfile.Item(nil), h.undone...), Clone(current))}
}

// Undone takes back the last change: it gives the list as it was before. ok is false when there is
// nothing to take back.
func (h History) Undone(current []playlistfile.Item) (next History, items []playlistfile.Item, ok bool) {
	if len(h.undone) == 0 {
		return h, nil, false
	}
	last := h.undone[len(h.undone)-1]
	next = History{
		undone: append([][]playlistfile.Item(nil), h.undone[:len(h.undone)-1]...),
		redone: append(append([][]playlistfile.Item(nil), h.redone...), Clone(current)),
	}

	return next, last, true
}

// Redone makes again the change that was taken back last.
func (h History) Redone(current []playlistfile.Item) (next History, items []playlistfile.Item, ok bool) {
	if len(h.redone) == 0 {
		return h, nil, false
	}
	last := h.redone[len(h.redone)-1]
	next = History{
		undone: append(append([][]playlistfile.Item(nil), h.undone...), Clone(current)),
		redone: append([][]playlistfile.Item(nil), h.redone[:len(h.redone)-1]...),
	}

	return next, last, true
}
