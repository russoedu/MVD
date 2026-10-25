package review

import "youtube-downloader/libs/mvd-core/engine"

// Effect is something a call on the session asks the host to do. The host does it and, when the
// effect has an answer, tells the session (Opened, Described, FilePicked).
type Effect interface{ isEffect() }

// Leave says the person left the editor.
type Leave struct{}

// PlanFinished says the plan run is over: the host closes it.
type PlanFinished struct{}

// Download asks the host to download these songs; Keys say which rows they are.
type Download struct {
	Plan []engine.PlannedEntry
	Keys []string
}

// OpenURL asks the host to show an address in the person's browser, then to tell the session with
// Opened.
type OpenURL struct{ URL string }

// DescribeVideo asks the host for the title of a video the person picked, then to tell the session
// with Described. It may be slow.
type DescribeVideo struct{ Key, VideoID string }

// PickOpenFile asks the host for the system's open-file window, starting in Start, then to tell
// the session with FilePicked.
type PickOpenFile struct{ Start string }

// PickSaveFile asks the host for the system's save-file window, starting in Dir with Name offered,
// then to tell the session with FilePicked.
type PickSaveFile struct{ Dir, Name string }

func (Leave) isEffect()         {}
func (PlanFinished) isEffect()  {}
func (Download) isEffect()      {}
func (OpenURL) isEffect()       {}
func (DescribeVideo) isEffect() {}
func (PickOpenFile) isEffect()  {}
func (PickSaveFile) isEffect()  {}
