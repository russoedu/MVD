package console

import "golang.org/x/sys/windows"

// ShowInfo tells the person something, in a message box, because a windowed program may
// have nowhere else to print. It returns when the box is closed.
func ShowInfo(message string) {
	text, err := windows.UTF16PtrFromString(message)
	if err != nil {
		return
	}
	title, err := windows.UTF16PtrFromString("MVD")
	if err != nil {
		return
	}
	_, _ = windows.MessageBox(0, text, title, windows.MB_OK|windows.MB_ICONINFORMATION|windows.MB_SETFOREGROUND)
}
