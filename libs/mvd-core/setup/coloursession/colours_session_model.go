package coloursession

import "youtube-downloader/libs/mvd-core/config"

// Labels are what the colours are called, in the order the screen lists them.
var Labels = []string{
	"Accent (title)", "Focus (active border)", "Highlight (keys, bars)", "Success",
	"Error", "Dim (borders, hints)", "Text", "Selected row (background)",
}

// InvalidColourNotice is what the screen says when the colour typed is not one.
const InvalidColourNotice = "Use a colour like #ff007f or #f07."

// Session edits the interface colours. A colour applies as soon as it is accepted.
type Session struct {
	// Cfg is the settings as they stand, including changes not yet saved.
	Cfg config.Config
	// Cursor is the colour the person is on.
	Cursor int
	// Editing is whether the colour of the cursor is being typed; Text is what was typed.
	Editing bool
	Text    string
	// Status is what the screen says of the last try, "" for nothing.
	Status string
}

// New is the session for the settings cfg.
func New(cfg config.Config) Session { return Session{Cfg: cfg} }

// slot points at the colour setting of item i.
func (s *Session) slot(i int) *string {
	c := &s.Cfg.Colors

	return []*string{&c.Accent, &c.Focus, &c.Highlight, &c.Success, &c.Error, &c.Dim, &c.Text, &c.Selected}[i]
}

// Value is the colour of item i as it stands.
func (s Session) Value(i int) string { return *s.slot(i) }

func defaultFor(i int) string {
	d := config.DefaultThemeColors()

	return []string{d.Accent, d.Focus, d.Highlight, d.Success, d.Error, d.Dim, d.Text, d.Selected}[i]
}
