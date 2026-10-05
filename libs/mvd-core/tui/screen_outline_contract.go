package tui

// ScreenOutline describes the screen a setup model currently shows, in plain
// terms and without any styling, for hosts that need more than the drawn
// text: a screen reader, for instance.
type ScreenOutline struct {
	Title string
	// Text is the content of the multi-line editor (the download list), or the
	// current folder on the folder picker; HasText says whether the screen has one.
	Text    string
	HasText bool
	// ItemsLabel names what Items are: "Settings", "Sub-folders", "Downloads".
	ItemsLabel string
	Items      []OutlineItem
	Keys       []OutlineKey
	// Prompt is a question waiting for an answer, such as the quit confirmation.
	Prompt string
}

// OutlineItem is one selectable row: a setting, or a folder.
type OutlineItem struct {
	Label    string
	Value    string
	Selected bool
	// Editing is true while this item's editor is open; Value is then the
	// editor's current content, and Choices lists the options of a radio editor.
	Editing bool
	Choices []string
	Chosen  int
}

// OutlineKey is one entry of the key bar.
type OutlineKey struct {
	Key         string
	Description string
}
