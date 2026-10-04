package install

import (
	"strings"
	"testing"
)

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
