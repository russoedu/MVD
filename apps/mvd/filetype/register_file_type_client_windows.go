package filetype

import (
	"errors"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

// Register tells the system that .mvd files are download lists of the program at path. It
// can be called at every start: what is already right is written again as it was.
func Register(program string) error {
	for _, v := range RegistryValues(program) {
		key, _, err := registry.CreateKey(registry.CURRENT_USER, v.Key, registry.SET_VALUE)
		if err != nil {
			return err
		}
		err = key.SetStringValue(v.Name, v.Value)
		_ = key.Close()
		if err != nil {
			return err
		}
	}
	notifyShell()
	return nil
}

// Unregister takes the kind of file away again. What is not there is not a problem.
func Unregister() error {
	var errs []error
	for _, k := range RegistryKeys() {
		if err := deleteTree(k); err != nil {
			errs = append(errs, err)
		}
	}
	notifyShell()
	return errors.Join(errs...)
}

// deleteTree deletes a key and what is under it.
func deleteTree(path string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, path, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	names, err := key.ReadSubKeyNames(-1)
	_ = key.Close()
	if err != nil {
		return err
	}
	for _, name := range names {
		if err := deleteTree(path + `\` + name); err != nil {
			return err
		}
	}
	if err := registry.DeleteKey(registry.CURRENT_USER, path); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return err
	}
	return nil
}

// notifyShell tells the file manager that what it shows for files may have changed, so the
// new icon appears without signing out.
func notifyShell() {
	const assocChanged, flushNoWait = 0x08000000, 0x2000
	proc := syscall.NewLazyDLL("shell32.dll").NewProc("SHChangeNotify")
	_, _, _ = proc.Call(assocChanged, flushNoWait, 0, 0)
}
