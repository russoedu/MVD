package workspace

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/menubar"
)

func withTools(state *ToolsState, addresses ...string) Model {
	start := func(config.Config, []string, []engine.PlannedEntry, bool) (ListRun, error) { return nil, nil }
	m := New(Host{Start: start, Tools: func() ToolsState { return *state }}, config.Config{}, downloadlist.New(addresses))

	return sized(m)
}

func TestWhileTheToolsAreBeingFetchedStartCannotBeUsedAndSaysWhy(t *testing.T) {
	state := &ToolsState{Message: "Getting the download tools (one time): yt-dlp, ffmpeg..."}
	m := withTools(state, "https://a")
	buttons := m.buttons()
	if !buttons[0].disabled || !buttons[1].disabled {
		t.Errorf("Start and Look up first should be disabled: %+v", buttons)
	}
	if buttons[2].disabled {
		t.Error("Add link should still work while the tools come")
	}
	if screen := screenOf(m); !strings.Contains(screen, "Getting the download tools") {
		t.Errorf("the screen does not say the tools are coming:\n%s", screen)
	}
	next, _ := m.Update(menubar.SelectedMsg{Name: entryStart})
	if got := next.(Model).caution; !strings.Contains(got, "Getting the download tools") {
		t.Errorf("starting too early says %q", got)
	}
}

func TestOnceTheToolsAreReadyStartIsNoLongerHeldBack(t *testing.T) {
	state := &ToolsState{Message: "Getting the download tools..."}
	m := withTools(state, "https://a")
	state.Ready, state.Message = true, ""
	if buttons := m.buttons(); buttons[0].disabled {
		t.Error("Start is still disabled after the tools are ready")
	}
	if got := m.toolsMessage(); got != "" {
		t.Errorf("the tools still say %q", got)
	}
}

func TestAToolThatCouldNotBeGottenIsSaidAndNotWaitedFor(t *testing.T) {
	state := &ToolsState{Message: "Could not get yt-dlp: no connection. Check the connection and start MVD again.", Failed: true}
	m := withTools(state, "https://a")
	if screen := screenOf(m); !strings.Contains(screen, "Could not get yt-dlp") {
		t.Errorf("the failure is not shown:\n%s", screen)
	}
	if cmd := m.toolsTick(); cmd != nil {
		t.Error("a failure does not change, so there is nothing to wait for")
	}
}

func TestTheStatusIsLookedAtAgainWhileTheToolsAreComing(t *testing.T) {
	state := &ToolsState{Message: "Getting the download tools..."}
	if cmd := withTools(state, "https://a").toolsTick(); cmd == nil {
		t.Error("nothing wakes the screen to look at the tools again")
	}
	state.Ready = true
	if cmd := withTools(state, "https://a").toolsTick(); cmd != nil {
		t.Error("a screen with its tools ready should not keep waking up")
	}
}

func TestWithoutAToolsStateEverythingWorksAsBefore(t *testing.T) {
	m := listOf("https://a")
	if buttons := m.buttons(); buttons[0].disabled {
		t.Error("Start is disabled though nothing says the tools are missing")
	}
}

func TestPointingAtStartWhileTheToolsComeSaysSo(t *testing.T) {
	state := &ToolsState{Message: "Getting the download tools (one time)..."}
	m := withTools(state, "https://a")
	x, y, w, ok := m.buttonRows()
	if !ok {
		t.Fatal("no buttons")
	}
	m.hoverX, m.hoverY = x+w/2, y
	if got := m.disabledReason(); !strings.Contains(got, "Getting the download tools") {
		t.Errorf("the reason is %q", got)
	}
}

func TestTheBeatsOfTheWorkspaceAreBackgroundMessagesAndKeysAreNot(t *testing.T) {
	if !Background(toolsTickMsg(time.Now())) || !Background(saveTickMsg(time.Now())) {
		t.Error("the workspace's own beats must reach it under a popup")
	}
	if Background(tea.KeyPressMsg{Code: 'a'}) || Background(nil) {
		t.Error("a key press or nothing is not a beat")
	}
}
