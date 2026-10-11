package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
)

// RunApp runs the whole interactive app (the setup screens, the playlist editor and the
// download screen, one after the other) until the person quits the setup screens or ctx
// ends. It is what NewAppModel is for hosts that run the program themselves.
func RunApp(ctx context.Context, in AppInput) error {
	_, err := tea.NewProgram(NewAppModel(in), tea.WithContext(ctx)).Run()
	return err
}
