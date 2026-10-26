package workspace

import (
	"strings"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/workspace/addsource"
)

// labelCells is how wide the names of the fields are, with the gap after them.
const labelCells = 10

// addLayout is the window for adding music drawn once: its lines, and where on them the choices, the
// fields and the buttons are, counted from the top left of the inside of the frame.
type addLayout struct {
	lines   []string
	radioY  int
	radio   []dialog.Span
	fieldY  map[addsource.Field]int
	boxX0   int
	boxW    int
	buttonY int
	buttons []dialog.Span
}

// addLayout draws the window for the screen m.width by m.height: it covers all of it.
func (m Model) addLayout() addLayout {
	inner := max(m.width-2, 10)
	rows := max(m.height-2, 6)
	l := addLayout{fieldY: map[addsource.Field]int{}, boxX0: 2 + labelCells}
	l.boxW = max(inner-l.boxX0-2, 8)
	s := m.add

	add := func(text string) { l.lines = append(l.lines, text) }
	add("")
	add("  " + look.Dim.Render("How do you want to add music?"))
	l.radioY = len(l.lines)
	var radio strings.Builder
	radio.WriteString("  ")
	x := 2
	for i, name := range addsource.Modes {
		mark := "( )"
		if addsource.Mode(i) == s.Mode {
			mark = "(•)"
		}
		text := mark + " " + name
		style := look.Text
		if s.Focus == addsource.FieldMode && addsource.Mode(i) == s.Mode {
			style = look.Title.Reverse(true)
		}
		l.radio = append(l.radio, dialog.Span{X0: x, X1: x + len([]rune(text))})
		radio.WriteString(style.Render(text) + "     ")
		x += len([]rune(text)) + 5
	}
	add(strings.TrimRight(radio.String(), " "))
	add("")

	for _, f := range s.Fields() {
		if !addsource.IsText(f) {
			continue
		}
		l.fieldY[f] = len(l.lines)
		in := m.addBoxes[f]
		bracket := look.Dim
		if s.Focus == f {
			bracket = look.Yellow
		}
		view := cells.Fit(in.View(), l.boxW-2)
		add("  " + look.Text.Render(cells.Fit(f.Label(), labelCells-2)) + "  " + bracket.Render("[") + view + bracket.Render("]"))
	}
	add("")
	if s.Problem != "" {
		add("  " + look.Red.Render(cells.Fit(s.Problem, inner-4)))
	} else {
		add("")
	}
	for _, line := range addsource.Help(s.Mode) {
		add("  " + look.Dim.Render(cells.Fit(line, inner-4)))
	}

	// The buttons are on the last row, wherever the help ends.
	for len(l.lines) < rows-1 {
		add("")
	}
	if len(l.lines) > rows-1 {
		l.lines = l.lines[:rows-1]
	}
	focus := -1
	switch s.Focus {
	case addsource.FieldAdd:
		focus = 0
	case addsource.FieldCancel:
		focus = 1
	}
	label, spans := dialog.Buttons(addsource.Buttons, focus, nil)
	pad := 2
	for i := range spans {
		spans[i].X0 += pad
		spans[i].X1 += pad
	}
	l.buttonY, l.buttons = len(l.lines), spans
	add(strings.Repeat(" ", pad) + label)

	return l
}

// addView is the window for adding music: it takes the whole screen.
func (m Model) addView() string {
	l := m.addLayout()
	lines := make([]string, len(l.lines))
	for i, line := range l.lines {
		lines[i] = cells.Fit(line, max(m.width-2, 1))
	}

	return dialog.Frame(addsource.Heading, lines, m.width, m.height)
}
