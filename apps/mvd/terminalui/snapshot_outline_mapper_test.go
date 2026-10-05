package terminalui

import (
	"testing"

	"youtube-downloader/libs/mvd-core/tui"
)

func TestListScreenBecomesATextboxWithKeyboardShortcuts(t *testing.T) {
	snapshot := snapshotFromOutline(tui.ScreenOutline{
		Title: "MVD · Download list (1 item)", Text: "https://a", HasText: true,
		Keys: []tui.OutlineKey{{Key: "ctrl+s", Description: "start"}},
	})

	if snapshot.Nodes[0].Role != "textbox" || snapshot.Nodes[0].Value != "https://a" || !snapshot.Nodes[0].Focused {
		t.Fatalf("unexpected textbox %+v", snapshot.Nodes[0])
	}
	if shortcuts := snapshot.Nodes[1]; shortcuts.Role != "list" || shortcuts.Children[0].Value != "ctrl+s: start" {
		t.Fatalf("unexpected shortcuts %+v", shortcuts)
	}
}

func TestSettingsBecomeAListboxAndAnOpenRadioEditorItsChoices(t *testing.T) {
	snapshot := snapshotFromOutline(tui.ScreenOutline{
		Title: "MVD · Preferences",
		Items: []tui.OutlineItem{
			{Label: "Output Folder", Value: "/tmp"},
			{Label: "Video Quality", Value: "best", Selected: true, Editing: true, Choices: []string{"best", "1080p"}, Chosen: 0},
		},
	})

	listbox := snapshot.Nodes[0]
	if listbox.Role != "listbox" || listbox.Children[1].Label != "Video Quality: best" || !listbox.Children[1].Selected {
		t.Fatalf("unexpected listbox %+v", listbox)
	}
	choices := snapshot.Nodes[1]
	if choices.Label != "Choices for Video Quality, 2 options" || len(choices.Children) != 2 || !choices.Children[0].Selected {
		t.Fatalf("unexpected choices %+v", choices)
	}
}

func TestDownloadsAreListedUnderTheirOwnLabel(t *testing.T) {
	snapshot := snapshotFromOutline(tui.ScreenOutline{
		Title: "MVD · Downloads: 0 done, 1 running, 0 queued, 0 failed", ItemsLabel: "Downloads",
		Items: []tui.OutlineItem{{Label: "01 One", Value: "downloading 34%", Selected: true}},
	})

	listbox := snapshot.Nodes[0]
	if listbox.Label != "Downloads" || listbox.Children[0].Label != "01 One: downloading 34%" {
		t.Fatalf("unexpected listbox %+v", listbox)
	}
}
