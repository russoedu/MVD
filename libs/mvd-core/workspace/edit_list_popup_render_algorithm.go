package workspace

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/look"
)

// The columns of a row, counted from the left of the inside of the popup. The icons sit at
// the right end, so the address takes what is left.
const (
	addressHandleX = 0 // the grip of an address, 3 cells
	songHandleX    = 4 // the grip of a song, 3 cells
	iconCells      = 3
)

// editButton is what a button of the popup does.
type editButton int

const (
	buttonAdd editButton = iota
	buttonPaste
	buttonUndo
	buttonRedo
	buttonSave
	buttonCancel
)

var (
	editButtonsLeft  = []string{"Add", "Paste", "Undo", "Redo"}
	editButtonsRight = []string{"Save", "Cancel"}
)

// inner is the width and height of the inside of the frame.
func (p editPopup) inner() (int, int) { return max(p.w-2, 1), max(p.h-2, 1) }

// room is how many rows of the list show: the inside less the row of buttons and the line
// of help.
func (p editPopup) room() int {
	_, ih := p.inner()
	return max(ih-2, 1)
}

// editX and removeX are the first columns of the icons that edit and remove the row.
func (p editPopup) editX() int {
	iw, _ := p.inner()
	return iw - 2*iconCells - 2
}

func (p editPopup) removeX() int {
	iw, _ := p.inner()
	return iw - iconCells - 2
}

// sized gives the popup the room it has and keeps the selection in view.
func (p editPopup) sized(w, h int) editPopup {
	p.w, p.h = w, h
	return p.reveal()
}

// reveal scrolls the list so the selected row shows.
func (p editPopup) reveal() editPopup {
	room := p.room()
	if p.sel < p.top {
		p.top = p.sel
	}
	if p.sel >= p.top+room {
		p.top = p.sel - room + 1
	}
	return p.clampTop()
}

func (p editPopup) clampTop() editPopup {
	p.top = max(min(p.top, len(p.rows())-p.room()), 0)
	return p
}

// buttonSpans are the columns of the buttons in the row under the list, counted from the
// left of the inside: the ones that change the list at the left, save and cancel at the right.
func (p editPopup) buttonSpans() (line string, at map[editButton]dialog.Span) {
	iw, _ := p.inner()
	disabled := map[string]bool{"Undo": len(p.undone) == 0, "Redo": len(p.redone) == 0}
	left, leftSpans := dialog.Buttons(editButtonsLeft, -1, disabled)
	right, rightSpans := dialog.Buttons(editButtonsRight, -1, nil)
	rightX := max(iw-ansi.StringWidth(right)-1, ansi.StringWidth(left)+3)

	at = map[editButton]dialog.Span{}
	for i, s := range leftSpans {
		at[editButton(i)] = dialog.Span{X0: s.X0 + 1, X1: s.X1 + 1}
	}
	for i, s := range rightSpans {
		at[buttonSave+editButton(i)] = dialog.Span{X0: s.X0 + rightX, X1: s.X1 + rightX}
	}
	gap := strings.Repeat(" ", max(rightX-1-ansi.StringWidth(left), 1))
	return " " + left + gap + right, at
}

// View draws the popup.
func (p editPopup) View() string {
	iw, ih := p.inner()
	room := p.room()
	rows := p.rows()
	var lines []string
	for i := p.top; i < min(len(rows), p.top+room); i++ {
		lines = append(lines, p.rowLine(i, rows[i], iw))
	}
	if len(rows) == 0 {
		lines = append(lines, look.Dim.Render("  The list is empty: press Add to put an address in it"))
	}
	for len(lines) < room {
		lines = append(lines, "")
	}
	buttons, _ := p.buttonSpans()
	lines = append(lines, buttons)
	if p.asking != "" {
		lines = append(lines, look.Key.Render(" "+p.input.View()))
	} else {
		lines = append(lines, look.Dim.Render(" Drag ≡ to reorder · click ✎ to change an address · ✕ to remove · Esc cancels"))
	}
	for len(lines) < ih {
		lines = append(lines, "")
	}
	return dialog.ClosableFrame(fmt.Sprintf("Edit the list · %d addresses", len(p.items)), lines, p.w, p.h)
}

// rowLine draws row i: its grip, the address or song, and the icons to change or remove it.
func (p editPopup) rowLine(i int, r editRow, iw int) string {
	song := r.song >= 0
	name := ""
	handle := addressHandleX
	if song {
		name, handle = songName(p.items[r.item].Songs[r.song]), songHandleX
	} else {
		name = cells.Display(p.items[r.item].Source)
	}
	textX := handle + iconCells
	textW := max(p.editX()-textX, 1)

	style := lipgloss.NewStyle()
	switch {
	case p.dragging && i == p.sel:
		style = look.Dim.Faint(true)
	case p.dragging && i == p.dropAt:
		style = look.Title.Reverse(true)
	case i == p.sel:
		style = look.Key
	case song:
		style = look.Dim
	}
	marker := " "
	if i == p.sel {
		marker = "▸"
	}
	line := cells.Fit(strings.Repeat(" ", handle)+marker+"≡ "+style.Render(cells.Fit(name, textW)), p.editX())
	if song {
		line += strings.Repeat(" ", iconCells)
	} else {
		line += look.Cyan.Render(" ✎ ")
	}
	line += look.Red.Render(" ✕ ")
	return cells.Fit(line, iw)
}
