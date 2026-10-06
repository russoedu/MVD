package folderdialog

import "errors"

// ErrNoDialog is what Pick returns on a machine that cannot show a chooser (no desktop,
// or none of the tools it relies on).
var ErrNoDialog = errors.New("this machine has no folder chooser")
