package install

import (
	"errors"

	"golang.org/x/sys/windows/registry"
)

// uninstallKeyPath is where Windows looks for the programs listed in Settings > Apps.
const uninstallKeyPath = `Software\Microsoft\Windows\CurrentVersion\Uninstall\MVD`

// registerUninstallEntry lists the program in Settings > Apps: for the person alone, or
// for every account when it is installed for everyone (which the app only does when it
// already runs as an administrator, so writing there is allowed).
func registerUninstallEntry(target installTarget, version string) error {
	root := registry.CURRENT_USER
	if target.Everyone {
		root = registry.LOCAL_MACHINE
	}

	return writeUninstallEntry(root, uninstallKeyPath, uninstallEntryFor(target, version))
}

func writeUninstallEntry(root registry.Key, path string, entry uninstallEntry) error {
	key, _, err := registry.CreateKey(root, path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer func() { _ = key.Close() }()

	for name, value := range entry.strings() {
		if err := key.SetStringValue(name, value); err != nil {
			return err
		}
	}
	for name, value := range entry.numbers() {
		if err := key.SetDWordValue(name, value); err != nil {
			return err
		}
	}

	return nil
}

// RemoveUninstallEntries takes the program out of Settings > Apps, for the person and,
// when this account may, for everyone. An entry that is not there is not a problem.
func RemoveUninstallEntries() error {
	return errors.Join(
		deleteUninstallEntry(registry.CURRENT_USER, uninstallKeyPath),
		deleteUninstallEntry(registry.LOCAL_MACHINE, uninstallKeyPath),
	)
}

func deleteUninstallEntry(root registry.Key, path string) error {
	err := registry.DeleteKey(root, path)
	if errors.Is(err, registry.ErrNotExist) {
		return nil
	}

	return err
}
