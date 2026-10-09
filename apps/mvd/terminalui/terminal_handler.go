package terminalui

import (
	"net/http"

	tea "charm.land/bubbletea/v2"

	ttygo "github.com/meta-tui/treactui/packages/tty-go"
)

// NewHandler serves the app as one program that every window of the page shares,
// the model asks for the mouse itself. Only a page served from the same host may connect, plus the
// extra origins given (the page's dev server): the endpoint runs the app, so a
// foreign web page must not be able to drive it.
func NewHandler(newModel func() tea.Model, extraOrigins []string) http.Handler {
	return ttygo.SharedHandler(newModel, ttygo.Options{AllowedOrigins: extraOrigins})
}
