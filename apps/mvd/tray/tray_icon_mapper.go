package tray

import (
	_ "embed"
)

// The icons are drawn for each platform, from assets/app-mvd.ico,
// assets/app-mvd-mac-2x.png and assets/app-mvd-linux.png. Replace the copies here
// when the logo changes.
var (
	// windowsICO holds several sizes, so Windows picks the one for the tray's
	// size at the screen's scaling instead of shrinking one image.
	//
	//go:embed tray_icon_windows.ico
	windowsICO []byte

	// macPNG is 44 px tall, twice the height of the menu bar icon, so it stays
	// sharp on a Retina screen.
	//
	//go:embed tray_icon_mac.png
	macPNG []byte

	// linuxPNG is sent over D-Bus to whatever shows the tray.
	//
	//go:embed tray_icon_linux.png
	linuxPNG []byte
)

// trayIcon returns the icon in the format the tray of goos expects: an .ico on
// Windows, a PNG everywhere else.
func trayIcon(goos string) []byte {
	switch goos {
	case "windows":
		return windowsICO
	case "darwin":
		return macPNG
	default:
		return linuxPNG
	}
}
