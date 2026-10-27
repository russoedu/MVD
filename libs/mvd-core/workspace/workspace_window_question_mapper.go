package workspace

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/workspace/closequestion"
	"youtube-downloader/libs/mvd-core/workspace/windowsnapshot"
)

// The ids of the buttons of the questions, as a themed window sends them back.
const (
	questionQuit   = "quit"
	questionKeep   = "keep"
	questionOK     = "ok"
	questionCancel = "cancel"
)

// windowQuestion is the question or note open over the window, for a theme to draw as its system's
// dialog, or nil when there is none.
func (m Model) windowQuestion() *windowsnapshot.Question {
	switch {
	case m.closing.On:
		what, ask := closequestion.Words(m.downloading())

		return &windowsnapshot.Question{
			ID: windowsnapshot.QuestionQuit, Title: "Quit MVD?", Tone: windowsnapshot.ToneCaution, Lines: []string{what, ask},
			Buttons: []windowsnapshot.Button{
				{ID: questionQuit, Label: closequestion.Answers[closequestion.AnswerQuit], Role: windowsnapshot.RoleOther, Enabled: true},
				// Keeping the downloads is the safe answer: Return and Esc both choose it.
				{ID: questionKeep, Label: closequestion.Answers[closequestion.AnswerKeep], Role: windowsnapshot.RoleDefault, Enabled: true},
			},
		}
	case m.asking != askNone:
		prompt := strings.TrimSuffix(strings.TrimSpace(m.input.Prompt), ":")

		return &windowsnapshot.Question{
			ID: windowsnapshot.QuestionAsk, Title: prompt, Tone: windowsnapshot.ToneAsk,
			Field: &windowsnapshot.QuestionField{Label: prompt, Text: m.input.Value()},
			Buttons: []windowsnapshot.Button{
				{ID: questionOK, Label: "OK", Role: windowsnapshot.RoleDefault, Enabled: true},
				{ID: questionCancel, Label: "Cancel", Role: windowsnapshot.RoleCancel, Enabled: true},
			},
		}
	case m.about:
		version := m.host.Version
		if version == "" {
			version = "dev"
		}
		lines := append([]string{"MVD - Music Video Downloader", "Version " + version}, aboutParagraphs...)

		return &windowsnapshot.Question{
			ID: windowsnapshot.QuestionAbout, Title: entryAbout, Tone: windowsnapshot.ToneNote,
			Lines:   append(lines, aboutFooter, projectURL),
			Buttons: []windowsnapshot.Button{{ID: questionOK, Label: "OK", Role: windowsnapshot.RoleDefault, Enabled: true}},
		}
	case m.shortcuts:
		var sections []windowsnapshot.QuestionSection
		for _, section := range shortcutSections() {
			out := windowsnapshot.QuestionSection{Heading: section.heading}
			for _, k := range section.keys {
				out.Rows = append(out.Rows, windowsnapshot.QuestionRow{Keys: k.keys, Does: k.does})
			}
			sections = append(sections, out)
		}

		return &windowsnapshot.Question{
			ID: windowsnapshot.QuestionShortcuts, Title: entryShortcuts, Tone: windowsnapshot.ToneNote, Sections: sections,
			Buttons: []windowsnapshot.Button{{ID: questionOK, Label: "OK", Role: windowsnapshot.RoleDefault, Enabled: true}},
		}
	}

	return nil
}

// questionIntent answers the question or note open over the window, as its keys would.
func (m Model) questionIntent(in windowsnapshot.Intent) (tea.Model, tea.Cmd) {
	switch {
	case m.closing.On && in.Kind == windowsnapshot.IntentQuestionButton:
		if in.ID == questionQuit {
			return m.quitNow()
		}
		m.closing = closequestion.State{}
	case m.asking != askNone && in.Kind == windowsnapshot.IntentQuestionText:
		m.input.SetValue(in.Text)
		m.input.CursorEnd()
	case m.asking != askNone && in.Kind == windowsnapshot.IntentQuestionButton:
		if in.ID == questionOK {
			return m.answer(tea.KeyPressMsg{Code: tea.KeyEnter})
		}
		m.asking = askNone
	case in.Kind == windowsnapshot.IntentQuestionButton:
		m.about, m.shortcuts = false, false
	}

	return m, nil
}
