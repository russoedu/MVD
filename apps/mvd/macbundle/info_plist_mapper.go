package macbundle

import "html"

// Identifier names the app to macOS.
const Identifier = "io.github.russoedu.mvd"

// InfoPlist is the Info.plist of the MVD.app bundle: the one the app builds around
// itself, and the one the release build puts in the .dmg. withIcon names the icon file
// that Write places in Resources. The app is a normal one, with a Dock icon and a window.
func InfoPlist(version string, withIcon bool) string {
	icon, docIcon, typeIcon := "", "", ""
	if withIcon {
		icon = "\t<key>CFBundleIconFile</key>\n\t<string>" + IconName + "</string>\n"
		// The lists wear the icon of the app.
		docIcon = "\t\t\t<key>CFBundleTypeIconFile</key>\n\t\t\t<string>" + IconName + "</string>\n"
		typeIcon = "\t\t\t<key>UTTypeIconFile</key>\n\t\t\t<string>" + IconName + "</string>\n"
	}

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
	<string>` + ExecutableName + `</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleVersion</key>
	<string>` + html.EscapeString(version) + `</string>
	<key>CFBundleShortVersionString</key>
	<string>` + html.EscapeString(version) + `</string>
` + icon + `	<key>CFBundleDocumentTypes</key>
	<array>
		<dict>
			<key>CFBundleTypeName</key>
			<string>MVD download list</string>
			<key>CFBundleTypeRole</key>
			<string>Editor</string>
			<key>LSHandlerRank</key>
			<string>Owner</string>
			<key>LSItemContentTypes</key>
			<array>
				<string>` + Identifier + `.list</string>
			</array>
` + docIcon + `		</dict>
	</array>
	<key>UTExportedTypeDeclarations</key>
	<array>
		<dict>
			<key>UTTypeIdentifier</key>
			<string>` + Identifier + `.list</string>
			<key>UTTypeDescription</key>
			<string>MVD download list</string>
			<key>UTTypeConformsTo</key>
			<array>
				<string>public.data</string>
			</array>
` + typeIcon + `			<key>UTTypeTagSpecification</key>
			<dict>
				<key>public.filename-extension</key>
				<array>
					<string>mvd</string>
				</array>
			</dict>
		</dict>
	</array>
	<key>LSMinimumSystemVersion</key>
	<string>11.0</string>
	<key>NSHighResolutionCapable</key>
	<true/>
</dict>
</plist>
`
}
