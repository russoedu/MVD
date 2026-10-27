package appwindow

import (
	"context"
	"encoding/json"
	"sync"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"youtube-downloader/libs/mvd-core/prefsform"
	"youtube-downloader/libs/mvd-core/workspace/windowsnapshot"
)

// The events of the classic themed window: the style and what the window shows go to the page,
// and what the person does comes back. The page asks for the style when it starts, as it may
// start after the program told it.
const (
	styleEvent        = "mvd:style"
	styleRequestEvent = "mvd:style-request"
	snapshotEvent     = "mvd:snapshot"
	intentEvent       = "mvd:intent"
	prefsEvent        = "mvd:prefs"
	prefsIntentEvent  = "mvd:prefs-intent"
)

// ClassicWindow carries the windows the page draws between the program and the page: the window
// style chosen in the preferences, what the classic window shows, the preferences while they are
// open (in every style), each as JSON sent only when it changed, and the intents the page sends
// back. It is made before the window and works once the window is up; before that, what it is told
// is kept and sent when the page asks.
type ClassicWindow struct {
	// OnIntent is given each intent of the main window the page sends, and OnPrefsIntent each one
	// of the preferences; nil drops them.
	OnIntent      func(windowsnapshot.Intent)
	OnPrefsIntent func(prefsform.WindowIntent)

	mu       sync.Mutex
	ctx      context.Context
	style    string
	snapshot string
	prefs    string
}

// attach starts listening to the page, and returns how to stop.
func (w *ClassicWindow) attach(ctx context.Context) func() {
	w.mu.Lock()
	w.ctx = ctx
	w.mu.Unlock()
	stopIntents := wailsruntime.EventsOn(ctx, intentEvent, func(data ...interface{}) {
		if len(data) == 0 || w.OnIntent == nil {
			return
		}
		text, ok := data[0].(string)
		var in windowsnapshot.Intent
		if ok && json.Unmarshal([]byte(text), &in) == nil {
			w.OnIntent(in)
		}
	})
	stopPrefsIntents := wailsruntime.EventsOn(ctx, prefsIntentEvent, func(data ...interface{}) {
		if len(data) == 0 || w.OnPrefsIntent == nil {
			return
		}
		text, ok := data[0].(string)
		var in prefsform.WindowIntent
		if ok && json.Unmarshal([]byte(text), &in) == nil {
			w.OnPrefsIntent(in)
		}
	})
	stopRequests := wailsruntime.EventsOn(ctx, styleRequestEvent, func(...interface{}) { w.resend() })

	return func() {
		stopIntents()
		stopPrefsIntents()
		stopRequests()
	}
}

// SetStyle says which window style the settings have.
func (w *ClassicWindow) SetStyle(style string) {
	w.mu.Lock()
	w.style = style
	ctx := w.ctx
	w.mu.Unlock()
	if ctx != nil {
		wailsruntime.EventsEmit(ctx, styleEvent, style)
	}
}

// Show says what the window shows; the page is told only when it changed.
func (w *ClassicWindow) Show(snap windowsnapshot.Snapshot) {
	data, err := json.Marshal(snap)
	if err != nil {
		return
	}
	w.mu.Lock()
	changed := string(data) != w.snapshot
	w.snapshot = string(data)
	ctx := w.ctx
	w.mu.Unlock()
	if changed && ctx != nil {
		wailsruntime.EventsEmit(ctx, snapshotEvent, string(data))
	}
}

// ShowPreferences says what the preferences window shows, or nil once it is closed; the page is
// told only when it changed.
func (w *ClassicWindow) ShowPreferences(prefs *prefsform.Window) {
	data, err := json.Marshal(prefs)
	if err != nil {
		return
	}
	w.mu.Lock()
	changed := string(data) != w.prefs
	w.prefs = string(data)
	ctx := w.ctx
	w.mu.Unlock()
	if changed && ctx != nil {
		wailsruntime.EventsEmit(ctx, prefsEvent, string(data))
	}
}

// resend tells the page the style and the last snapshot again, for a page that has just started.
func (w *ClassicWindow) resend() {
	w.mu.Lock()
	ctx, style, snapshot, prefs := w.ctx, w.style, w.snapshot, w.prefs
	w.mu.Unlock()
	if ctx == nil || style == "" {
		return
	}
	wailsruntime.EventsEmit(ctx, styleEvent, style)
	if snapshot != "" {
		wailsruntime.EventsEmit(ctx, snapshotEvent, snapshot)
	}
	if prefs != "" {
		wailsruntime.EventsEmit(ctx, prefsEvent, prefs)
	}
}
