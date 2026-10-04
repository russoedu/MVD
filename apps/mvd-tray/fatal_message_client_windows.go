package main

import "golang.org/x/sys/windows"

// showFatal tells the person why the app could not start, in a message box, because
// a windowed program may have nowhere else to print.
func showFatal(message string) {
	text, err := windows.UTF16PtrFromString(message)
	if err != nil {
		return
	}
	title, err := windows.UTF16PtrFromString("MVD")
	if err != nil {
		return
	}
	_, _ = windows.MessageBox(0, text, title, windows.MB_OK|windows.MB_ICONERROR|windows.MB_SETFOREGROUND)
}
