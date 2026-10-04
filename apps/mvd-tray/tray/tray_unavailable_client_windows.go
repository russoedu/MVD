package tray

import "youtube-downloader/apps/mvd-tray/browser"

// Unavailable is what to do when the tray icon could not be created: a windowed
// program then has neither an icon nor a console, so the page is the only way left to
// reach it, and it is opened whatever -no-browser said.
func Unavailable(url string) {
	_ = browser.Open(url)
}
