//go:build darwin

package appwindow

// Wails' macOS code uses UTType, which lives in the UniformTypeIdentifiers framework. Its
// own build tool adds the framework to the link; a plain go build, which is how this app
// is built, needs it said here.

/*
#cgo darwin LDFLAGS: -framework UniformTypeIdentifiers
*/
import "C"
