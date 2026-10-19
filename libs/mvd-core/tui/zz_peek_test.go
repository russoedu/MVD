package tui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/workspace"
)

func TestPeek(t *testing.T) {
	cfg := config.Config{OutputDir: t.TempDir(), MaxConcurrentDownloads: 2}
	a := NewWorkspaceApp(WorkspaceInput{Setup: SetupInput{Cfg: cfg}, List: downloadlist.New([]string{"https://youtube.com/playlist?list=A", "https://youtu.be/x"})})
	m, _ := a.Update(tea.WindowSizeMsg{Width: 110, Height: 32})
	fmt.Println(m.View().Content)
	m, _ = m.Update(tea.KeyPressMsg{Text: "e", Code: 'e'})
	fmt.Println(m.View().Content)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m, _ = m.Update(workspace.PreferencesRequestedMsg{})
	fmt.Println(m.View().Content)
}
