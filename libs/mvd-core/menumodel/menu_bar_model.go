package menumodel

// None is the value of an open menu, a cursor or a hover that is not on anything.
const None = -1

// Bar is the state of the menu bar. It is a value; a copy keeps its own state, so the maps are
// replaced and never changed in place.
type Bar struct {
	menus []Menu
	// open is the menu whose box is showing, or None.
	open int
	// cursor is the entry the keyboard or the pointer is on in the open menu, or None;
	// menuCursor is the heading the keyboard is on.
	cursor, menuCursor int
	// active is true once the keyboard has taken the bar (F10, an Alt letter or a press): until
	// then the arrow keys are left to the program.
	active bool
	// disabled are the entries, by name, that are greyed out and cannot be chosen.
	disabled map[string]bool
	// hover is the heading the pointer is over while no menu is open, or None.
	hover int
}

// NewBar returns a closed bar with no menus.
func NewBar() Bar {
	return Bar{open: None, cursor: None, hover: None}
}

// Menus are the menus.
func (b Bar) Menus() []Menu { return b.menus }

// Open is the menu showing, or None.
func (b Bar) Open() int { return b.open }

// Cursor is the entry the keyboard or the pointer is on in the open menu, or None.
func (b Bar) Cursor() int { return b.cursor }

// MenuCursor is the heading the keyboard is on.
func (b Bar) MenuCursor() int { return b.menuCursor }

// Active says whether the keyboard has taken the bar.
func (b Bar) Active() bool { return b.active }

// Hover is the heading the pointer is over while no menu is open, or None.
func (b Bar) Hover() int { return b.hover }

// Enabled reports whether the entry with the given name can be chosen.
func (b Bar) Enabled(name string) bool { return !b.disabled[name] }

// off says whether an entry is greyed out.
func (b Bar) off(e Entry) bool { return b.disabled[e.Name] }

// Off says whether an entry is greyed out, for a renderer that draws it so.
func (b Bar) Off(e Entry) bool { return b.off(e) }

// WithMenus is the bar with these menus, closed. Menu i is open when open is i.
func (b Bar) WithMenus(menus []Menu, open int) Bar {
	b.menus = menus
	b.open = b.validMenu(open)
	b.cursor, b.active = None, false
	b.menuCursor = max(b.open, 0)

	return b
}

// SetMenus replaces the menus and closes the bar.
func (b *Bar) SetMenus(menus []Menu) {
	b.menus = menus
	b.Close()
	b.menuCursor = 0
}

// SetEnabled greys out the entry with the given name, or brings it back. A greyed entry is
// skipped by the arrow keys and chooses nothing, by pointer or by shortcut.
func (b *Bar) SetEnabled(name string, enabled bool) {
	disabled := make(map[string]bool, len(b.disabled)+1)
	for k, v := range b.disabled {
		disabled[k] = v
	}
	disabled[name] = !enabled
	b.disabled = disabled
}

func (b Bar) validMenu(i int) int {
	if i < 0 || i >= len(b.menus) {
		return None
	}

	return i
}
