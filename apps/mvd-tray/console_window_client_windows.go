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

// What a message box returns for the button that was pressed.
const (
	idCancel = 2
	idYes    = 6
	idNo     = 7
)

// askChoice puts a question to the person in a message box and returns their choice.
// With two choices it is Yes and No; with three it is Yes, No and Cancel, and the box
// says which button means what. Closing the box counts as the last choice, leaving
// things as they are.
func askChoice(title, question string, choices []string) answer {
	flags := uint32(windows.MB_YESNO)
	text := question
	if len(choices) == 3 {
		flags = windows.MB_YESNOCANCEL
		text += "\n\nYes: " + choices[0] + "      No: " + choices[1] + "      Cancel: " + choices[2]
	}
	body, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return answerLeave
	}
	caption, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return answerLeave
	}

	pressed, _ := windows.MessageBox(0, body, caption, flags|windows.MB_ICONQUESTION|windows.MB_SETFOREGROUND)
	switch pressed {
	case idYes:
		return answerFirst
	case idNo:
		if len(choices) == 3 {
			return answerSecond
		}
	}

	return answerLeave
}

// trayUnavailable is what to do when the tray icon could not be created: a windowed
// program then has neither an icon nor a console, so the page is the only way left to
// reach it, and it is opened whatever -no-browser said.
func trayUnavailable(url string) {
	_ = openBrowser(url)
}
