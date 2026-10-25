// Package menubar is the menu bar of the app: headings along the top (File, Edit,
// Help), and under the one that is open a box of entries, each with an optional
// shortcut, and separators between groups. It began as the menu bar component of
// Cuppa (https://github.com/meta-ui-labs/cuppa, MIT) and is kept here, so nothing
// of the build depends on Cuppa.
//
//	bar := menubar.New(
//	    menubar.WithMenus(
//	        menubar.Menu{Name: "File", Entries: []menubar.Entry{
//	            {Name: "Open…", Shortcut: "Ctrl+O"},
//	            {Separator: true},
//	            {Name: "Quit", Shortcut: "Ctrl+Q"},
//	        }},
//	    ),
//	    menubar.WithSize(40, 8),
//	)
//
// The bar is one row. The box of the open menu is drawn on the rows under it,
// inside the room the component is given, so give it the height of the bar and
// of its tallest menu (see ContentHeight).
package menubar

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/menumodel"
)

// Entry and Menu are what the menus are made of; the menus, which one is open and what choosing
// an entry does are in package menumodel.
type (
	Entry = menumodel.Entry
	Menu  = menumodel.Menu
)

// ParseMenus reads menus from the text a design file carries them as (see menumodel.ParseMenus).
func ParseMenus(text string) ([]Menu, error) { return menumodel.ParseMenus(text) }

// MenusJSON writes menus as ParseMenus reads them.
func MenusJSON(menus []Menu) string { return menumodel.MenusJSON(menus) }

// ShortcutLabel is a shortcut as the person reads it (see menumodel.ShortcutLabel).
func ShortcutLabel(text string) string { return menumodel.ShortcutLabel(text) }

// SelectedMsg is sent when an entry is chosen with the mouse or the keyboard,
// or by pressing its shortcut.
type SelectedMsg struct {
	// Menu and Entry are the position of the entry: the menu in the bar and the
	// entry in that menu, both from 0.
	Menu, Entry int
	// MenuName, Name and Shortcut are what the entry says.
	MenuName, Name, Shortcut string
}

// padX is the space at each side of an entry, inside the box.
const padX = 2

// Model is the menu bar component: the bar's state is a menumodel.Bar, and this draws it and
// turns the pointer and the keys into its intents.
type Model struct {
	bar   menumodel.Bar
	title string
	// open is what WithOpen asked for, until New gives the bar its menus.
	open                int
	w, h                int
	originX, originY    int
	accent, muted, rule string
}

// Option configures New.
type Option func(*Model)

// New returns a closed bar with no menus.
func New(opts ...Option) Model {
	m := Model{bar: menumodel.NewBar(), open: menumodel.None, w: 40, h: 8, accent: "212", muted: "245", rule: "240"}
	for _, o := range opts {
		o(&m)
	}
	m.bar = m.bar.WithMenus(m.bar.Menus(), m.open)

	return m
}

// WithMenus sets the menus.
func WithMenus(menus ...Menu) Option {
	return func(m *Model) { m.bar = m.bar.WithMenus(menus, menumodel.None) }
}

// WithMenusJSON sets the menus from the text a design file carries them as
// (see ParseMenus). Text that does not parse leaves the menus empty.
func WithMenusJSON(text string) Option {
	return func(m *Model) {
		menus, err := ParseMenus(text)
		if err != nil {
			menus = nil
		}
		m.bar = m.bar.WithMenus(menus, menumodel.None)
	}
}

// WithTitle sets the text at the left of the bar, before the menus.
func WithTitle(title string) Option { return func(m *Model) { m.title = title } }

// WithOpen shows menu i open, from 0; -1 shows none.
func WithOpen(i int) Option { return func(m *Model) { m.open = i } }

// WithSize sets the room the component takes, in cells.
func WithSize(w, h int) Option { return func(m *Model) { m.w, m.h = w, h } }

// WithOrigin tells the component where on the screen it is drawn, so mouse
// clicks line up with its headings.
func WithOrigin(x, y int) Option { return func(m *Model) { m.originX, m.originY = x, y } }

// WithColours sets the accent (the open heading and the highlighted entry), the
// muted colour (the other headings and the shortcuts) and the rule colour (the
// box). Empty keeps the current one.
func WithColours(accent, muted, rule string) Option {
	return func(m *Model) {
		if accent != "" {
			m.accent = accent
		}
		if muted != "" {
			m.muted = muted
		}
		if rule != "" {
			m.rule = rule
		}
	}
}

// WithAccent, WithMuted and WithRule set one colour each.
func WithAccent(c string) Option { return func(m *Model) { m.accent = c } }
func WithMuted(c string) Option  { return func(m *Model) { m.muted = c } }
func WithRule(c string) Option   { return func(m *Model) { m.rule = c } }

// SetAccent, SetMuted and SetRule change the colours later.
func (m *Model) SetAccent(c string) { m.accent = c }
func (m *Model) SetMuted(c string)  { m.muted = c }
func (m *Model) SetRule(c string)   { m.rule = c }

// SetMenus replaces the menus and closes the bar.
func (m *Model) SetMenus(menus []Menu) { m.bar.SetMenus(menus) }

// SetMenusJSON replaces the menus from the text a design file carries them as
// and reports whether it parsed; the menus are left as they were when it did
// not.
func (m *Model) SetMenusJSON(text string) bool {
	menus, err := ParseMenus(text)
	if err != nil {
		return false
	}
	m.SetMenus(menus)

	return true
}

// SetEnabled greys out the entry with the given name, or brings it back. A greyed entry
// is skipped by the arrow keys and chooses nothing, by pointer or by shortcut.
func (m *Model) SetEnabled(name string, enabled bool) { m.bar.SetEnabled(name, enabled) }

// Enabled reports whether the entry with the given name can be chosen.
func (m Model) Enabled(name string) bool { return m.bar.Enabled(name) }

// SetTitle changes the text at the left of the bar.
func (m *Model) SetTitle(title string) { m.title = title }

// SetSize sets the room the component takes.
func (m *Model) SetSize(w, h int) { m.w, m.h = w, h }

// SetOrigin tells the component where on the screen it is drawn.
func (m *Model) SetOrigin(x, y int) { m.originX, m.originY = x, y }

// SetOpen opens menu i, or closes the bar for -1 or a number there is no menu
// for.
func (m *Model) SetOpen(i int) { m.bar.SetOpen(i) }

// Close hides the box and gives the keyboard back to the program.
func (m *Model) Close() { m.bar.Close() }

// Open is the menu showing, or -1.
func (m Model) Open() int { return m.bar.Open() }

// Menus are the menus.
func (m Model) Menus() []Menu { return m.bar.Menus() }

// MenusJSON is the menus as WithMenusJSON reads them.
func (m Model) MenusJSON() string { return MenusJSON(m.bar.Menus()) }

// Width and Height are the room the component takes.
func (m Model) Width() int  { return m.w }
func (m Model) Height() int { return m.h }

// ContentHeight is the rows the bar and the box of its tallest menu need.
func (m Model) ContentHeight() int {
	tallest := 0
	for _, mn := range m.bar.Menus() {
		tallest = max(tallest, len(mn.Entries))
	}
	if tallest == 0 {
		return 1
	}

	return 1 + tallest + 2
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update handles a click or a move of the mouse over the bar, and the keys:
// F10 or Alt and a heading's first letter open a menu, the arrows move through
// them, Enter chooses and Esc closes. The shortcut of an entry chooses it at
// any time. A choice arrives as a SelectedMsg.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft {
			return m, nil
		}

		return m.click(msg.X-m.originX, msg.Y-m.originY)
	case tea.MouseMotionMsg:
		x, y := msg.X-m.originX, msg.Y-m.originY
		heading := menumodel.None
		if y == 0 {
			heading = m.headingAt(x)
		}
		m.bar.PointerOver(heading, m.entryAt(x, y))

		return m, nil
	case tea.KeyPressMsg:
		return m.key(msg.String())
	}

	return m, nil
}

func (m Model) click(x, y int) (Model, tea.Cmd) {
	if y == 0 {
		if i := m.headingAt(x); i >= 0 {
			m.bar.ClickHeading(i)

			return m, nil
		}
	}
	if m.bar.Open() >= 0 {
		if j := m.entryAt(x, y); j >= 0 {
			sel, ok := m.bar.ClickEntry(j)

			return m.chosen(sel, ok)
		}
		m.bar.ClickAway()
	}

	return m, nil
}

func (m Model) key(name string) (Model, tea.Cmd) {
	// Shortcuts work whether the menu is open or not.
	if sel, ok := m.bar.ChooseShortcut(name); ok {
		return m.chosen(sel, true)
	}
	if name == "f10" {
		m.bar.Activate()

		return m, nil
	}
	if letter, ok := strings.CutPrefix(name, "alt+"); ok && len([]rune(letter)) == 1 {
		m.bar.OpenByLetter(letter)

		return m, nil
	}
	switch name {
	case "left":
		m.bar.MoveHeading(-1)
	case "right":
		m.bar.MoveHeading(1)
	case "down":
		m.bar.MoveEntry(1)
	case "up":
		m.bar.MoveEntry(-1)
	case "home":
		m.bar.First()
	case "end":
		m.bar.Last()
	case "enter", "space":
		sel, ok := m.bar.Confirm()

		return m.chosen(sel, ok)
	case "esc":
		m.bar.Escape()
	}

	return m, nil
}

// chosen is the bar after an entry was chosen, with the command that tells of it.
func (m Model) chosen(sel menumodel.Selection, ok bool) (Model, tea.Cmd) {
	if !ok {
		return m, nil
	}
	out := SelectedMsg{Menu: sel.Menu, Entry: sel.Entry, MenuName: sel.MenuName, Name: sel.Name, Shortcut: sel.Shortcut}

	return m, func() tea.Msg { return out }
}

// bar geometry -------------------------------------------------------------

// headingX are the first columns of each heading's label.
func (m Model) headingX() []int {
	x := 1
	if m.title != "" {
		x += lipgloss.Width(m.title) + 1 + 2
	}
	xs := make([]int, len(m.bar.Menus()))
	for i, mn := range m.bar.Menus() {
		xs[i] = x
		x += lipgloss.Width(mn.Name) + 3
	}
	return xs
}

// headingAt is the heading under column x, or -1.
func (m Model) headingAt(x int) int {
	for i, x0 := range m.headingX() {
		if x >= x0 && x < x0+lipgloss.Width(m.bar.Menus()[i].Name)+2 {
			return i
		}
	}
	return -1
}

// boxWidth is the inner width of the open menu's box.
func (m Model) boxWidth() int {
	w := 0
	for _, e := range m.bar.Menus()[m.bar.Open()].Entries {
		n := lipgloss.Width(e.Name)
		if e.Shortcut != "" {
			n += 3 + lipgloss.Width(ShortcutLabel(e.Shortcut))
		}
		w = max(w, n)
	}
	return w + 2*padX
}

// boxX is the column of the box's left edge: under its heading, kept inside
// the component.
func (m Model) boxX() int {
	return min(m.headingX()[m.bar.Open()], max(m.w-m.boxWidth()-2, 0))
}

// entryAt is the entry of the open menu under (x, y), or -1: row 0 is the bar,
// row 1 the top edge of the box.
func (m Model) entryAt(x, y int) int {
	if m.bar.Open() < 0 {
		return -1
	}
	j := y - 2
	left := m.boxX()
	if x <= left || x >= left+m.boxWidth()+1 || j < 0 || j >= len(m.bar.Menus()[m.bar.Open()].Entries) || j >= m.h-2 {
		return -1
	}
	if e := m.bar.Menus()[m.bar.Open()].Entries[j]; e.Separator || m.bar.Off(e) {
		return -1
	}
	return j
}

// View draws the bar on the first row and the open menu's box under it.
func (m Model) View() string {
	rows := make([]string, max(m.h, 1))
	for i := range rows {
		rows[i] = strings.Repeat(" ", m.w)
	}
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color(m.accent))
	selected := accent.Reverse(true)
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(m.muted))
	rule := lipgloss.NewStyle().Foreground(lipgloss.Color(m.rule))

	var bar strings.Builder
	bar.WriteString(" ")
	if m.title != "" {
		bar.WriteString(accent.Bold(true).Render(m.title) + " " + rule.Render("│") + " ")
	}
	for i, mn := range m.bar.Menus() {
		label := " " + mn.Name + " "
		if i == m.bar.Open() || (i == m.bar.MenuCursor() && m.bar.Active()) {
			bar.WriteString(selected.Render(label))
		} else if i == m.bar.Hover() {
			bar.WriteString(accent.Underline(true).Render(label))
		} else {
			bar.WriteString(muted.Render(label))
		}
		bar.WriteString(" ")
	}
	rows[0] = fit(bar.String(), m.w)

	if m.bar.Open() >= 0 && m.h > 1 {
		x, w := m.boxX(), m.boxWidth()
		box := []string{rule.Render("╭" + strings.Repeat("─", w) + "╮")}
		for j, e := range m.bar.Menus()[m.bar.Open()].Entries {
			if e.Separator {
				box = append(box, rule.Render("│"+strings.Repeat("─", w)+"│"))
				continue
			}
			shortcut := ShortcutLabel(e.Shortcut)
			gap := max(w-2*padX-lipgloss.Width(e.Name)-lipgloss.Width(shortcut), 1)
			pad := strings.Repeat(" ", padX)
			var text string
			if m.bar.Off(e) {
				text = muted.Faint(true).Render(pad + e.Name + strings.Repeat(" ", gap) + shortcut + pad)
			} else if j == m.bar.Cursor() {
				text = selected.Render(pad + e.Name + strings.Repeat(" ", gap) + shortcut + pad)
			} else {
				text = pad + e.Name + strings.Repeat(" ", gap) + muted.Render(shortcut) + pad
			}
			box = append(box, rule.Render("│")+text+rule.Render("│"))
		}
		box = append(box, rule.Render("╰"+strings.Repeat("─", w)+"╯"))
		for k, line := range box {
			if y := 1 + k; y < len(rows) {
				rows[y] = fit(strings.Repeat(" ", x)+line, m.w)
			}
		}
	}
	return strings.Join(rows, "\n")
}

// fit cuts or pads a line to exactly w cells.
func fit(s string, w int) string {
	if ansi.StringWidth(s) > w {
		return ansi.Truncate(s, w, "")
	}
	return s + strings.Repeat(" ", w-ansi.StringWidth(s))
}
