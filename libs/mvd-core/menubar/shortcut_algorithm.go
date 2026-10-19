package menubar

import "strings"

// keyAliases are the names a shortcut text may use for a key whose Bubble Tea
// name is another.
var keyAliases = map[string]string{
	"del":    "delete",
	"return": "enter",
	"escape": "esc",
	"ins":    "insert",
	"pgdn":   "pgdown",
}

// shortcutKeys are the key names a shortcut text stands for, in Bubble Tea's
// spelling: "Ctrl+Shift+S" is "ctrl+shift+s", and "Ctrl+Y / Ctrl+Shift+Z" is
// two shortcuts.
func shortcutKeys(text string) []string {
	var keys []string
	for _, alt := range strings.Split(text, "/") {
		alt = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(alt), " ", ""))
		if alt == "" {
			continue
		}
		parts := strings.Split(alt, "+")
		for i, p := range parts {
			if a, ok := keyAliases[p]; ok {
				parts[i] = a
			}
		}
		keys = append(keys, strings.Join(parts, "+"))
	}
	return keys
}
