package macbundle

import (
	"strings"
	"testing"
)

func TestTheBundleDescribesAMenuBarOnlyApplicationNamedMVD(t *testing.T) {
	plist := InfoPlist("0.0.7")

	for _, want := range []string{
		"<key>CFBundleExecutable</key>\n\t<string>mvd-tray</string>",
		"<key>CFBundlePackageType</key>\n\t<string>APPL</string>",
		"<key>CFBundleIdentifier</key>\n\t<string>" + Identifier + "</string>",
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
	plist := InfoPlist(`1.0 </string><key>Evil</key><string>x`)

	if strings.Contains(plist, "<key>Evil</key>") {
		t.Error("a version string injected markup into the Info.plist")
	}
}
