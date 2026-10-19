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
)

// padX is the space at each side of an entry, inside the box.
const padX = 2

// Model is the menu bar component.
type Model struct {
	menus []Menu
	title string
	// open is the menu whose box is showing, or -1.
	open int
	// cursor is the entry the keyboard or the pointer is on in the open menu,
	// or -1; menuCursor is the heading the keyboard is on.
	cursor, menuCursor int
	// active is true once the keyboard has taken the bar (F10, an Alt letter or
	// a press): until then the arrow keys are left to the program.
	active              bool
	// disabled are the entries, by name, that are greyed out and cannot be chosen.
	disabled map[string]bool
	w, h                int
	originX, originY    int
	accent, muted, rule string
}

// Option configures New.
type Option func(*Model)

// New returns a closed bar with no menus.
func New(opts ...Option) Model {
	m := Model{open: -1, cursor: -1, w: 40, h: 8, accent: "212", muted: "245", rule: "240", disabled: map[string]bool{}}
	for _, o := range opts {
		o(&m)
	}
	m.open = m.validMenu(m.open)
	m.menuCursor = max(m.open, 0)
	return m
}

// WithMenus sets the menus.
func WithMenus(menus ...Menu) Option { return func(m *Model) { m.menus = menus } }

// WithMenusJSON sets the menus from the text a design file carries them as
// (see ParseMenus). Text that does not parse leaves the menus empty.
func WithMenusJSON(text string) Option {
	return func(m *Model) {
		menus, err := ParseMenus(text)
		if err != nil {
			menus = nil
		}
		m.menus = menus
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
func (m *Model) SetMenus(menus []Menu) {
	m.menus = menus
	m.Close()
	m.menuCursor = 0
}

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
func (m *Model) SetEnabled(name string, enabled bool) {
	if m.disabled == nil {
		m.disabled = map[string]bool{}
	}
	m.disabled[name] = !enabled
}

func (m Model) off(e Entry) bool { return m.disabled[e.Name] }

// SetTitle changes the text at the left of the bar.
func (m *Model) SetTitle(title string) { m.title = title }

// SetSize sets the room the component takes.
func (m *Model) SetSize(w, h int) { m.w, m.h = w, h }

// SetOrigin tells the component where on the screen it is drawn.
func (m *Model) SetOrigin(x, y int) { m.originX, m.originY = x, y }

// SetOpen opens menu i, or closes the bar for -1 or a number there is no menu
// for.
func (m *Model) SetOpen(i int) {
	m.open = m.validMenu(i)
	m.cursor = -1
	if m.open >= 0 {
		m.menuCursor = m.open
		m.cursor = m.step(-1, 1)
	}
}

// Close hides the box and gives the keyboard back to the program.
func (m *Model) Close() {
	m.open, m.cursor, m.active = -1, -1, false
}

// Open is the menu showing, or -1.
func (m Model) Open() int { return m.open }

// Menus are the menus.
func (m Model) Menus() []Menu { return m.menus }

// MenusJSON is the menus as WithMenusJSON reads them.
func (m Model) MenusJSON() string { return MenusJSON(m.menus) }

// Width and Height are the room the component takes.
func (m Model) Width() int  { return m.w }
func (m Model) Height() int { return m.h }

// ContentHeight is the rows the bar and the box of its tallest menu need.
func (m Model) ContentHeight() int {
	tallest := 0
	for _, mn := range m.menus {
		tallest = max(tallest, len(mn.Entries))
	}
	if tallest == 0 {
		return 1
	}
	return 1 + tallest + 2
}

func (m Model) validMenu(i int) int {
	if i < 0 || i >= len(m.menus) {
		return -1
	}
	return i
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
		if m.open >= 0 {
			x, y := msg.X-m.originX, msg.Y-m.originY
			if y == 0 {
				if i := m.headingAt(x); i >= 0 && i != m.open {
					m.open, m.menuCursor = i, i
				}
			}
			m.cursor = m.entryAt(x, y)
		}
		return m, nil
	case tea.KeyPressMsg:
		return m.key(msg.String())
	}
	return m, nil
}

func (m Model) click(x, y int) (Model, tea.Cmd) {
	if y == 0 {
		if i := m.headingAt(x); i >= 0 {
			if m.open == i {
				m.Close()
			} else {
				m.SetOpen(i)
				m.active = true
				m.cursor = -1
			}
			return m, nil
		}
	}
	if m.open >= 0 {
		if j := m.entryAt(x, y); j >= 0 {
			return m.choose(m.open, j)
		}
		m.Close()
	}
	return m, nil
}

func (m Model) key(name string) (Model, tea.Cmd) {
	// Shortcuts work whether the menu is open or not.
	for i, mn := range m.menus {
		for j, e := range mn.Entries {
			if e.Separator || e.Shortcut == "" || m.off(e) {
				continue
			}
			for _, k := range shortcutKeys(e.Shortcut) {
				if k == name {
					return m.choose(i, j)
				}
			}
		}
	}
	if len(m.menus) == 0 {
		return m, nil
	}
	if name == "f10" {
		m.SetOpen(m.menuCursor)
		m.active = true
		return m, nil
	}
	if letter, ok := strings.CutPrefix(name, "alt+"); ok && len([]rune(letter)) == 1 {
		for i, mn := range m.menus {
			if first := []rune(strings.ToLower(strings.TrimSpace(mn.Name))); len(first) > 0 && string(first[0]) == letter {
				m.SetOpen(i)
				m.active = true
				return m, nil
			}
		}
		return m, nil
	}
	if !m.active && m.open < 0 {
		return m, nil
	}
	switch name {
	case "left", "right":
		dir := 1
		if name == "left" {
			dir = -1
		}
		next := ((m.menuCursor+dir)%len(m.menus) + len(m.menus)) % len(m.menus)
		m.menuCursor = next
		if m.open >= 0 {
			m.SetOpen(next)
		}
	case "down", "up":
		if m.open < 0 {
			if name == "down" {
				m.SetOpen(m.menuCursor)
			}
			break
		}
		dir := 1
		if name == "up" {
			dir = -1
		}
		from := m.cursor
		if from < 0 && dir < 0 {
			from = 0
		}
		m.cursor = m.step(from, dir)
	case "home":
		if m.open >= 0 {
			m.cursor = m.step(-1, 1)
		}
	case "end":
		if m.open >= 0 {
			m.cursor = m.step(0, -1)
		}
	case "enter", "space":
		if m.open < 0 {
			m.SetOpen(m.menuCursor)
			break
		}
		if m.cursor >= 0 {
			return m.choose(m.open, m.cursor)
		}
	case "esc":
		if m.open >= 0 {
			m.open, m.cursor = -1, -1
		} else {
			m.active = false
		}
	}
	return m, nil
}

// step is the next entry from index from in direction dir that can be chosen:
// separators are skipped and it wraps. It is -1 when there is none.
func (m Model) step(from, dir int) int {
	if m.open < 0 {
		return -1
	}
	entries := m.menus[m.open].Entries
	n := len(entries)
	for k := 1; k <= n; k++ {
		i := ((from+dir*k)%n + n) % n
		if !entries[i].Separator && !m.off(entries[i]) {
			return i
		}
	}
	return -1
}

// choose closes the bar and reports entry j of menu i.
func (m Model) choose(i, j int) (Model, tea.Cmd) {
	e := m.menus[i].Entries[j]
	if m.off(e) {
		return m, nil
	}
	m.Close()
	m.menuCursor = i
	out := SelectedMsg{Menu: i, Entry: j, MenuName: m.menus[i].Name, Name: e.Name, Shortcut: e.Shortcut}
	return m, func() tea.Msg { return out }
}

// bar geometry -------------------------------------------------------------

// headingX are the first columns of each heading's label.
func (m Model) headingX() []int {
	x := 1
	if m.title != "" {
		x += lipgloss.Width(m.title) + 1 + 2
	}
	xs := make([]int, len(m.menus))
	for i, mn := range m.menus {
		xs[i] = x
		x += lipgloss.Width(mn.Name) + 3
	}
	return xs
}

// headingAt is the heading under column x, or -1.
func (m Model) headingAt(x int) int {
	for i, x0 := range m.headingX() {
		if x >= x0 && x < x0+lipgloss.Width(m.menus[i].Name)+2 {
			return i
		}
	}
	return -1
}

// boxWidth is the inner width of the open menu's box.
func (m Model) boxWidth() int {
	w := 0
	for _, e := range m.menus[m.open].Entries {
		n := lipgloss.Width(e.Name)
		if e.Shortcut != "" {
			n += 3 + lipgloss.Width(e.Shortcut)
		}
		w = max(w, n)
	}
	return w + 2*padX
}

// boxX is the column of the box's left edge: under its heading, kept inside
// the component.
func (m Model) boxX() int {
	return min(m.headingX()[m.open], max(m.w-m.boxWidth()-2, 0))
}

// entryAt is the entry of the open menu under (x, y), or -1: row 0 is the bar,
// row 1 the top edge of the box.
func (m Model) entryAt(x, y int) int {
	if m.open < 0 {
		return -1
	}
	j := y - 2
	left := m.boxX()
	if x <= left || x >= left+m.boxWidth()+1 || j < 0 || j >= len(m.menus[m.open].Entries) || j >= m.h-2 {
		return -1
	}
	if e := m.menus[m.open].Entries[j]; e.Separator || m.off(e) {
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
	for i, mn := range m.menus {
		label := " " + mn.Name + " "
		if i == m.open || (i == m.menuCursor && m.active) {
			bar.WriteString(selected.Render(label))
		} else {
			bar.WriteString(muted.Render(label))
		}
		bar.WriteString(" ")
	}
	rows[0] = fit(bar.String(), m.w)

	if m.open >= 0 && m.h > 1 {
		x, w := m.boxX(), m.boxWidth()
		box := []string{rule.Render("╭" + strings.Repeat("─", w) + "╮")}
		for j, e := range m.menus[m.open].Entries {
			if e.Separator {
				box = append(box, rule.Render("│"+strings.Repeat("─", w)+"│"))
				continue
			}
			gap := max(w-2*padX-lipgloss.Width(e.Name)-lipgloss.Width(e.Shortcut), 1)
			pad := strings.Repeat(" ", padX)
			var text string
			if m.off(e) {
				text = muted.Faint(true).Render(pad + e.Name + strings.Repeat(" ", gap) + e.Shortcut + pad)
			} else if j == m.cursor {
				text = selected.Render(pad + e.Name + strings.Repeat(" ", gap) + e.Shortcut + pad)
			} else {
				text = pad + e.Name + strings.Repeat(" ", gap) + muted.Render(e.Shortcut) + pad
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
