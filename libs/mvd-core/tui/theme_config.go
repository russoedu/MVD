// Package tui shows the run on a fixed full screen interface: counters,
// playlists, entries, the output of the selected item and a key bar.
package tui

import (
	"youtube-downloader/libs/mvd-core/look"

	"youtube-downloader/libs/mvd-core/config"
)

// ApplyTheme sets the interface colours. Call it before the first screen is
// built; it is not safe to call while a program is rendering.
func ApplyTheme(c config.ThemeColors) { look.Apply(c) }
