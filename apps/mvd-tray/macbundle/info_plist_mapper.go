package macbundle

import "html"

// Identifier names the app to macOS.
const Identifier = "io.github.russoedu.mvd"

// InfoPlist is the Info.plist of the MVD.app bundle the app builds around itself on
// macOS. LSUIElement keeps it out of the Dock and the application switcher: it lives in
// the menu bar, like the tray icon on the other systems.
func InfoPlist(version string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key>
	<string>MVD</string>
	<key>CFBundleDisplayName</key>
	<string>MVD</string>
	<key>CFBundleIdentifier</key>
	<string>` + Identifier + `</string>
	<key>CFBundleExecutable</key>
	<string>mvd-tray</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleVersion</key>
	<string>` + html.EscapeString(version) + `</string>
	<key>CFBundleShortVersionString</key>
	<string>` + html.EscapeString(version) + `</string>
	<key>LSMinimumSystemVersion</key>
	<string>11.0</string>
	<key>LSUIElement</key>
	<true/>
	<key>NSHighResolutionCapable</key>
	<true/>
</dict>
</plist>
`
}
