package main

import (
	"fmt"
	"strings"
	"youtube-downloader/apps/mvd-tray/notification"
	"youtube-downloader/libs/mvd-core/deps"
)

// depsReporter turns the installer's steps into log lines and, for the two moments
// the person needs to know about while the app is starting, a notification: when
// downloading starts (the app will take a while to appear, and why) and when a tool
// could not be installed. Checking for updates only logs; the person is told once, by
// toolUpdates, if something was actually replaced.
func depsReporter(logf func(string, ...interface{}), notify func(notification.Notice)) deps.Reporter {
	return func(e deps.Event) {
		switch e.Kind {
		case deps.EventMissing:
			logf("Missing: %s. Downloading into %s", strings.Join(e.Names, ", "), e.Dir)
			notify(notification.Notice{
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
			notify(notification.Notice{
				Title:   "MVD could not set up " + e.Name,
				Text:    fmt.Sprintf("%v. Check your internet connection and start MVD again.", e.Err),
				Failure: true,
			})
		case deps.EventUpToDate:
			logf("%s is up to date (%s)", e.Name, e.Detail)
		case deps.EventUpdated:
			logf("Updated %s to %s", e.Name, e.Detail)
		case deps.EventUpdateFailed:
			logf("Could not check or update %s: %v", e.Name, e.Err)
		}
	}
}
