package filetype

import (
	"strings"
	"testing"
)

func TestWindowsIsToldTheKindOfFileItsIconAndHowToOpenIt(t *testing.T) {
	byKey := map[string]string{}
	for _, v := range RegistryValues(`C:\Apps\MVD\mvd.exe`) {
		if v.Name != "" {
			t.Errorf("%s: only default values are written", v.Key)
		}
		byKey[v.Key] = v.Value
	}
	want := map[string]string{
		`Software\Classes\.mvd`:                                "MVD.DownloadList",
		`Software\Classes\MVD.DownloadList`:                    "MVD download list",
		`Software\Classes\MVD.DownloadList\DefaultIcon`:        `"C:\Apps\MVD\mvd.exe",0`,
		`Software\Classes\MVD.DownloadList\shell\open\command`: `"C:\Apps\MVD\mvd.exe" "%1"`,
	}
	for key, value := range want {
		if byKey[key] != value {
			t.Errorf("%s = %q, want %q", key, byKey[key], value)
		}
	}
}

func TestEveryKeyThatIsWrittenIsUnderOneOfTheKeysThatAreTakenAway(t *testing.T) {
	for _, v := range RegistryValues(`C:\mvd.exe`) {
		found := false
		for _, k := range RegistryKeys() {
			found = found || v.Key == k || strings.HasPrefix(v.Key, k+`\`)
		}
		if !found {
			t.Errorf("%s would be left behind by an uninstall", v.Key)
		}
	}
}

func TestLinuxIsToldTheExtensionAndTheIcon(t *testing.T) {
	xml := MimeXML()
	for _, want := range []string{`type="application/x-mvd-list"`, `<glob pattern="*.mvd"/>`, `<icon name="mvd"/>`} {
		if !strings.Contains(xml, want) {
			t.Errorf("the mime description lacks %q", want)
		}
	}
}

func TestAListIsFoundAmongTheArgumentsOfAStart(t *testing.T) {
	if got := ListInArgs([]string{"mvd.exe", `C:\Music\Rock.MVD`}); got != `C:\Music\Rock.MVD` {
		t.Errorf("got %q", got)
	}
	if got := ListInArgs([]string{"mvd.exe", "-move"}); got != "" {
		t.Errorf("got %q", got)
	}
}
