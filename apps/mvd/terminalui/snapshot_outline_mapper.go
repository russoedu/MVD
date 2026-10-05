package terminalui

import (
	"strconv"

	ttygo "github.com/TReactUI/TReactUI/packages/tty-go"

	"youtube-downloader/libs/mvd-core/tui"
)

// snapshotFromOutline turns mvd's description of a screen into the ARIA
// description the browser exposes to assistive technology.
func snapshotFromOutline(out tui.ScreenOutline) ttygo.Snapshot {
	var nodes []ttygo.A11yNode

	if out.Prompt != "" {
		nodes = append(nodes, ttygo.A11yNode{Role: "status", Value: out.Prompt})
	}
	if out.HasText && len(out.Items) == 0 {
		label := out.TextLabel
		if label == "" {
			label = "Download list, one URL per line"
		}
		nodes = append(nodes, ttygo.A11yNode{Role: "textbox", Label: label, Value: out.Text, Focused: true})
	} else if out.HasText {
		nodes = append(nodes, ttygo.A11yNode{Role: "text", Value: "Current folder: " + out.Text})
	}
	if len(out.Items) > 0 {
		nodes = append(nodes, itemsListbox(out))
		nodes = append(nodes, editorNodes(out.Items)...)
	}
	if len(out.Keys) > 0 {
		nodes = append(nodes, keysList(out.Keys))
	}
	return ttygo.Snapshot{Title: out.Title, Nodes: nodes}
}

func itemsListbox(out tui.ScreenOutline) ttygo.A11yNode {
	label := out.ItemsLabel
	if label == "" {
		label = "Items"
	}
	options := make([]ttygo.A11yNode, len(out.Items))
	for i, item := range out.Items {
		text := item.Label
		if item.Value != "" {
			text += ": " + item.Value
		}
		options[i] = ttygo.A11yNode{Role: "option", Label: text, Selected: item.Selected, Focused: item.Selected && !item.Editing}
	}
	return ttygo.A11yNode{Role: "listbox", Label: label, Children: options}
}

// editorNodes describes the open editor, if any: a text field, or the choices of a radio selector.
func editorNodes(items []tui.OutlineItem) []ttygo.A11yNode {
	for _, item := range items {
		if !item.Editing {
			continue
		}
		if len(item.Choices) == 0 {
			return []ttygo.A11yNode{{Role: "textbox", Label: "Editing " + item.Label, Value: item.Value, Focused: true}}
		}
		choices := make([]ttygo.A11yNode, len(item.Choices))
		for i, choice := range item.Choices {
			choices[i] = ttygo.A11yNode{Role: "option", Label: choice, Selected: i == item.Chosen, Focused: i == item.Chosen}
		}
		return []ttygo.A11yNode{{Role: "listbox", Label: "Choices for " + item.Label + ", " + strconv.Itoa(len(item.Choices)) + " options", Children: choices}}
	}
	return nil
}

func keysList(keys []tui.OutlineKey) ttygo.A11yNode {
	items := make([]ttygo.A11yNode, len(keys))
	for i, k := range keys {
		items[i] = ttygo.A11yNode{Role: "listitem", Value: k.Key + ": " + k.Description}
	}
	return ttygo.A11yNode{Role: "list", Label: "Keyboard shortcuts", Children: items}
}
