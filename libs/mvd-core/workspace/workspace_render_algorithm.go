package workspace

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/menubar"
)

// render draws the menu bar on the top row and, under it, the review or the download
// screen. An open menu is laid over them, and the question or the notice takes the bottom
// row.
// Render is the screen as text, for a host that lays something over it.
func (m Model) Render() string { return m.render() }

func (m Model) render() string {
	if m.width == 0 || m.height == 0 {
		return "starting..."
	}
	if m.addOpen {
		return m.addView()
	}
	below := m.view().Render()
	lines := strings.Split(below, "\n")
	if len(lines) > m.height-topRows {
		lines = lines[:m.height-topRows]
	}
	if bottom := m.bottomLine(); bottom != "" && len(lines) > 0 {
		lines[len(lines)-1] = bottom
	}
	screen := m.buttonsView(m.barRow() + "\n" + strings.Join(lines, "\n"))
	if m.inline.on {
		screen = m.inlineView(screen)
	}
	if m.closing.On {
		x, y := m.closeOrigin()
		screen = dialog.Overlay(dialog.Dim(screen), m.closeView(), x, y)
	}
	if m.about {
		x, y := m.aboutOrigin()
		screen = dialog.Overlay(dialog.Dim(screen), m.aboutView(), x, y)
	}
	if m.shortcuts {
		x, y := m.shortcutsOrigin()
		screen = dialog.Overlay(dialog.Dim(screen), m.shortcutsView(), x, y)
	}
	if m.bar.Open() >= 0 {
		screen = overlayMenu(screen, m.bar.View())
	}
	return screen
}

// barRow is the top row: the bar alone.
func (m Model) barRow() string {
	view := m.bar
	view.Close()
	view.SetSize(m.width, 1)
	return view.View()
}

// bottomLine is the question being asked, else what the last action said.
func (m Model) bottomLine() string {
	switch {
	case m.asking != askNone:
		return cells.Fit(" "+look.Key.Render(m.input.View()), m.width)
	case m.life.Starting:
		return cells.Fit(" "+look.Dim.Render("Getting ready..."), m.width)
	case m.life.Stopping:
		return cells.Fit(" "+look.Dim.Render("Stopping the downloads..."), m.width)
	case m.notice != "":
		return cells.Fit(" "+look.Red.Render(m.notice), m.width)
	case m.caution != "":
		return cells.Fit(" "+look.Yellow.Render(m.caution), m.width)
	case m.done != "":
		return cells.Fit(" "+look.Green.Render("✓ "+m.done), m.width)
	case m.toolsMessage() != "":
		return cells.Fit(" "+m.toolsLine(), m.width)
	case m.inline.on:
		return cells.Fit(" "+look.Dim.Render("Paste or type a link, then press Enter. Esc cancels."), m.width)
	case m.disabledReason() != "":
		return cells.Fit(" "+look.Dim.Render(m.disabledReason()), m.width)
	case m.finished():
		return cells.Fit(" "+m.summaryLine(), m.width)
	case m.editable():
		return cells.Fit(" "+look.Dim.Render(m.idleHints()), m.width)
	}
	return ""
}

// overlayMenu lays the open menu over the screen: the bar on the top row, and the box of
// the menu alone under it, so the screen shows at both sides of the box and below it.
func overlayMenu(screen, barView string) string {
	lines := strings.Split(screen, "\n")
	rows := strings.Split(barView, "\n")
	lines[0] = rows[0]
	screen = strings.Join(lines, "\n")
	for y := 1; y < len(rows); y++ {
		plain := ansi.Strip(rows[y])
		left := len(plain) - len(strings.TrimLeft(plain, " "))
		right := ansi.StringWidth(strings.TrimRight(plain, " "))
		if right <= left {
			continue
		}
		screen = dialog.Overlay(screen, ansi.Cut(rows[y], left, right), left, y)
	}
	return screen
}

// idleHints is the line of keys at the bottom while nothing runs: the few that start the work,
// and the one that lists them all.
func (m Model) idleHints() string {
	if _, _, _, drawn := m.buttonRows(); !drawn && m.width > 0 {
		// The window is too narrow for the buttons: the Run menu has what they did.
		return "No room for buttons at this width: use the Run menu  ·  " + menubar.ShortcutLabel("F1") + " shortcuts"
	}
	return strings.Join([]string{
		menubar.ShortcutLabel("F1") + " shortcuts",
		menubar.ShortcutLabel("Ctrl+L") + " add source",
		menubar.ShortcutLabel("Ctrl+R") + " load playlists",
		menubar.ShortcutLabel("Ctrl+O") + " open list",
		menubar.ShortcutLabel("Ctrl+,") + " preferences",
	}, "  ·  ")
}

// toolsMessage is what to say about the tools while they are not ready, or "".
func (m Model) toolsMessage() string {
	state, waiting := m.toolsWaiting()
	if !waiting {
		return ""
	}

	return state.Message
}

// toolsLine is that message dressed for the bottom line: calm while it is a wait, red when it is
// a failure.
func (m Model) toolsLine() string {
	if state, _ := m.toolsWaiting(); state.Failed {
		return look.Red.Render(state.Message)
	}

	return look.Yellow.Render(m.toolsMessage())
}

// summaryLine is how a finished run went and where its videos are, in a line the width of the
// screen: the folder is shortened in the middle, so its end, which names it, stays.
func (m Model) summaryLine() string {
	t := m.dl.State.Tally()
	head := fmt.Sprintf("✓ %d downloaded", t.Done)
	plain := head
	line := look.Green.Render(head)
	if t.Failed > 0 {
		failed := fmt.Sprintf("✗ %d failed", t.Failed)
		plain += "  " + failed
		line += "  " + look.Red.Render(failed)
	}
	if unread := m.unreadLinks(); unread > 0 {
		said := fmt.Sprintf("✗ %d %s could not be read", unread, plural("source", unread))
		plain += "  " + said
		line += "  " + look.Red.Render(said)
	}
	const label = "  ·  Videos in "
	hint := "  ·  " + menubar.ShortcutLabel("Ctrl+Shift+O") + " opens it"
	room := m.width - ansi.StringWidth(plain) - ansi.StringWidth(label) - 2
	if room < 30 {
		return line // too narrow for the folder: the counts say it
	}
	if room-ansi.StringWidth(hint) >= ansi.StringWidth(m.cfg.OutputDir) || room-ansi.StringWidth(hint) >= 24 {
		room -= ansi.StringWidth(hint)
	} else {
		hint = ""
	}

	return line + look.Dim.Render(label) + look.Text.Render(cells.TruncateMiddle(m.cfg.OutputDir, room)) + look.Dim.Render(hint)
}

// unreadLinks is how many links of the run could not be read at all, a different thing from a song
// that failed: nothing of them was listed.
func (m Model) unreadLinks() int {
	n := 0
	for _, pl := range m.dl.State.Playlists {
		if pl.Err != "" {
			n++
		}
	}

	return n
}

// plural is the noun, with an s when there is not exactly one.
func plural(noun string, n int) string {
	if n == 1 {
		return noun
	}

	return noun + "s"
}
