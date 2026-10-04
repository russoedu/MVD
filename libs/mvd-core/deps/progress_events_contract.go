package deps

import (
	"fmt"
	"runtime"
	"strings"
)

// EventKind says what an Event reports.
type EventKind int

const (
	// EventMissing: some tools are not installed. Names lists them and Dir is where
	// they are about to be put.
	EventMissing EventKind = iota
	// EventDownloading: Name is being fetched from URL.
	EventDownloading
	// EventInstalled: Name is ready.
	EventInstalled
	// EventFailed: Name could not be installed; Err says why. Installing carries on
	// with the next tool.
	EventFailed
)

// Event is one step of making the tools available.
type Event struct {
	Kind  EventKind
	Name  string
	Names []string
	URL   string
	Dir   string
	Err   error
}

// Reporter receives the steps as they happen, on the goroutine doing the work.
type Reporter func(Event)

// PrintProgress writes each step to the terminal.
func PrintProgress(e Event) {
	switch e.Kind {
	case EventMissing:
		fmt.Printf("\n[!] Missing dependency/dependencies detected: %s\n", strings.Join(e.Names, ", "))
		fmt.Printf("[+] Automatically downloading dependencies for [%s/%s] into %s...\n\n", runtime.GOOS, runtime.GOARCH, e.Dir)
	case EventDownloading:
		fmt.Printf("    Downloading %s from %s...\n", e.Name, e.URL)
	case EventInstalled:
		fmt.Printf("    [OK] Successfully installed %s!\n\n", e.Name)
	case EventFailed:
		fmt.Printf("    [!] Could not install %s: %v\n", e.Name, e.Err)
	}
}
