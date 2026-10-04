package main

import "youtube-downloader/apps/mvd-tray/browser"

// trayUnavailable is what to do when the tray icon could not be created: a windowed
// program then has neither an icon nor a console, so the page is the only way left to
// reach it, and it is opened whatever -no-browser said.
func trayUnavailable(url string) {
	_ = browser.Open(url)
}
