package windowsnapshot

import "strings"

// ShortcutOf reads a shortcut as the terminal look writes it ("Ctrl+Shift+S", "Del", "F1 / Ctrl+K")
// into keys; of several choices it takes the first. It is nil for no shortcut.
func ShortcutOf(text string) *Shortcut {
	text, _, _ = strings.Cut(text, " / ")
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	parts := strings.Split(text, "+")
	// "Ctrl++" ends in two empty parts: the key is the plus itself.
	if strings.HasSuffix(text, "++") {
		parts = append(parts[:len(parts)-2], "+")
	}
	s := Shortcut{Key: parts[len(parts)-1]}
	for _, part := range parts[:len(parts)-1] {
		switch strings.ToLower(part) {
		case "ctrl", "cmd":
			s.Ctrl = true
		case "shift":
			s.Shift = true
		case "alt", "option":
			s.Alt = true
		}
	}

	return &s
}
