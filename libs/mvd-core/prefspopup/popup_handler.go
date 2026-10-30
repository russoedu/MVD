package prefspopup

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/prefsform"
)

// Model is the preferences popup: the form (package prefsform) holds the tabs, the settings and
// what the person is doing to them; this is the terminal's side of it, the text box, the scroll
// of the page, the pointer and the drawing. A change is in Config as soon as it is made; the
// popup writes nothing itself.
type Model struct {
	form   prefsform.Form
	host   Host
	scroll int // the first line of the page that shows
	input  textinput.Model
	// hx and hy are where the pointer last was, in the cells of the popup.
	hx, hy int
	// dragging is true while a block of the name pattern is carried, from boundary dragFrom to
	// dragTo.
	dragging         bool
	dragFrom, dragTo int

	w, h int
}

// New returns the popup for cfg, on the first tab.
func New(cfg config.Config, host Host) Model {
	in := textinput.New()
	in.Prompt = "> "

	return Model{form: prefsform.New(cfg, capabilities(host)), host: host, input: in, hx: -1, hy: -1}
}

// capabilities are what the host can do, as the form asks for it.
func capabilities(host Host) prefsform.Capabilities {
	state := ""
	if host.SignInState != nil {
		state = host.SignInState()
	}

	return prefsform.Capabilities{SignInState: state, CanPickFolder: host.PickFolder != nil, CanMove: host.Move != nil, CanUninstall: host.Uninstall != nil, CanSignIn: host.SignIn != nil, CanStyleWindow: host.StylesWindow}
}

// Config is the settings as changed so far.
func (m Model) Config() config.Config { return m.form.Cfg }

// Tab is the index of the tab shown.
func (m Model) Tab() int { return m.form.Tab }

// Editing reports whether a text is being typed.
func (m Model) Editing() bool { return m.form.Editing }

// SetSize gives the popup w by h cells.
func (m Model) SetSize(w, h int) Model {
	m.w, m.h = w, h
	m.input.SetWidth(max(10, w-8))

	return m.reveal()
}

// Init returns the command that makes the text cursor blink.
func (m Model) Init() tea.Cmd { return textinput.Blink }

// show puts the cursor on the first field of tab.
func (m Model) show(tab int) Model {
	m.form.ShowTab(tab)
	m.scroll = 0

	return m.reveal()
}

// Update handles msg and returns the next popup and its command. A key press ends the welcome state, and a picked theme shows at once.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if _, pressed := msg.(tea.KeyPressMsg); pressed {
		m.form.Welcome = false // the person has started to use it
	}
	next, cmd := m.update(msg)
	if look.Colours != next.form.Cfg.Colors {
		look.Apply(next.form.Cfg.Colors) // the screen shows a theme as soon as it is picked
	}

	return next, cmd
}

func (m Model) update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case WindowIntentMsg:
		return m.windowIntent(prefsform.WindowIntent(msg))
	case pickedMsg:
		return m.picked(msg)
	case doneMsg:
		m.form.ActionDone(msg.message, msg.err)

		return m, nil
	case tea.KeyPressMsg:
		if m.form.Busy {
			return m, nil
		}
		if m.form.Editing {
			return m.typing(msg)
		}

		return m.key(msg)
	case tea.MouseMsg:
		if m.form.Busy {
			return m, nil
		}

		return m.mouse(msg)
	}
	if m.form.Editing {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)

		return m, cmd
	}

	return m, nil
}

func (m Model) key(k tea.KeyPressMsg) (Model, tea.Cmd) {
	m.form.ClearStatus()
	switch k.String() {
	case "up", "k":
		m.form.MoveCursor(-1)
	case "down", "j":
		m.form.MoveCursor(1)
	case "left", "h":
		m.form.Change(-1)
	case "right", "l":
		m.form.Change(1)
	case "tab":
		return m.show(m.form.Tab + 1), nil
	case "shift+tab":
		return m.show(m.form.Tab - 1), nil
	case "enter", " ":
		return m.activate()
	case "d":
		if m.form.Current().Kind == prefsform.KindColour {
			m.form.UseDefaultColour()
		}
	case "s", "ctrl+s":
		return m, func() tea.Msg { return SaveMsg{} }
	case "esc":
		return m, func() tea.Msg { return CloseMsg{} }
	}

	return m.reveal(), nil
}

// activate does what enter does on the field, and runs what the form asks of the host.
func (m Model) activate() (Model, tea.Cmd) {
	was := m.form.Editing
	effect := m.form.Activate()

	return m.hostDoes(effect, was)
}

// hostDoes begins what the form asked: the text box when typing began, or the host's window.
func (m Model) hostDoes(effect prefsform.Effect, wasEditing bool) (Model, tea.Cmd) {
	switch e := effect.(type) {
	case prefsform.PickFolder:
		pick := m.host.PickFolder

		return m, func() tea.Msg {
			path, chosen, err := pick(e.Start)

			return pickedMsg{field: e.Field, path: path, chosen: chosen, err: err}
		}
	case prefsform.RunAction:
		run := m.host.Move
		switch e.Kind {
		case prefsform.ActionUninstall:
			run = m.host.Uninstall
		case prefsform.ActionSignIn:
			run = m.host.SignIn
		}

		return m, func() tea.Msg {
			message, err := run()

			return doneMsg{message: message, err: err}
		}
	}
	if m.form.Editing && !wasEditing {
		return m.startTyping(), textinput.Blink
	}

	return m, nil
}

// WindowIntentMsg is what the person did in a preferences window the page draws. It goes through
// the same steps as the keys, so both ways of showing the preferences behave alike.
type WindowIntentMsg prefsform.WindowIntent

// Window is the preferences as a window draws them.
func (m Model) Window() prefsform.Window { return m.form.Window() }

// windowIntent does what the person did in the window; the host's part (a folder chooser, an
// action) is run as a key would run it.
func (m Model) windowIntent(in prefsform.WindowIntent) (Model, tea.Cmd) {
	switch in.Kind {
	case prefsform.IntentSave:
		return m, func() tea.Msg { return SaveMsg{} }
	case prefsform.IntentClose:
		return m, func() tea.Msg { return CloseMsg{} }
	}
	was := m.form.Editing
	effect := m.form.Apply(in)
	if in.Kind == prefsform.IntentTab {
		m.scroll = 0
	}

	return m.hostDoes(effect, was)
}

// startTyping fills the text box with the value being changed.
func (m Model) startTyping() Model {
	if m.form.Current().Kind != prefsform.KindTemplate {
		m.input.SetValue(m.form.Before)
		m.input.CursorEnd()
		m.input.Focus()
	}

	return m
}

// picked takes the answer of the folder chooser.
func (m Model) picked(p pickedMsg) (Model, tea.Cmd) {
	was := m.form.Editing
	m.form.Picked(p.field, p.path, p.chosen, p.err)

	return m.hostDoes(nil, was)
}

func (m Model) typing(k tea.KeyPressMsg) (Model, tea.Cmd) {
	if m.form.Current().Kind == prefsform.KindTemplate {
		return m.typingTemplate(k), nil
	}
	switch k.String() {
	case "esc":
		m.form.CancelEdit()

		return m, nil
	case "enter":
		m.form.Accept(m.input.Value())

		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(k)

	return m, cmd
}

// typingTemplate works the keys on the pattern: text goes in at the cursor, which moves
// over characters and blocks, and backspace and delete take either out.
func (m Model) typingTemplate(k tea.KeyPressMsg) Model {
	switch k.String() {
	case "esc":
		m.form.TemplateCancel()
	case "enter":
		m.form.TemplateDone()
	case "left":
		m.form.TemplateCaretBy(-1)
	case "right":
		m.form.TemplateCaretBy(1)
	case "home":
		m.form.TemplateCaretTo(0)
	case "end":
		m.form.TemplateCaretTo(m.form.TemplateLen())
	case "backspace":
		m.form.TemplateBackspace()
	case "delete":
		m.form.TemplateDelete()
	default:
		m.form.TemplateInsert(k.Text)
	}

	return m
}

// acceptTyped takes what is typed as the value, if a value is being typed.
func (m Model) acceptTyped() Model {
	if m.form.Editing {
		m.form.Accept(m.input.Value())
	}

	return m
}

// mouse works out a click on a tab or on the page, a block being carried, or the wheel over
// it. The position is in the cells of the popup.
func (m Model) mouse(msg tea.MouseMsg) (Model, tea.Cmd) {
	pos := msg.Mouse()
	m.hx, m.hy = pos.X, pos.Y
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
	if m.dragging {
		return m.carry(msg, pos), nil
	}
	click, ok := msg.(tea.MouseClickMsg)
	if !ok || click.Button != tea.MouseLeft {
		return m, nil
	}
	if pos.Y < tabRows {
		for i, s := range m.tabSpans() {
			if pos.X+1 >= s.first && pos.X+1 < s.last {
				return m.acceptTyped().show(i), nil
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
	m.form.ClearStatus()
	for _, h := range l.hits {
		if h.y != line || pos.X < h.x0 || pos.X >= h.x1 {
			continue
		}
		if m.form.Editing && h.field != m.form.Cursor {
			if m = m.acceptTyped(); m.form.Editing {
				return m, nil
			}
		}
		m.form.SelectSetting(h.field)

		return m.hit(h)
	}
	if m.form.Editing {
		if m = m.acceptTyped(); m.form.Editing {
			return m, nil
		}
	}
	for i, s := range l.spans {
		if line >= s.first && line <= s.last {
			m.form.SelectSetting(i)
		}
	}

	return m, nil
}

// carry follows a block being carried: the pointer chooses the boundary it would land on,
// and letting go puts it there.
func (m Model) carry(msg tea.MouseMsg, pos tea.Mouse) Model {
	l := m.layout()
	switch msg.(type) {
	case tea.MouseMotionMsg:
		if at, ok := nearest(l.bounds, pos.X, pos.Y-tabRows+m.scroll); ok {
			m.dragTo = at
		}
	case tea.MouseReleaseMsg:
		if at, ok := nearest(l.bounds, pos.X, pos.Y-tabRows+m.scroll); ok {
			m.dragTo = at
		}
		m.form.MoveBlock(m.dragFrom, m.dragTo)
		m.dragging = false
	}

	return m
}

// hit does what a click on a part of the page does.
func (m Model) hit(h hit) (Model, tea.Cmd) {
	switch h.kind {
	case hitOption:
		m.form.ChooseOption(h.arg)
	case hitMinus:
		m.form.Change(-1)
	case hitPlus:
		m.form.Change(1)
	case hitBlock:
		m.form.AddBlock(h.arg)
	case hitValue:
		return m.activate()
	case hitCaret:
		m.form.PlaceCaret(h.arg)
	case hitChipRemove:
		m.form.RemoveBlock(h.arg)
	case hitChipHandle:
		m.dragging, m.dragFrom, m.dragTo = true, h.arg, h.arg
	}

	return m, nil
}

// WithConfig replaces the settings being edited.
func (m Model) WithConfig(cfg config.Config) Model {
	m.form = m.form.WithConfig(cfg)

	return m
}

// WithWelcome makes the popup of the first start: its bottom line tells the person what to do
// until they have pressed a key of their own.
func (m Model) WithWelcome() Model {
	m.form = m.form.WithWelcome()

	return m
}
