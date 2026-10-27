// Package windowsnapshot is renderer-neutral: it says what MVD's main window shows, as a classic
// themed window draws it, and what such a window sends back. A snapshot describes roles, not
// places: the menus with their entries, shortcuts and keyboard letters, the buttons with which is
// the default one, the sources and the songs with their tones, the details, the status, and the
// window for adding music. A theme decides where each goes and how it looks.
//
// The shapes and their JSON names are the ones of apps/mvd-web's workspace-snapshot contract, which
// reads them; the two change together.
package windowsnapshot
