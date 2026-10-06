package uninstall

import "errors"

// ErrDeclined is what asking to remove the app returns when the person said no.
var ErrDeclined = errors.New("the person declined to remove the app")

// ErrNoDialog is what it returns on a machine with no way to ask (no desktop, or none
// of the tools it relies on), so nothing was removed.
var ErrNoDialog = errors.New("this machine cannot ask for confirmation")
