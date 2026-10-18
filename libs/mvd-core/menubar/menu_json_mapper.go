package menubar

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseMenus reads menus from the text a design file carries them as: a JSON
// list of {"name", "entries": [{"name", "shortcut"} or {"separator": true}]}.
// Empty text is no menus. An entry that is not a separator needs a name.
func ParseMenus(text string) ([]Menu, error) {
	if strings.TrimSpace(text) == "" {
		return nil, nil
	}
	var menus []Menu
	if err := json.Unmarshal([]byte(text), &menus); err != nil {
		return nil, fmt.Errorf("menubar: menus are not a list of menus: %w", err)
	}
	for i, m := range menus {
		if strings.TrimSpace(m.Name) == "" {
			return nil, fmt.Errorf("menubar: menu %d has no name", i)
		}
		for j, e := range m.Entries {
			if !e.Separator && strings.TrimSpace(e.Name) == "" {
				return nil, fmt.Errorf("menubar: entry %d of menu %q has no name", j, m.Name)
			}
		}
	}
	return menus, nil
}

// MenusJSON writes menus as ParseMenus reads them.
func MenusJSON(menus []Menu) string {
	if len(menus) == 0 {
		return "[]"
	}
	for i := range menus {
		if menus[i].Entries == nil {
			menus[i].Entries = []Entry{}
		}
	}
	data, err := json.Marshal(menus)
	if err != nil {
		return "[]"
	}
	return string(data)
}
