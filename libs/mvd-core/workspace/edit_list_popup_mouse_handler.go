package workspace

import (
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/playlistfile"
)

// mouse works out a click, a drag or a turn of the wheel on the popup. The position is in
// the cells of the popup, its frame included.
func (p editPopup) mouse(msg tea.MouseMsg) (editPopup, tea.Cmd) {
	pos := msg.Mouse()
	x, y := pos.X-1, pos.Y-1 // inside the frame
	rows := p.rows()
	room := p.room()

	switch msg := msg.(type) {
	case tea.MouseWheelMsg:
		if p.asking != "" {
			return p, nil
		}
		by := 3
		if msg.Button == tea.MouseWheelUp {
			by = -3
		}
		p.top += by
		return p.clampTop(), nil

	case tea.MouseMotionMsg:
		if p.dragging {
			// The list follows the pointer when it is carried past either end.
			switch {
			case y < 0:
				p.top--
			case y >= room:
				p.top++
			}
			p = p.clampTop()
			p.dropAt = max(min(p.top+max(min(y, room-1), 0), len(rows)-1), 0)
		}
		return p, nil

	case tea.MouseReleaseMsg:
		if !p.dragging {
			return p, nil
		}
		p.dragging = false
		return p.drop(), nil

	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft || p.asking != "" {
			return p, nil
		}
	default:
		return p, nil
	}

	if pos.Y == 0 {
		if dialog.CloseSpan(p.w).Contains(pos.X) {
			return p, func() tea.Msg { return editCancelledMsg{} }
		}
		return p, nil
	}
	if y == room {
		_, spans := p.buttonSpans()
		for button, s := range spans {
			if s.Contains(x) {
				return p.press(button)
			}
		}
		return p, nil
	}
	if y < 0 || y >= room || p.top+y >= len(rows) {
		return p, nil
	}
	row := p.top + y
	p.sel = row
	r := rows[row]
	switch {
	case x >= p.removeX() && x < p.removeX()+iconCells:
		return p.remove().reveal(), nil
	case r.song < 0 && x >= p.editX() && x < p.editX()+iconCells:
		return p.ask("edit", "Address", p.items[r.item].Source), nil
	case x >= handleX(r) && x < handleX(r)+iconCells:
		p.dragging, p.dropAt = true, row
	}
	return p, nil
}

// handleX is the first column of the grip a row is carried by.
func handleX(r editRow) int {
	if r.song >= 0 {
		return songHandleX
	}
	return addressHandleX
}

// press does what a button of the row under the list does.
func (p editPopup) press(b editButton) (editPopup, tea.Cmd) {
	switch b {
	case buttonAdd:
		return p.ask("add", "Address to add", ""), nil
	case buttonSave:
		items := p.items
		return p, func() tea.Msg { return editSavedMsg{items: items} }
	case buttonCancel:
		return p, func() tea.Msg { return editCancelledMsg{} }
	}
	return p, nil
}

// drop puts the row that was carried where it was let go: an address among the addresses,
// a song among the songs of its own address.
func (p editPopup) drop() editPopup {
	rows := p.rows()
	if p.sel < 0 || p.sel >= len(rows) || p.dropAt < 0 || p.dropAt >= len(rows) || p.dropAt == p.sel {
		return p
	}
	from, to := rows[p.sel], rows[p.dropAt]
	if from.song < 0 {
		if from.item == to.item {
			return p
		}
		p = p.remember()
		moved := p.items[from.item]
		rest := append(append([]playlistfile.Item(nil), p.items[:from.item]...), p.items[from.item+1:]...)
		items := append([]playlistfile.Item(nil), rest[:to.item]...)
		items = append(items, moved)
		p.items = append(items, rest[to.item:]...)
		return p.selectRow(editRow{to.item, -1}).reveal()
	}
	if to.item != from.item || to.song < 0 {
		return p
	}
	p = p.remember()
	songs := p.items[from.item].Songs
	moved := songs[from.song]
	rest := append(append([]playlistfile.Entry(nil), songs[:from.song]...), songs[from.song+1:]...)
	out := append([]playlistfile.Entry(nil), rest[:to.song]...)
	out = append(out, moved)
	p.items[from.item].Songs = append(out, rest[to.song:]...)
	return p.selectRow(editRow{from.item, to.song}).reveal()
}
