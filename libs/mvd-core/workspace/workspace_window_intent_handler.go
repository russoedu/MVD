package workspace

import (
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/workspace/addsource"
	"youtube-downloader/libs/mvd-core/workspace/windowsnapshot"
)

// WindowIntentMsg is what the person did in a classic themed window. The workspace does what the
// terminal look does for the same click or key, so both looks behave alike.
type WindowIntentMsg windowsnapshot.Intent

// windowIntent does what an intent of a themed window asks.
func (m Model) windowIntent(in windowsnapshot.Intent) (tea.Model, tea.Cmd) {
	switch in.Kind {
	case windowsnapshot.IntentMenu:
		if !m.bar.Enabled(in.ID) {
			return m, nil
		}
		return m.choose(in.ID)
	case windowsnapshot.IntentButton:
		for _, b := range m.buttons() {
			if b.entry == in.ID && !b.disabled {
				return m.choose(b.entry)
			}
		}
	case windowsnapshot.IntentSelectSource:
		m.show(m.view().ChoosePlaylist(in.Index))
	case windowsnapshot.IntentSelectSong:
		m.show(m.view().ChooseEntry(in.Index))
	case windowsnapshot.IntentTickSong:
		if ids := m.dl.VisibleEntries(); in.Index >= 0 && in.Index < len(ids) {
			m.show(m.view().SetTicked(ids[in.Index], in.Ticked))
		}
	case windowsnapshot.IntentAddingMode, windowsnapshot.IntentAddingText, windowsnapshot.IntentAddingButton:
		if m.addOpen {
			return m.windowAddIntent(in)
		}
	case windowsnapshot.IntentQuestionButton, windowsnapshot.IntentQuestionText:
		return m.questionIntent(in)
	case windowsnapshot.IntentKey:
		if key, ok := keyPress(in.Key); ok {
			return m.update(key)
		}
	}
	// The window's own buttons (minimise, maximise, close) are the host's: closing reaches the
	// workspace as CloseRequestedMsg, as the operating system's close button does.

	return m, nil
}

// windowAddIntent does what is asked in the window for adding music.
func (m Model) windowAddIntent(in windowsnapshot.Intent) (tea.Model, tea.Cmd) {
	switch in.Kind {
	case windowsnapshot.IntentAddingMode:
		mode := addsource.ModeURL
		if in.Mode == windowsnapshot.ModeSearch {
			mode = addsource.ModeSearch
		}
		m.add = m.add.WithMode(mode)

		return m.syncAddFocus(), nil
	case windowsnapshot.IntentAddingText:
		for f, name := range addFields {
			if name == in.Field {
				m.add = m.add.SetText(f, in.Text).Focused(f)
				box := m.addBoxes[f]
				box.SetValue(in.Text)
				m.addBoxes[f] = box

				return m.syncAddFocus(), nil
			}
		}
	case windowsnapshot.IntentAddingButton:
		if in.ID == "add" {
			return m.submitAdd()
		}

		return m.closeAdd(), nil
	}

	return m, nil
}

// keys are the keys an intent may name that are not a character.
var keys = map[string]rune{
	"Enter": tea.KeyEnter, "Space": tea.KeySpace, "Escape": tea.KeyEscape, "Delete": tea.KeyDelete, "Tab": tea.KeyTab,
}

// keyPress is the key press an intent names: a character, or one of keys.
func keyPress(name string) (tea.KeyPressMsg, bool) {
	if code, ok := keys[name]; ok {
		return tea.KeyPressMsg{Code: code}, true
	}
	if r, size := utf8.DecodeRuneInString(name); size > 0 && size == len(name) {
		return tea.KeyPressMsg{Code: r, Text: name}, true
	}

	return tea.KeyPressMsg{}, false
}
