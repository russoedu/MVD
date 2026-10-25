package menumodel

// Entry is one line of a menu: a name with an optional shortcut, or a
// separator rule. The shortcut is shown at the right of the name, and pressing
// it chooses the entry.
type Entry struct {
	Name      string `json:"name,omitempty"`
	Shortcut  string `json:"shortcut,omitempty"`
	Separator bool   `json:"separator,omitempty"`
}

// Menu is one heading of the bar and the entries that drop down under it.
type Menu struct {
	Name    string  `json:"name"`
	Entries []Entry `json:"entries"`
}

// Selection is an entry that was chosen, with the mouse or the keyboard, or by pressing its
// shortcut.
type Selection struct {
	// Menu and Entry are the position of the entry: the menu in the bar and the
	// entry in that menu, both from 0.
	Menu, Entry int
	// MenuName, Name and Shortcut are what the entry says.
	MenuName, Name, Shortcut string
}
