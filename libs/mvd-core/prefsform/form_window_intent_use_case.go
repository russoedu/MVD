package prefsform

import "strconv"

// Apply does what the person did in the preferences window, through the same steps the keys
// take, and returns what the host has to do for it (a folder to choose, an action to run). Save
// and Close are the host's: they leave the form as it is.
func (f *Form) Apply(in WindowIntent) Effect {
	f.Welcome = false
	if f.Busy {
		return nil
	}
	if in.Kind == IntentTab {
		f.ShowTab(in.Index)

		return nil
	}
	fields := f.Pages[f.Tab].Fields
	if in.Field < 0 || in.Field >= len(fields) {
		return nil
	}
	f.SelectSetting(in.Field)
	field := f.Current()
	switch in.Kind {
	case IntentSet:
		f.set(field, in.Text)
	case IntentActivate:
		if field.Kind == KindFolder || field.Kind == KindAction {
			return f.Activate()
		}
	case IntentDefaultColour:
		if field.Kind == KindColour {
			f.UseDefaultColour()
		}
	case IntentAddBlock, IntentAddText, IntentRemoveUnit:
		if field.Kind == KindTemplate {
			f.changePattern(in)
		}
	}

	return nil
}

// set gives a setting a value, kept as the keys would keep it: an option of a choice, a number
// within its bounds, or a text taken as if typed.
func (f *Form) set(field Field, value string) {
	switch field.Kind {
	case KindChoice:
		for i, option := range field.OptionsFor(f.Cfg) {
			if option == value {
				f.ChooseOption(i)
			}
		}
	case KindStepper:
		if n, err := strconv.Atoi(value); err == nil {
			field.Set(&f.Cfg, intText(min(max(n, field.Min), field.Max)))
		}
	case KindText, KindFolder, KindColour:
		f.Edit()
		f.Accept(value)
	}
}

// changePattern adds a block or text at the end of the name pattern, or takes a piece out.
func (f *Form) changePattern(in WindowIntent) {
	switch in.Kind {
	case IntentAddBlock:
		f.PlaceCaret(f.TemplateLen())
		f.AddBlock(in.Index)
	case IntentAddText:
		f.PlaceCaret(f.TemplateLen())
		f.TemplateInsert(in.Text)
	case IntentRemoveUnit:
		f.RemoveBlock(in.Index)
	}
	f.TemplateDone()
}
