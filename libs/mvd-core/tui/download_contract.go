package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	downloadpage "youtube-downloader/libs/mvd-core/run/download"
	"youtube-downloader/libs/mvd-core/runstate"
)

// Controller is what the download screen needs from the engine.
type Controller = downloadpage.Controller

// SourceAdder is what a run offers when it can take more links while it works.
type SourceAdder = downloadpage.SourceAdder

// Enabled decides between the full screen interface and plain log output.
func Enabled(noTUIFlag bool) bool { return downloadpage.Enabled(noTUIFlag) }

// RunDownload drives the download screen until the user quits and returns the final
// state for the summary.
func RunDownload(ctx context.Context, c Controller) (*runstate.State, error) {
	return downloadpage.Run(ctx, c)
}

// NewDownloadModel returns the download screen as a Bubble Tea model, for a host that
// runs the program itself. The returned model also has an Outline method (see
// ScreenOutline). It quits when the user leaves the screen.
func NewDownloadModel(c Controller) tea.Model { return downloadpage.New(c) }
