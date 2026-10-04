package macbundle

import (
	"strings"
	"testing"
)

func TestTheBundleDescribesAMenuBarOnlyApplicationNamedMVD(t *testing.T) {
	plist := InfoPlist("0.0.7", false)

	for _, want := range []string{
		"<key>CFBundleExecutable</key>\n\t<string>mvd</string>",
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
	plist := InfoPlist(`1.0 </string><key>Evil</key><string>x`, false)

	if strings.Contains(plist, "<key>Evil</key>") {
		t.Error("a version string injected markup into the Info.plist")
	}
}

func TestTheIconIsNamedOnlyWhenThereIsOne(t *testing.T) {
	if strings.Contains(InfoPlist("1", false), "CFBundleIconFile") {
		t.Error("the Info.plist names an icon the bundle does not have")
	}
	if want := "<key>CFBundleIconFile</key>\n\t<string>" + IconName + "</string>"; !strings.Contains(InfoPlist("1", true), want) {
		t.Errorf("the Info.plist lacks %q", want)
	}
}
