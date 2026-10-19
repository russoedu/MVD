package install

import (
	"strings"
	"testing"
)

func TestTheMenuEntryStartsTheProgramWithoutATerminal(t *testing.T) {
	entry := desktopEntry("/home/me/.local/bin/mvd")

	for _, want := range []string{"[Desktop Entry]", "Type=Application", "Name=MVD", "Exec=/home/me/.local/bin/mvd %f\n", "MimeType=application/x-mvd-list;\n", "Terminal=false"} {
		if !strings.Contains(entry, want) {
			t.Errorf("the entry lacks %q:\n%s", want, entry)
		}
	}
}

func TestAProgramPathWithSpacesOrSpecialCharactersIsQuotedAsOneWord(t *testing.T) {
	cases := map[string]string{
		"/home/my user/bin/mvd": `Exec="/home/my user/bin/mvd" %f`,
		`/home/me/a"b/mvd`:      `Exec="/home/me/a\"b/mvd" %f`,
		"/home/me/$HOME/mvd":    `Exec="/home/me/\$HOME/mvd" %f`,
	}
	for program, want := range cases {
		if got := desktopEntry(program); !strings.Contains(got, want+"\n") {
			t.Errorf("for %q want %s in:\n%s", program, want, got)
		}
	}
}
