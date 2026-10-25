package prefsform

// The name pattern of the files is edited as units: blocks, and the characters typed between
// them (see ParseTemplate). The cursor sits at a boundary between units.

// units are the units of the pattern the cursor is on, with the caret held inside them.
func (f *Form) units() []Unit {
	units := ParseTemplate(f.Current().Get(f.Cfg))
	f.Caret = max(min(f.Caret, len(units)), 0)

	return units
}

// store makes units the pattern.
func (f *Form) store(units []Unit) { f.Current().Set(&f.Cfg, FormatTemplate(units)) }

// TemplateCaretTo puts the cursor at boundary pos of the pattern.
func (f *Form) TemplateCaretTo(pos int) {
	f.units()
	f.Caret = max(min(pos, f.TemplateLen()), 0)
}

// TemplateCaretBy moves the cursor over delta units, a block as one.
func (f *Form) TemplateCaretBy(delta int) {
	f.units()
	f.Caret = max(min(f.Caret+delta, f.TemplateLen()), 0)
}

// TemplateInsert types text into the pattern at the cursor.
func (f *Form) TemplateInsert(text string) {
	if text == "" {
		return
	}
	units := f.units()
	var add []Unit
	for _, r := range text {
		add = append(add, Unit{R: r})
	}
	f.store(InsertUnits(units, f.Caret, add...))
	f.Caret += len(add)
}

// TemplateBackspace takes out the unit before the cursor, a block as a whole.
func (f *Form) TemplateBackspace() {
	units := f.units()
	if f.Caret > 0 {
		units, f.Caret = RemoveUnit(units, f.Caret-1), f.Caret-1
	}
	f.store(units)
}

// TemplateDelete takes out the unit at the cursor.
func (f *Form) TemplateDelete() {
	f.store(RemoveUnit(f.units(), f.Caret))
}

// TemplateCancel puts the pattern back as it was when typing began and ends it.
func (f *Form) TemplateCancel() {
	f.Current().Set(&f.Cfg, f.Before)
	f.Editing = false
}

// TemplateDone ends typing in the pattern.
func (f *Form) TemplateDone() { f.Editing = false }

// AddBlock puts block index of the palette into the pattern where the cursor is; typing in it
// starts when it was not, with the cursor at the end.
func (f *Form) AddBlock(index int) {
	if f.Current().Kind != KindTemplate || index < 0 || index >= len(TemplateBlocks) {
		return
	}
	if !f.Editing {
		f.Edit()
	}
	b := TemplateBlocks[index]
	units := f.units()
	f.store(InsertUnits(units, f.Caret, Unit{Token: b.Token, Label: b.Name}))
	f.Caret++
}

// PlaceCaret puts the cursor of the pattern at boundary pos, starting to type when it was not.
func (f *Form) PlaceCaret(pos int) {
	if !f.Editing {
		f.Edit()
	}
	f.Caret = pos
}

// RemoveBlock takes unit i out of the pattern.
func (f *Form) RemoveBlock(i int) {
	f.store(RemoveUnit(ParseTemplate(f.Current().Get(f.Cfg)), i))
	if f.Caret > i {
		f.Caret--
	}
}

// MoveBlock takes unit from out of the pattern and puts it where the boundary to (counted
// before it was taken out) is.
func (f *Form) MoveBlock(from, to int) {
	f.store(MoveUnit(ParseTemplate(f.Current().Get(f.Cfg)), from, to))
}
