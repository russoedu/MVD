package menumodel

import "strings"

// SetOpen opens menu i, or closes the bar for None or a number there is no menu for.
func (b *Bar) SetOpen(i int) {
	b.open = b.validMenu(i)
	b.cursor = None
	if b.open >= 0 {
		b.menuCursor = b.open
		b.cursor = b.step(None, 1)
	}
}

// Close hides the box and gives the keyboard back to the program.
func (b *Bar) Close() {
	b.open, b.cursor, b.active = None, None, false
}

// ClickHeading is a click on heading i: it opens its box, or closes it when it was open.
func (b *Bar) ClickHeading(i int) {
	if b.open == i {
		b.Close()

		return
	}
	b.SetOpen(i)
	b.active = true
	b.cursor = None
}

// ClickEntry is a click on entry j of the open menu: it chooses it, unless it is greyed out.
func (b *Bar) ClickEntry(j int) (Selection, bool) { return b.choose(b.open, j) }

// ClickAway is a click away from the bar and its box: an open box closes.
func (b *Bar) ClickAway() {
	if b.open >= 0 {
		b.Close()
	}
}

// PointerOver is the pointer moving over the bar: heading is the heading under it on the bar's
// row and entry the entry of the open menu under it, each None for nothing. While a box is open
// the pointer over another heading opens that one, and the entry under it is the cursor.
func (b *Bar) PointerOver(heading, entry int) {
	b.hover = None
	if b.open < 0 {
		b.hover = heading

		return
	}
	if heading >= 0 && heading != b.open {
		b.open, b.menuCursor = heading, heading
	}
	b.cursor = entry
}

// ChooseShortcut chooses the entry whose shortcut is the key named name (see ShortcutKeys), with
// the box open or not. Greyed entries are skipped.
func (b *Bar) ChooseShortcut(name string) (Selection, bool) {
	for i, mn := range b.menus {
		for j, e := range mn.Entries {
			if e.Separator || e.Shortcut == "" || b.off(e) {
				continue
			}
			for _, k := range shortcutKeys(e.Shortcut) {
				if k == name {
					return b.choose(i, j)
				}
			}
		}
	}

	return Selection{}, false
}

// Activate gives the bar the keyboard and opens the menu it was on (F10).
func (b *Bar) Activate() {
	if len(b.menus) == 0 {
		return
	}
	b.SetOpen(b.menuCursor)
	b.active = true
}

// OpenByLetter opens the menu whose name begins with letter (Alt and a letter). It says whether
// the bar has menus, which is whether the key was for it.
func (b *Bar) OpenByLetter(letter string) bool {
	if len(b.menus) == 0 {
		return false
	}
	for i, mn := range b.menus {
		if first := []rune(strings.ToLower(strings.TrimSpace(mn.Name))); len(first) > 0 && string(first[0]) == letter {
			b.SetOpen(i)
			b.active = true

			return true
		}
	}

	return true
}

// Engaged says whether the keyboard has the bar: it was taken, or a box is open. Until then the
// arrow keys and the like are left to the program.
func (b Bar) Engaged() bool { return len(b.menus) > 0 && (b.active || b.open >= 0) }

// MoveHeading moves the keyboard to the next heading (dir 1) or the previous (dir -1), wrapping;
// an open box follows.
func (b *Bar) MoveHeading(dir int) {
	if !b.Engaged() {
		return
	}
	next := ((b.menuCursor+dir)%len(b.menus) + len(b.menus)) % len(b.menus)
	b.menuCursor = next
	if b.open >= 0 {
		b.SetOpen(next)
	}
}

// MoveEntry moves down (dir 1) or up (dir -1) the open menu, skipping what cannot be chosen. With
// no box open, down opens the one the keyboard is on.
func (b *Bar) MoveEntry(dir int) {
	if !b.Engaged() {
		return
	}
	if b.open < 0 {
		if dir > 0 {
			b.SetOpen(b.menuCursor)
		}

		return
	}
	from := b.cursor
	if from < 0 && dir < 0 {
		from = 0
	}
	b.cursor = b.step(from, dir)
}

// First moves to the first entry that can be chosen in the open menu.
func (b *Bar) First() {
	if b.Engaged() && b.open >= 0 {
		b.cursor = b.step(None, 1)
	}
}

// Last moves to the last entry that can be chosen in the open menu.
func (b *Bar) Last() {
	if b.Engaged() && b.open >= 0 {
		b.cursor = b.step(0, -1)
	}
}

// Confirm is Enter or Space: with no box open it opens the one the keyboard is on, else it
// chooses the entry the cursor is on.
func (b *Bar) Confirm() (Selection, bool) {
	if !b.Engaged() {
		return Selection{}, false
	}
	if b.open < 0 {
		b.SetOpen(b.menuCursor)

		return Selection{}, false
	}
	if b.cursor >= 0 {
		return b.choose(b.open, b.cursor)
	}

	return Selection{}, false
}

// Escape closes the open box, and with none open gives the keyboard back to the program.
func (b *Bar) Escape() {
	if !b.Engaged() {
		return
	}
	if b.open >= 0 {
		b.open, b.cursor = None, None
	} else {
		b.active = false
	}
}

// step is the next entry from index from in direction dir that can be chosen: separators are
// skipped and it wraps. It is None when there is none.
func (b Bar) step(from, dir int) int {
	if b.open < 0 {
		return None
	}
	entries := b.menus[b.open].Entries
	n := len(entries)
	for k := 1; k <= n; k++ {
		i := ((from+dir*k)%n + n) % n
		if !entries[i].Separator && !b.off(entries[i]) {
			return i
		}
	}

	return None
}

// choose closes the bar and reports entry j of menu i.
func (b *Bar) choose(i, j int) (Selection, bool) {
	e := b.menus[i].Entries[j]
	if b.off(e) {
		return Selection{}, false
	}
	b.Close()
	b.menuCursor = i

	return Selection{Menu: i, Entry: j, MenuName: b.menus[i].Name, Name: e.Name, Shortcut: e.Shortcut}, true
}
