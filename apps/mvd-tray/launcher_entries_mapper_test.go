package main

import (
	"strings"
	"testing"
)

func TestTheBundleDescribesAMenuBarOnlyApplicationNamedMVD(t *testing.T) {
	plist := infoPlist("0.0.7")

	for _, want := range []string{
		"<key>CFBundleExecutable</key>\n\t<string>mvd-tray</string>",
		"<key>CFBundlePackageType</key>\n\t<string>APPL</string>",
		"<key>CFBundleIdentifier</key>\n\t<string>" + bundleIdentifier + "</string>",
		"<key>CFBundleShortVersionString</key>\n\t<string>0.0.7</string>",
		"<key>LSUIElement</key>\n\t<true/>",
	} {
		if !strings.Contains(plist, want) {
			t.Errorf("the Info.plist lacks %q", want)
		}
	}
	if !strings.HasPrefix(plist, "<?xml") || !strings.HasSuffix(plist, "</plist>\n") {
		t.Error("the Info.plist is not a complete XML document")
	}
}

func TestAVersionIsEscapedSoItCannotBreakTheDocument(t *testing.T) {
	plist := infoPlist(`1.0 </string><key>Evil</key><string>x`)

	if strings.Contains(plist, "<key>Evil</key>") {
		t.Error("a version string injected markup into the Info.plist")
	}
}

func TestTheMenuEntryStartsTheProgramWithoutATerminal(t *testing.T) {
	entry := desktopEntry("/home/me/.local/bin/mvd-tray")

	for _, want := range []string{"[Desktop Entry]", "Type=Application", "Name=MVD", "Exec=/home/me/.local/bin/mvd-tray\n", "Terminal=false"} {
		if !strings.Contains(entry, want) {
			t.Errorf("the entry lacks %q:\n%s", want, entry)
		}
	}
}

func TestAProgramPathWithSpacesOrSpecialCharactersIsQuotedAsOneWord(t *testing.T) {
	cases := map[string]string{
		"/home/my user/bin/mvd-tray": `Exec="/home/my user/bin/mvd-tray"`,
		`/home/me/a"b/mvd-tray`:      `Exec="/home/me/a\"b/mvd-tray"`,
		"/home/me/$HOME/mvd-tray":    `Exec="/home/me/\$HOME/mvd-tray"`,
	}
	for program, want := range cases {
		if got := desktopEntry(program); !strings.Contains(got, want+"\n") {
			t.Errorf("for %q want %s in:\n%s", program, want, got)
		}
	}
}
