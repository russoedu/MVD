//go:build windows

package install

import (
	"fmt"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	clsidShellLink = windows.GUID{Data1: 0x00021401, Data4: [8]byte{0xC0, 0, 0, 0, 0, 0, 0, 0x46}}
	iidShellLink   = windows.GUID{Data1: 0x000214F9, Data4: [8]byte{0xC0, 0, 0, 0, 0, 0, 0, 0x46}}
	iidPersistFile = windows.GUID{Data1: 0x0000010B, Data4: [8]byte{0xC0, 0, 0, 0, 0, 0, 0, 0x46}}

	procCoCreateInstance = windows.NewLazySystemDLL("ole32.dll").NewProc("CoCreateInstance")
)

// The places of the methods in the tables of the objects used: IUnknown's first three, then
// those of IShellLinkW and of IPersistFile.
const (
	methodQueryInterface = 0
	methodRelease        = 2

	linkSetDescription      = 7
	linkSetWorkingDirectory = 9
	linkSetPath             = 20

	fileSave = 6
)

// object is any COM object: a pointer to the table of its methods.
type object struct{ table *[32]uintptr }

func (o *object) call(method int, args ...uintptr) uintptr {
	hr, _, _ := syscall.SyscallN(o.table[method], append([]uintptr{uintptr(unsafe.Pointer(o))}, args...)...)

	return hr
}

func (o *object) release() { o.call(methodRelease) }

// createShortcut makes the shortcut file at link, to target, through the shell's own
// interfaces: no program is started to do it. The paths are values given to the system,
// never text read as code.
func createShortcut(link, target string) error {
	errs := make(chan error, 1)
	go func() {
		// COM wants the thread set up, and to stay the same until it is done.
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		errs <- saveShortcut(link, target)
	}()

	return <-errs
}

func saveShortcut(link, target string) error {
	if err := windows.CoInitializeEx(0, windows.COINIT_APARTMENTTHREADED); err != nil {
		return err
	}
	defer windows.CoUninitialize()

	var raw unsafe.Pointer
	hr, _, _ := procCoCreateInstance.Call(uintptr(unsafe.Pointer(&clsidShellLink)), 0, windows.CLSCTX_INPROC_SERVER, uintptr(unsafe.Pointer(&iidShellLink)), uintptr(unsafe.Pointer(&raw)))
	if int32(hr) < 0 {
		return fmt.Errorf("cannot create a shortcut: %w", syscall.Errno(uint32(hr)))
	}
	shellLink := (*object)(raw)
	defer shellLink.release()

	for method, value := range map[int]string{
		linkSetPath:             target,
		linkSetWorkingDirectory: filepath.Dir(target),
		linkSetDescription:      "MVD, the music video downloader",
	} {
		text, err := windows.UTF16PtrFromString(value)
		if err != nil {
			return err
		}
		if hr := shellLink.call(method, uintptr(unsafe.Pointer(text))); int32(hr) < 0 {
			return fmt.Errorf("cannot describe the shortcut: %w", syscall.Errno(uint32(hr)))
		}
	}

	var file *object
	if hr := shellLink.call(methodQueryInterface, uintptr(unsafe.Pointer(&iidPersistFile)), uintptr(unsafe.Pointer(&file))); int32(hr) < 0 {
		return fmt.Errorf("cannot save the shortcut: %w", syscall.Errno(uint32(hr)))
	}
	defer file.release()

	name, err := windows.UTF16PtrFromString(link)
	if err != nil {
		return err
	}
	if hr := file.call(fileSave, uintptr(unsafe.Pointer(name)), 1); int32(hr) < 0 {
		return fmt.Errorf("cannot save the shortcut: %w", syscall.Errno(uint32(hr)))
	}

	return nil
}
