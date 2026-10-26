package windowsnapshot

import (
	"reflect"
	"testing"
)

func TestMnemonicsTakeTheFirstFreeLetterOfAWordThenAnyFreeLetter(t *testing.T) {
	got := Mnemonics([]string{"Undo", "Redo", "", "Cut", "Copy", "Paste", "Delete", "Move up", "Move down"})
	want := []string{"U", "R", "", "C", "o", "P", "D", "M", "v"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Mnemonics = %q, want %q", got, want)
	}
}

func TestMnemonicsSkipSymbolsAndGiveNoneWhenEveryLetterIsTaken(t *testing.T) {
	got := Mnemonics([]string{"► Load playlists", "Ab", "ba"})
	want := []string{"L", "A", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Mnemonics = %q, want %q", got, want)
	}
	if got := Mnemonics([]string{"a", "A"}); got[1] != "" {
		t.Fatalf("a second label with only a taken letter got %q", got[1])
	}
}
