package macbundle

import (
	"strings"
	"testing"
)

func TestTheBundleDescribesANormalApplicationNamedMVD(t *testing.T) {
	plist := InfoPlist("0.0.7", false)

	for _, want := range []string{
		"<key>CFBundleExecutable</key>\n\t<string>mvd</string>",
		"<key>CFBundlePackageType</key>\n\t<string>APPL</string>",
		"<key>CFBundleIdentifier</key>\n\t<string>" + Identifier + "</string>",
		"<key>CFBundleShortVersionString</key>\n\t<string>0.0.7</string>",
	} {
		if !strings.Contains(plist, want) {
			t.Errorf("the Info.plist lacks %q", want)
		}
	}
	if strings.Contains(plist, "LSUIElement") {
		t.Error("the app has a window and a Dock icon, so it is not a menu-bar-only app")
	}
	if !strings.HasPrefix(plist, "<?xml") || !strings.HasSuffix(plist, "</plist>\n") {
		t.Error("the Info.plist is not a complete XML document")
	}
}

func TestTheBundleSaysItOpensMVDFilesAndGivesThemItsIcon(t *testing.T) {
	plist := InfoPlist("0.0.7", true)

	for _, want := range []string{
		"<key>CFBundleDocumentTypes</key>",
		"<string>" + Identifier + ".list</string>",
		"<key>UTExportedTypeDeclarations</key>",
		"<key>public.filename-extension</key>\n\t\t\t\t<array>\n\t\t\t\t\t<string>mvd</string>",
		"<key>CFBundleTypeIconFile</key>\n\t\t\t<string>" + IconName + "</string>",
		"<key>UTTypeIconFile</key>\n\t\t\t<string>" + IconName + "</string>",
	} {
		if !strings.Contains(plist, want) {
			t.Errorf("the Info.plist lacks %q", want)
		}
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
