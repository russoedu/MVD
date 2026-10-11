package tui

import (
	"strings"
	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/look"

	"github.com/charmbracelet/x/ansi"
)

// renderBox draws a rounded box of exactly w x h cells with a title in the
// top border and an optional right aligned label.
func renderBox(title, right string, w, h int, lines []string, focused bool) string {
	bs := look.Border
	if focused {
		bs = look.BorderOn
	}
	innerW := w - 2
	innerH := h - 2
	if innerW < 1 {
		innerW = 1
	}
	if innerH < 0 {
		innerH = 0
	}

	// top border: ╭─ Title ───── right ─╮
	tw := innerW - 4
	if right != "" {
		tw -= ansi.StringWidth(right) + 3
	}
	t := cells.Truncate(title, max(tw, 1))
	fill := innerW - 3 - ansi.StringWidth(t)
	if right != "" {
		fill -= ansi.StringWidth(right) + 3
	}
	if fill < 0 {
		fill = 0
	}
	top := bs.Render("╭─ ") + look.Title.Render(t) + bs.Render(" "+strings.Repeat("─", fill))
	if right != "" {
		top += bs.Render(" ") + look.Dim.Render(right) + bs.Render(" ─")
	}
	top += bs.Render("╮")

	var b strings.Builder
	b.WriteString(top)
	for i := 0; i < innerH; i++ {
		b.WriteString("\n")
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		b.WriteString(bs.Render("│") + cells.PadRight(cells.Truncate(line, innerW), innerW) + bs.Render("│"))
	}
	b.WriteString("\n" + bs.Render("╰"+strings.Repeat("─", innerW)+"╯"))
	return b.String()
}
