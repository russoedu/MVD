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
	// EventUpToDate: Name was checked and is the newest build there is, or a newer one
	// is not new enough yet to be worth fetching. Detail says which.
	EventUpToDate
	// EventUpdated: Name was replaced by a newer build. Detail is its version.
	EventUpdated
	// EventUpdateFailed: Name could not be checked or updated (no connection, GitHub
	// limiting requests, a bad download). The copy that was there is untouched.
	EventUpdateFailed
)

// Event is one step of making the tools available or keeping them current.
type Event struct {
	Kind   EventKind
	Name   string
	Names  []string
	URL    string
	Dir    string
	Detail string
	Err    error
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
	case EventUpToDate:
		fmt.Printf("    %s is up to date (%s)\n", e.Name, e.Detail)
	case EventUpdated:
		fmt.Printf("    [OK] Updated %s to %s\n", e.Name, e.Detail)
	case EventUpdateFailed:
		fmt.Printf("    [!] Could not update %s: %v\n", e.Name, e.Err)
	}
}
