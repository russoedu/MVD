package tui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// renderBox draws a rounded box of exactly w x h cells with a title in the
// top border and an optional right aligned label.
func renderBox(title, right string, w, h int, lines []string, focused bool) string {
	bs := styBorder
	if focused {
		bs = styBorderOn
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
	t := truncate(title, max(tw, 1))
	fill := innerW - 3 - ansi.StringWidth(t)
	if right != "" {
		fill -= ansi.StringWidth(right) + 3
	}
	if fill < 0 {
		fill = 0
	}
	top := bs.Render("╭─ ") + styTitle.Render(t) + bs.Render(" "+strings.Repeat("─", fill))
	if right != "" {
		top += bs.Render(" ") + styDim.Render(right) + bs.Render(" ─")
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
		b.WriteString(bs.Render("│") + padRight(truncate(line, innerW), innerW) + bs.Render("│"))
	}
	b.WriteString("\n" + bs.Render("╰"+strings.Repeat("─", innerW)+"╯"))
	return b.String()
}

// window returns the slice of rows that keeps index sel visible in h rows.
func window(rows []string, sel, h int) []string {
	if h <= 0 || len(rows) <= h {
		return rows
	}
	start := sel - h/2
	if start < 0 {
		start = 0
	}
	if start+h > len(rows) {
		start = len(rows) - h
	}
	return rows[start : start+h]
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if ansi.StringWidth(s) <= w {
		return s
	}
	if w == 1 {
		return ansi.Truncate(s, 1, "")
	}
	return ansi.Truncate(s, w, "…")
}

// wrap breaks s into lines of at most w cells on word boundaries.
func wrap(s string, w int) []string {
	if w <= 0 {
		return nil
	}
	var lines []string
	line := ""
	for _, word := range strings.Fields(s) {
		for ansi.StringWidth(word) > w {
			if line != "" {
				lines = append(lines, line)
				line = ""
			}
			lines = append(lines, ansi.Truncate(word, w, ""))
			word = word[len(ansi.Truncate(word, w, "")):]
		}
		switch {
		case line == "":
			line = word
		case ansi.StringWidth(line)+1+ansi.StringWidth(word) <= w:
			line += " " + word
		default:
			lines = append(lines, line)
			line = word
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// fit pads or truncates s to exactly w cells.
func fit(s string, w int) string {
	return padRight(truncate(s, w), w)
}

func padRight(s string, w int) string {
	sw := ansi.StringWidth(s)
	if sw >= w {
		return s
	}
	return s + strings.Repeat(" ", w-sw)
}

func clamp(v, lo, hi int) int {
	if hi < lo {
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
