package prefsform

// windowKinds are the window's names of the kinds of setting.
var windowKinds = map[FieldKind]WindowKind{
	KindChoice:   WindowChoice,
	KindStepper:  WindowNumber,
	KindText:     WindowText,
	KindFolder:   WindowFolder,
	KindTemplate: WindowTemplate,
	KindColour:   WindowColour,
	KindAction:   WindowAction,
}

// welcomeText is what the first start says over the settings.
const welcomeText = "Welcome to MVD! Check where the videos go and how they are named, then Save."

// Window is the form as a window draws it: the tab shown with its settings and their values.
func (f Form) Window() Window {
	w := Window{Title: "Preferences", Tab: f.Tab, Status: f.Status, Bad: f.StatusBad, Busy: f.Busy, Fields: []WindowField{}}
	if f.Welcome {
		w.Welcome = welcomeText
	}
	for _, page := range f.Pages {
		w.Tabs = append(w.Tabs, page.Name)
	}
	for _, field := range f.Pages[f.Tab].Fields {
		w.Fields = append(w.Fields, f.windowField(field))
	}
	c := f.Cfg.Colors
	w.Colours = WindowColours{
		Background: c.Background, Accent: c.Accent, Focus: c.Focus, Highlight: c.Highlight,
		Success: c.Success, Error: c.Error, Dim: c.Dim, Text: c.Text, Selected: c.Selected,
	}

	return w
}

// windowField is a setting as the window draws it.
func (f Form) windowField(field Field) WindowField {
	out := WindowField{
		Label: field.Label, Hint: field.Hint, Kind: windowKinds[field.Kind], Placeholder: field.Placeholder,
		Min: field.Min, Max: field.Max, Step: field.Step, Off: field.Off, Optional: field.Optional,
		Hidden: !field.ShownFor(f.Cfg),
	}
	if field.Get != nil {
		out.Value = field.Get(f.Cfg)
	}
	for _, option := range field.OptionsFor(f.Cfg) {
		o := WindowOption{Value: option, Label: field.OptionTextFor(f.Cfg, option)}
		if field.Swatches != nil {
			o.Swatches = field.Swatches(option)
		}
		out.Options = append(out.Options, o)
	}
	if field.Kind == KindTemplate {
		for _, u := range ParseTemplate(out.Value) {
			if u.IsBlock() {
				out.Units = append(out.Units, WindowUnit{Text: u.Label, Block: true})
			} else {
				out.Units = append(out.Units, WindowUnit{Text: string(u.R)})
			}
		}
		for _, b := range TemplateBlocks {
			out.Blocks = append(out.Blocks, b.Name)
		}
		out.Example = ExampleName(out.Value, "mp4")
	}

	return out
}
