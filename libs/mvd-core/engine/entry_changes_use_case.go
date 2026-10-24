package engine

import (
	"context"
	"os"
	"time"

	"youtube-downloader/libs/mvd-core/ytdlp"
)

// isRemoved says whether the person took a playlist out of the run.
func (e *Engine) isRemoved(playlist int) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.removed[playlist]
}

// dropRemoved ends an entry of a playlist that was taken out while the entry waited for a stage.
func (e *Engine) dropRemoved(en *engineEntry) {
	e.setState(en, StateSkipped, "")
	e.finishOne()
}

// RemoveSource takes a playlist out of the run: what has not begun to download is dropped, a
// download that is under way finishes, and the playlist no longer shows in Snapshot. Renderers
// learn of it from EvPlaylistRemoved. It returns false for a playlist that is not there, or that
// was taken out already.
func (e *Engine) RemoveSource(playlist int) bool {
	e.mu.Lock()
	if e.closed || playlist < 0 || playlist >= len(e.sources) || e.removed[playlist] {
		e.mu.Unlock()
		return false
	}
	e.removed[playlist] = true
	e.mu.Unlock()

	e.log(playlist, -1, "removed from the run")
	e.emit(EvPlaylistRemoved{Playlist: playlist})

	return true
}

// Swap gives a song another video to download. It works for a song held for review, and also for
// one that was saved, failed or was cancelled: the files it saved are removed, and the new video
// takes its place. A song that is being looked up or downloaded cannot be swapped. It returns
// false when the song cannot be, or has nothing to swap in.
func (e *Engine) Swap(entryID int, videoID string) bool {
	if videoID == "" {
		return false
	}
	e.mu.Lock()
	if entryID < 0 || entryID >= len(e.entries) {
		e.mu.Unlock()
		return false
	}
	en := e.entries[entryID]
	if e.removed[en.info.Playlist] {
		e.mu.Unlock()
		return false
	}
	state := en.state
	switch state {
	case StatePlanned:
		e.mu.Unlock()
		return e.Release(entryID, videoID)
	case StateDone, StateDuplicate, StateFailed, StateNotFound, StateSkipped:
	default:
		e.mu.Unlock()
		return false
	}
	old := en.saved
	if claims := e.claims[en.info.Playlist]; claims != nil && claims[en.targetID] == entryID {
		delete(claims, en.targetID)
	}
	en.saved = nil
	en.targetID = videoID
	en.official, en.better, en.ownOfficial = false, false, false
	en.userChose, en.fixed = true, true
	en.reason = "chosen by you"
	en.err = ""
	en.deferred = false
	en.state = StateAccepted
	e.pending++
	e.idleSent = false
	e.mu.Unlock()

	for _, path := range old {
		_ = os.Remove(path)
	}
	e.emit(EvEntryState{Entry: entryID, State: StateAccepted, TargetID: videoID, Reason: "chosen by you"})
	e.log(en.info.Playlist, entryID, "swapped for %s by you", videoID)
	e.queue.Push(entryID)
	go e.rename(entryID, videoID)

	return true
}

// rename gives a song that was swapped the name of its new video, as soon as yt-dlp says it.
func (e *Engine) rename(entryID int, videoID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	title, channel, err := ytdlp.VideoDetails(ctx, e.opts.YtDlp, videoID)
	if err != nil {
		return
	}
	e.mu.Lock()
	closed := e.closed
	en := e.entries[entryID]
	current := en.targetID == videoID
	if current {
		en.info.Title, en.info.Channel = title, channel
		en.raw.Title, en.raw.Channel = title, channel
		en.named = nil
	}
	e.mu.Unlock()
	if current && !closed {
		e.emit(EvEntryRenamed{Entry: entryID, Title: title, Channel: channel})
	}
}
