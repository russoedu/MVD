package prefsform

import (
	"strings"

	"youtube-downloader/libs/mvd-core/config"
)

// ShowTab puts the cursor on the first setting of a tab; the tabs wrap round.
func (f *Form) ShowTab(tab int) {
	f.Tab = (tab + len(f.Pages)) % len(f.Pages)
	f.Cursor, f.Editing, f.Status = 0, false, ""
}

// MoveCursor moves to a setting of the page, up (negative) or down.
func (f *Form) MoveCursor(delta int) {
	f.Cursor = max(min(f.Cursor+delta, len(f.Pages[f.Tab].Fields)-1), 0)
}

// SelectSetting puts the cursor on setting i of the page.
func (f *Form) SelectSetting(i int) { f.Cursor = i }

// ClearStatus forgets the last thing that was said.
func (f *Form) ClearStatus() { f.Status = "" }

// Change moves the setting one option or one step: -1 back, 1 forward.
func (f *Form) Change(by int) {
	field := f.Current()
	value := field.Get(f.Cfg)
	switch field.Kind {
	case KindChoice:
		at := 0
		for i, o := range field.Options {
			if o == value {
				at = i
			}
		}
		if next := at + by; next >= 0 && next < len(field.Options) {
			field.Set(&f.Cfg, field.Options[next])
		}
	case KindStepper:
		n := min(max(parseInt(value)+by*field.Step, field.Min), field.Max)
		field.Set(&f.Cfg, intText(n))
	}
}

// ChooseOption sets the setting the cursor is on to its option i.
func (f *Form) ChooseOption(i int) {
	field := f.Current()
	field.Set(&f.Cfg, field.Options[i])
}

// Activate does what Enter does on the setting: another option, text to type, a folder to
// choose, or the action. The effect, when there is one, is for the host.
func (f *Form) Activate() Effect {
	field := f.Current()
	switch field.Kind {
	case KindChoice:
		at := 0
		for i, o := range field.Options {
			if o == field.Get(f.Cfg) {
				at = i
			}
		}
		field.Set(&f.Cfg, field.Options[(at+1)%len(field.Options)])
	case KindText, KindTemplate, KindColour:
		f.Edit()
	case KindFolder:
		if !f.caps.CanPickFolder {
			f.Edit()

			return nil
		}
		f.Busy, f.Status, f.StatusBad = true, "Choose a folder in the window that opened...", false

		return PickFolder{Field: f.Cursor, Start: field.Get(f.Cfg)}
	case KindAction:
		f.Busy, f.Status, f.StatusBad = true, "Answer the questions in the window that opened...", false

		return RunAction{Kind: field.Action}
	}

	return nil
}

// Picked takes the answer of the folder chooser. When it could not be shown, the path is typed
// instead.
func (f *Form) Picked(field int, path string, chosen bool, err error) {
	f.Busy, f.Status = false, ""
	f.Cursor = field
	switch {
	case err != nil:
		f.Edit()
	case chosen:
		f.Current().Set(&f.Cfg, path)
	}
}

// ActionDone takes what an action of the host came to.
func (f *Form) ActionDone(message string, err error) {
	f.Busy, f.Status, f.StatusBad = false, message, err != nil
	if err != nil {
		f.Status = err.Error()
	}
}

// Edit starts typing the value of the setting the cursor is on. For a text the renderer then
// shows Before in its text box.
func (f *Form) Edit() {
	f.Before = f.Current().Get(f.Cfg)
	if f.Current().Kind == KindTemplate {
		f.Caret = len(ParseTemplate(f.Before))
	}
	f.Editing, f.Status = true, ""
}

// Accept takes what was typed as the value of the setting.
func (f *Form) Accept(typed string) {
	field := f.Current()
	if field.Kind == KindTemplate {
		f.Editing, f.Status = false, ""

		return
	}
	text := strings.TrimSpace(typed)
	switch field.Kind {
	case KindColour:
		if text == "" && field.Optional {
			field.Set(&f.Cfg, "")
			f.Editing, f.Status = false, ""

			return
		}
		if !config.ValidColour(text) {
			f.Status, f.StatusBad = "Use a colour like #ff007f or #f07.", true

			return
		}
		field.Set(&f.Cfg, text)
	case KindText:
		field.Set(&f.Cfg, text)
	default:
		if text != "" {
			field.Set(&f.Cfg, text)
		}
	}
	f.Editing, f.Status = false, ""
}

// CancelEdit gives up typing a text; the setting keeps its value.
func (f *Form) CancelEdit() { f.Editing, f.Status = false, "" }

// UseDefaultColour puts the colour the cursor is on back to its default.
func (f *Form) UseDefaultColour() {
	field := f.Current()
	d := config.DefaultThemeColors()
	field.Set(&f.Cfg, *field.Slot(&d))
}
