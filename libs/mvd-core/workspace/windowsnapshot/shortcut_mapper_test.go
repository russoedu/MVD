package windowsnapshot

import (
	"reflect"
	"testing"
)

func TestShortcutOfReadsTheModifiersAndTheKey(t *testing.T) {
	cases := map[string]*Shortcut{
		"Ctrl+Shift+S": {Ctrl: true, Shift: true, Key: "S"},
		"Ctrl+,":       {Ctrl: true, Key: ","},
		"Ctrl++":       {Ctrl: true, Key: "+"},
		"Del":          {Key: "Del"},
		"Shift+Up":     {Shift: true, Key: "Up"},
		"F1 / Ctrl+K":  {Key: "F1"},
		"":             nil,
	}
	for text, want := range cases {
		if got := ShortcutOf(text); !reflect.DeepEqual(got, want) {
			t.Errorf("ShortcutOf(%q) = %+v, want %+v", text, got, want)
		}
	}
}
