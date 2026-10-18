package menubar

import (
	"strings"
	"testing"
)

func TestParseMenusReadsEntriesAndSeparators(t *testing.T) {
	menus, err := ParseMenus(`[{"name":"File","entries":[{"name":"Open","shortcut":"Ctrl+O"},{"separator":true},{"name":"Quit"}]}]`)
	if err != nil {
		t.Fatal(err)
	}
	if len(menus) != 1 || len(menus[0].Entries) != 3 {
		t.Fatalf("menus = %+v", menus)
	}
	if e := menus[0].Entries; e[0].Shortcut != "Ctrl+O" || !e[1].Separator || e[2].Name != "Quit" {
		t.Errorf("entries = %+v", e)
	}
}

func TestParseMenusRefusesWhatHasNoName(t *testing.T) {
	cases := map[string]string{
		`[{"name":"","entries":[]}]`:                     "has no name",
		`[{"name":"File","entries":[{"shortcut":"x"}]}]`: "has no name",
		`{"name":"File"}`:                                "not a list",
	}
	for text, want := range cases {
		if _, err := ParseMenus(text); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("ParseMenus(%s) = %v, want %q", text, err, want)
		}
	}
	if menus, err := ParseMenus("  "); err != nil || len(menus) != 0 {
		t.Errorf("empty text: %v, %v", menus, err)
	}
}

func TestMenusJSONWritesAnEmptyMenuAsAnEmptyList(t *testing.T) {
	got := MenusJSON([]Menu{{Name: "Help"}})
	if got != `[{"name":"Help","entries":[]}]` {
		t.Errorf("got %s", got)
	}
	if MenusJSON(nil) != "[]" {
		t.Error("no menus is []")
	}
}
