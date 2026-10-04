package install

import (
	"fmt"
	"os"
	"testing"

	"golang.org/x/sys/windows/registry"
)

func TestTheEntryIsWrittenToAndRemovedFromTheRegistry(t *testing.T) {
	path := fmt.Sprintf(`Software\MVDTest-%d`, os.Getpid())
	t.Cleanup(func() { _ = registry.DeleteKey(registry.CURRENT_USER, path) })
	entry := uninstallEntryFor(installTarget{Folder: `C:\x\MVD`, Program: `C:\x\MVD\mvd.exe`}, "4.5.6")

	if err := writeUninstallEntry(registry.CURRENT_USER, path, entry); err != nil {
		t.Fatal(err)
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, path, registry.QUERY_VALUE)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = key.Close() }()
	for name, want := range entry.strings() {
		if got, _, err := key.GetStringValue(name); err != nil || got != want {
			t.Errorf("%s = %q (%v), want %q", name, got, err, want)
		}
	}
	for name, want := range entry.numbers() {
		if got, _, err := key.GetIntegerValue(name); err != nil || uint32(got) != want {
			t.Errorf("%s = %d (%v), want %d", name, got, err, want)
		}
	}

	if err := deleteUninstallEntry(registry.CURRENT_USER, path); err != nil {
		t.Fatal(err)
	}
	if _, err := registry.OpenKey(registry.CURRENT_USER, path, registry.QUERY_VALUE); err == nil {
		t.Error("the key is still there")
	}
}

func TestRemovingAnEntryThatIsNotThereIsNotAnError(t *testing.T) {
	path := fmt.Sprintf(`Software\MVDTest-missing-%d`, os.Getpid())

	if err := deleteUninstallEntry(registry.CURRENT_USER, path); err != nil {
		t.Errorf("err = %v", err)
	}
}
