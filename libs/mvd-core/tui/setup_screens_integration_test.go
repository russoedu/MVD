package tui

import (
	"youtube-downloader/libs/mvd-core/testing/keys"

	tea "charm.land/bubbletea/v2"
)

func key(s string) tea.KeyPressMsg { return keys.Press(s) }

// keyText is the key press that types text.
