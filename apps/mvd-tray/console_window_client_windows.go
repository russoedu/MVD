package main

import (
	"os"

	"golang.org/x/sys/windows"
)

// attachParentProcess is ATTACH_PARENT_PROCESS: attach to the console of whoever started us.
const attachParentProcess = ^uint32(0)

var attachConsoleProc = windows.NewLazySystemDLL("kernel32.dll").NewProc("AttachConsole")

// attachParentConsole lets a windowed program print to the terminal that started it.
//
// The release build is linked as a windowed program (-H=windowsgui), which has no
// console and so no standard output. Started from a terminal it should still talk to
// that terminal, so it attaches to it and points any standard stream that is not
// already redirected at the console. Started from Explorer there is no terminal to
// attach to, and nothing changes. Nothing depends on this working: start-up errors
// also go to showFatal.
func attachParentConsole() {
	if attached, _, _ := attachConsoleProc.Call(uintptr(attachParentProcess)); attached == 0 {
		return
	}

	console, err := os.OpenFile("CONOUT$", os.O_RDWR, 0)
	if err != nil {
		return
	}
	if !streamIsOpen(windows.STD_OUTPUT_HANDLE) {
		os.Stdout = console
	}
	if !streamIsOpen(windows.STD_ERROR_HANDLE) {
		os.Stderr = console
	}
}

// streamIsOpen reports whether a standard stream was handed to this process, for
// example by a redirect, as opposed to being absent.
func streamIsOpen(which uint32) bool {
	handle, err := windows.GetStdHandle(which)

	return err == nil && handle != 0 && handle != windows.InvalidHandle
}

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

// trayUnavailable is what to do when the tray icon could not be created: a windowed
// program then has neither an icon nor a console, so the page is the only way left to
// reach it, and it is opened whatever -no-browser said.
func trayUnavailable(url string) {
	_ = openBrowser(url)
}
