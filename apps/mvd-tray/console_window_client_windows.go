package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32               = windows.NewLazySystemDLL("kernel32.dll")
	getConsoleProcessList  = kernel32.NewProc("GetConsoleProcessList")
	freeConsoleProc        = kernel32.NewProc("FreeConsole")
	processesSharingBuffer [2]uint32
)

// hideOwnConsole closes the console window Windows opened for this program, and
// reports whether it did.
//
// A Go program is a console program unless it is linked as a windowed one, so
// launching the tray app from Explorer opens a console window next to the tray icon.
// That window is only worth keeping when someone started the app from a terminal and
// is reading its output there. So it is released only if this process is the only one
// attached to its console: when a shell shares it, the shell's window stays untouched.
func hideOwnConsole() bool {
	attached, _, _ := getConsoleProcessList.Call(
		uintptr(unsafe.Pointer(&processesSharingBuffer[0])),
		uintptr(len(processesSharingBuffer)),
	)
	if attached != 1 {
		return false
	}
	released, _, _ := freeConsoleProc.Call()

	return released != 0
}
