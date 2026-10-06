package ytdlp

import (
	"reflect"
	"testing"
)

func TestACommaSeparatedRuntimeListBecomesOneFlagPerRuntime(t *testing.T) {
	got := NormalizeArgs([]string{"-4", "--js-runtimes", "deno,node", "--no-warnings"})
	want := []string{"-4", "--js-runtimes", "deno", "--js-runtimes", "node", "--no-warnings"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestTheEqualsFormIsSplitToo(t *testing.T) {
	got := NormalizeArgs([]string{"--js-runtimes=deno, node"})
	want := []string{"--js-runtimes", "deno", "--js-runtimes", "node"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFlagsAlreadyInTheRightFormAreLeftAlone(t *testing.T) {
	args := []string{"-4", "--js-runtimes", "deno", "--js-runtimes", "node:C:\tools\node.exe", "--cookies", "c.txt"}
	if got := NormalizeArgs(args); !reflect.DeepEqual(got, args) {
		t.Errorf("got %v, want %v", got, args)
	}
}

func TestOtherArgumentsAndAnEmptyListPassThrough(t *testing.T) {
	if got := NormalizeArgs(nil); len(got) != 0 {
		t.Errorf("got %v for no arguments", got)
	}
	if got := NormalizeArgs([]string{"--js-runtimes", ""}); len(got) != 0 {
		t.Errorf("an empty runtime list adds nothing, got %v", got)
	}
	if got := NormalizeArgs([]string{"--js-runtimes"}); !reflect.DeepEqual(got, []string{"--js-runtimes"}) {
		t.Errorf("a trailing flag with no value is not ours to fix, got %v", got)
	}
}
