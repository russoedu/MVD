package prefspopup

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/look"
)

// Model is the preferences popup: a tab at a time, and the setting the cursor is on.
// A change is in Config as soon as it is made; the popup writes nothing itself.
type Model struct {
	cfg  config.Config
	host Host

	pages   []page
	tab     int
	cursor  int // the field of the page
	scroll  int // the first line of the page that shows
	editing bool
	input   textinput.Model
	// before is the text of the field when typing in it began, for esc to put back.
	before string

	busy      bool // a window of the host's is open and has the answer
	status    string
	statusBad bool
	w, h      int
}

// New returns the popup for cfg, on the first tab.
func New(cfg config.Config, host Host) Model {
	in := textinput.New()
	in.Prompt = "> "
	return Model{cfg: cfg, host: host, pages: pages(host), input: in}
}

// Config is the settings as changed so far.
func (m Model) Config() config.Config { return m.cfg }

// Tab is the index of the tab shown.
func (m Model) Tab() int { return m.tab }

// Editing reports whether a text is being typed.
func (m Model) Editing() bool { return m.editing }

// SetSize gives the popup w by h cells.
func (m Model) SetSize(w, h int) Model {
	m.w, m.h = w, h
	m.input.SetWidth(max(10, w-8))
	return m.reveal()
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) current() field { return m.pages[m.tab].fields[m.cursor] }

// show puts the cursor on the first field of tab.
func (m Model) show(tab int) Model {
	m.tab = (tab + len(m.pages)) % len(m.pages)
	m.cursor, m.scroll, m.editing, m.status = 0, 0, false, ""
	return m.reveal()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pickedMsg:
		return m.picked(msg)
	case doneMsg:
		m.busy, m.status, m.statusBad = false, msg.message, msg.err != nil
		if msg.err != nil {
			m.status = msg.err.Error()
		}
		return m, nil
	case tea.KeyPressMsg:
		if m.busy {
			return m, nil
		}
		if m.editing {
			return m.typing(msg)
		}
		return m.key(msg)
	case tea.MouseMsg:
		if m.busy {
			return m, nil
		}
		return m.mouse(msg)
	}
	if m.editing {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) key(k tea.KeyPressMsg) (Model, tea.Cmd) {
	m.status = ""
	switch k.String() {
	case "up", "k":
		m.cursor = max(m.cursor-1, 0)
	case "down", "j":
		m.cursor = min(m.cursor+1, len(m.pages[m.tab].fields)-1)
	case "left", "h":
		m = m.change(-1)
	case "right", "l":
		m = m.change(1)
	case "tab":
		return m.show(m.tab + 1), nil
	case "shift+tab":
		return m.show(m.tab - 1), nil
	case "enter", " ":
		return m.activate()
	case "d":
		if m.current().kind == kindColour {
			m = m.useDefaultColour()
		}
	case "s", "ctrl+s":
		return m, func() tea.Msg { return SaveMsg{} }
	case "esc":
		return m, func() tea.Msg { return CloseMsg{} }
	}
	return m.reveal(), nil
}

// change moves the setting one option or one step: -1 back, 1 forward.
func (m Model) change(by int) Model {
	f := m.current()
	value := f.get(m.cfg)
	switch f.kind {
	case kindChoice:
		at := 0
		for i, o := range f.options {
			if o == value {
				at = i
			}
		}
		if next := at + by; next >= 0 && next < len(f.options) {
			f.set(&m.cfg, f.options[next])
		}
	case kindStepper:
		n := min(max(parseInt(value)+by*f.step, f.min), f.max)
		f.set(&m.cfg, intText(n))
	}
	return m
}

// activate does what enter does on the field: another option, text to type, a folder
// to choose, or the action.
func (m Model) activate() (Model, tea.Cmd) {
	f := m.current()
	switch f.kind {
	case kindChoice:
		at := 0
		for i, o := range f.options {
			if o == f.get(m.cfg) {
				at = i
			}
		}
		f.set(&m.cfg, f.options[(at+1)%len(f.options)])
	case kindText, kindTemplate, kindColour:
		return m.edit(), textinput.Blink
	case kindFolder:
		if m.host.PickFolder == nil {
			return m.edit(), textinput.Blink
		}
		start, pick, field := f.get(m.cfg), m.host.PickFolder, m.cursor
		m.busy, m.status, m.statusBad = true, "Choose a folder in the window that opened...", false
		return m, func() tea.Msg {
			path, chosen, err := pick(start)
			return pickedMsg{field: field, path: path, chosen: chosen, err: err}
		}
	case kindAction:
		return m.act(f.action)
	}
	return m, nil
}

// act asks the host for the action of a button.
func (m Model) act(a actionKind) (Model, tea.Cmd) {
	var run func() (string, error)
	switch a {
	case actionMove:
		run = m.host.Move
	case actionUninstall:
		run = m.host.Uninstall
	}
	if run == nil {
		return m, nil
	}
	m.busy, m.status, m.statusBad = true, "Answer the questions in the window that opened...", false
	return m, func() tea.Msg {
		message, err := run()
		return doneMsg{message: message, err: err}
	}
}

// picked takes the answer of the folder chooser. When it could not be shown, the path is
// typed instead.
func (m Model) picked(p pickedMsg) (Model, tea.Cmd) {
	m.busy, m.status = false, ""
	m.cursor = p.field
	switch {
	case p.err != nil:
		return m.edit(), textinput.Blink
	case p.chosen:
		m.current().set(&m.cfg, p.path)
	}
	return m, nil
}

// edit starts typing the value of the field the cursor is on.
func (m Model) edit() Model {
	m.before = m.current().get(m.cfg)
	m.input.SetValue(m.before)
	m.input.CursorEnd()
	m.input.Focus()
	m.editing, m.status = true, ""
	return m
}

func (m Model) typing(k tea.KeyPressMsg) (Model, tea.Cmd) {
	switch k.String() {
	case "esc":
		if f := m.current(); f.kind == kindTemplate {
			f.set(&m.cfg, m.before)
		}
		m.editing, m.status = false, ""
		return m, nil
	case "enter":
		return m.accept(), nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(k)
	return m.live(), cmd
}

// live keeps the name pattern of the files as typed, so the example under it follows every
// key; the other fields wait for enter.
func (m Model) live() Model {
	if f := m.current(); f.kind == kindTemplate {
		f.set(&m.cfg, m.input.Value())
	}
	return m
}

// accept takes what was typed as the value of the field.
func (m Model) accept() Model {
	f := m.current()
	text := strings.TrimSpace(m.input.Value())
	switch f.kind {
	case kindColour:
		if !config.ValidColour(text) {
			m.status, m.statusBad = "Use a colour like #ff007f or #f07.", true
			return m
		}
		f.set(&m.cfg, text)
		look.Apply(m.cfg.Colors)
	case kindText:
		f.set(&m.cfg, text)
	default:
		if text != "" {
			f.set(&m.cfg, text)
		}
	}
	m.editing, m.status = false, ""
	return m
}

// useDefaultColour puts the colour the cursor is on back to its default.
func (m Model) useDefaultColour() Model {
	d := config.DefaultThemeColors()
	defaults := []string{d.Accent, d.Focus, d.Highlight, d.Success, d.Error, d.Dim, d.Text, d.Selected}
	m.current().set(&m.cfg, defaults[m.cursor])
	look.Apply(m.cfg.Colors)
	return m
}

// addBlock puts a block of the template into the template, where the cursor is: typing in
// it starts when it was not, with the cursor at the end.
func (m Model) addBlock(index int) Model {
	f := m.current()
	if f.kind != kindTemplate || index < 0 || index >= len(templateBlocks) {
		return m
	}
	if !m.editing {
		m = m.edit()
	}
	token := templateBlocks[index].token
	text, at := []rune(m.input.Value()), m.input.Position()
	at = max(min(at, len(text)), 0)
	m.input.SetValue(string(text[:at]) + token + string(text[at:]))
	m.input.SetCursor(at + len([]rune(token)))
	return m.live()
}

// mouse works out a click on a tab or on the page, or the wheel over it. The position is
// in the cells of the popup.
func (m Model) mouse(msg tea.MouseMsg) (Model, tea.Cmd) {
	pos := msg.Mouse()
	pos.X-- // the page is inside the border of its box
	if wheel, ok := msg.(tea.MouseWheelMsg); ok {
		by := 3
		if wheel.Button == tea.MouseWheelUp {
			by = -3
		}
		l := m.layout()
		m.scroll = max(min(m.scroll+by, len(l.lines)-m.bodyRows()), 0)
		return m, nil
	}
	click, ok := msg.(tea.MouseClickMsg)
	if !ok || click.Button != tea.MouseLeft {
		return m, nil
	}
	if pos.Y < tabRows {
		for i, s := range m.tabSpans() {
			if pos.X+1 >= s.first && pos.X+1 < s.last {
				return m.finishTyping().show(i), nil
			}
		}
		return m, nil
	}
	row := pos.Y - tabRows
	if row >= m.bodyRows() {
		return m, nil
	}
	line := row + m.scroll
	l := m.layout()
	m.status = ""
	for _, h := range l.hits {
		if h.y != line || pos.X < h.x0 || pos.X >= h.x1 {
			continue
		}
		if m.editing && h.field == m.cursor && m.current().kind == kindTemplate {
			// In the template the cursor goes where the pointer is, and a block goes in there.
			switch h.kind {
			case hitValue:
				m.input.SetCursor(max(pos.X-gutterWidth-promptWidth, 0))
				return m, nil
			case hitBlock:
				return m.addBlock(h.arg), nil
			}
		}
		if m.editing {
			if m = m.accept(); m.editing {
				return m, nil
			}
		}
		m.cursor = h.field
		if h.kind == hitValue && m.current().kind == kindTemplate {
			m, cmd := m.hit(h)
			m.input.SetCursor(max(pos.X-gutterWidth-promptWidth, 0))
			return m, cmd
		}
		return m.hit(h)
	}
	if m.editing {
		if m = m.accept(); m.editing {
			return m, nil
		}
	}
	for i, s := range l.spans {
		if line >= s.first && line <= s.last {
			m.cursor = i
		}
	}
	return m, nil
}

// promptWidth is the columns of the prompt in front of what is typed.
const promptWidth = 2

// finishTyping takes what is typed as the value, if it is a value; typing goes on if it is not.
func (m Model) finishTyping() Model {
	if m.editing {
		return m.accept()
	}
	return m
}

// hit does what a click on a part of the page does.
func (m Model) hit(h hit) (Model, tea.Cmd) {
	f := m.current()
	switch h.kind {
	case hitOption:
		f.set(&m.cfg, f.options[h.arg])
	case hitMinus:
		m = m.change(-1)
	case hitPlus:
		m = m.change(1)
	case hitBlock:
		m = m.addBlock(h.arg)
	case hitValue:
		return m.activate()
	}
	return m, nil
}

// WithConfig replaces the settings being edited.
func (m Model) WithConfig(cfg config.Config) Model {
	m.cfg = cfg
	return m
}
