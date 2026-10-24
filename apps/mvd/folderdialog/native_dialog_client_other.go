//go:build !windows

package folderdialog

import "context"

// hasNativeDialog is false here: other systems show their file windows through a program of
// their own (osascript, zenity, kdialog).
const hasNativeDialog = false

// nativeDialog is never used where hasNativeDialog is false.
func nativeDialog(context.Context, dialogRequest) (string, bool, error) {
	return "", false, ErrNoDialog
}
