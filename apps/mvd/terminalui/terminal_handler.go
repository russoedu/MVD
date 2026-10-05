package terminalui

import (
	"net/http"

	tea "github.com/charmbracelet/bubbletea"

	ttygo "github.com/TReactUI/TReactUI/packages/tty-go"
)

// NewHandler serves the app as one program that every window of the page shares,
// with the mouse on. Only a page served from the same host may connect, plus the
// extra origins given (the page's dev server): the endpoint runs the app, so a
// foreign web page must not be able to drive it.
func NewHandler(newModel func() tea.Model, extraOrigins []string) http.Handler {
	return ttygo.SharedHandler(newModel, ttygo.Options{AllowedOrigins: extraOrigins, Mouse: true})
}
