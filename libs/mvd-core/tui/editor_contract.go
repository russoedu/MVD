package tui

import editorpage "youtube-downloader/libs/mvd-core/editor/editor"

// The contracts of the playlist editor are owned by its slice; a host reaches them here,
// with the app input.

// PlanDownloadStarter starts a run that downloads exactly what a plan names.
type PlanDownloadStarter = editorpage.PlanDownloadStarter

// Planner is what a plan-only run offers once it has dealt with every song.
type Planner = editorpage.Planner

// EditorHost is what the playlist editor needs from the machine it runs on.
type EditorHost = editorpage.EditorHost
