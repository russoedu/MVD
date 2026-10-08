// Package nativewindow shows the app's page in a window of its own, drawn by
// the operating system's web view (WebView2 on Windows), so no browser has to
// be installed. It is the alternative to opening a Chromium app window.
package nativewindow

import "errors"

// ErrUnavailable means the system has no web view this program can use here: the
// caller should show the page in a browser instead.
var ErrUnavailable = errors.New("the system web view is not available")
