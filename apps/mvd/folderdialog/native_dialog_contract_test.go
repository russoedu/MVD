package folderdialog

import "testing"

func TestEachKindOfWindowSaysWhatItIsAskingFor(t *testing.T) {
	for kind, want := range map[dialogKind]string{pickFolder: "Choose a folder", openFile: "Open a file", saveFile: "Save as"} {
		if got := (dialogRequest{kind: kind}).title(); got != want {
			t.Errorf("kind %d is titled %q, want %q", kind, got, want)
		}
	}
}
