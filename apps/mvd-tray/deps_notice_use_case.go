package main

import (
	"fmt"
	"strings"

	"youtube-downloader/libs/mvd-core/deps"
)

// depsReporter turns the installer's steps into log lines and, for the two moments
// the person needs to know about, a notification: when downloading starts (the app
// will take a while to appear, and why) and when a tool could not be installed.
func depsReporter(logf func(string, ...interface{}), notify func(userNotice)) deps.Reporter {
	return func(e deps.Event) {
		switch e.Kind {
		case deps.EventMissing:
			logf("Missing: %s. Downloading into %s", strings.Join(e.Names, ", "), e.Dir)
			notify(userNotice{
				Title: "MVD is setting itself up",
				Text: fmt.Sprintf("Downloading %s into %s. This happens once, can take a minute, and MVD opens when it is done.",
					strings.Join(e.Names, ", "), e.Dir),
			})
		case deps.EventDownloading:
			logf("Downloading %s from %s", e.Name, e.URL)
		case deps.EventInstalled:
			logf("Installed %s", e.Name)
		case deps.EventFailed:
			logf("Could not install %s: %v", e.Name, e.Err)
			notify(userNotice{
				Title:   "MVD could not set up " + e.Name,
				Text:    fmt.Sprintf("%v. Check your internet connection and start MVD again.", e.Err),
				Failure: true,
			})
		}
	}
}
