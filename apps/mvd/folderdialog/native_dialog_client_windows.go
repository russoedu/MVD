//go:build windows

package folderdialog

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// hasNativeDialog is true here: Windows shows its own file windows through the system's
// COM interfaces, with no program started to do it.
const hasNativeDialog = true

var (
	clsidFileOpenDialog = windows.GUID{Data1: 0xDC1C5A9C, Data2: 0xE88A, Data3: 0x4dde, Data4: [8]byte{0xA5, 0xA1, 0x60, 0xF8, 0x2A, 0x20, 0xAE, 0xF7}}
	clsidFileSaveDialog = windows.GUID{Data1: 0xC0B4E2F3, Data2: 0xBA21, Data3: 0x4773, Data4: [8]byte{0x8D, 0xBA, 0x33, 0x5E, 0xC9, 0x46, 0xEB, 0x8B}}
	iidFileDialog       = windows.GUID{Data1: 0x42F85136, Data2: 0xDB7E, Data3: 0x439C, Data4: [8]byte{0x85, 0xF1, 0xE4, 0x07, 0x5D, 0x13, 0x5F, 0xC8}}
	iidShellItem        = windows.GUID{Data1: 0x43826D1E, Data2: 0xE718, Data3: 0x42EE, Data4: [8]byte{0xBC, 0x55, 0xA1, 0xE2, 0x61, 0xC3, 0x7B, 0xFE}}

	ole32                        = windows.NewLazySystemDLL("ole32.dll")
	procCoCreateInstance         = ole32.NewProc("CoCreateInstance")
	shell32                      = windows.NewLazySystemDLL("shell32.dll")
	procSHCreateItemFromParsingN = shell32.NewProc("SHCreateItemFromParsingName")
)

// The places of the methods in the table of an IFileDialog (after IUnknown's three and
// IModalWindow's Show), and of an IShellItem.
const (
	methodRelease = 2

	dialogShow                = 3
	dialogSetFileTypes        = 4
	dialogSetFileTypeIndex    = 5
	dialogSetOptions          = 9
	dialogGetOptions          = 10
	dialogSetFolder           = 12
	dialogSetFileName         = 15
	dialogSetTitle            = 17
	dialogGetResult           = 20
	dialogSetDefaultExtension = 22

	itemGetDisplayName = 5
)

// Options of the window (FOS_*) and the form of a name asked of an item (SIGDN_*).
const (
	fosOverwritePrompt  = 0x2
	fosPickFolders      = 0x20
	fosForceFileSystem  = 0x40
	fosPathMustExist    = 0x800
	fosFileMustExist    = 0x1000
	sigdnFileSystemPath = 0x80058000

	// errCancelled is what showing the window answers when the person closes it
	// (HRESULT_FROM_WIN32(ERROR_CANCELLED)).
	errCancelled = 0x800704C7
)

// object is any COM object: a pointer to the table of its methods.
type object struct{ table *[32]uintptr }

// comDialog and shellItem are the two kinds of object used here.
type comDialog struct{ object }
type shellItem struct{ object }

// failed says whether an HRESULT is an error.
func failed(hr uintptr) bool { return int32(hr) < 0 }

func hresult(what string, hr uintptr) error {
	return fmt.Errorf("%s: %w", what, syscall.Errno(uint32(hr)))
}

func (o *object) release() {
	_, _, _ = syscall.SyscallN(o.table[methodRelease], uintptr(unsafe.Pointer(o)))
}

// nativeDialog shows the window on a thread of its own, set up the way COM wants, and waits.
// When ctx ends first it gives up waiting; the window then goes with the program.
func nativeDialog(ctx context.Context, req dialogRequest) (string, bool, error) {
	type answer struct {
		path   string
		chosen bool
		err    error
	}
	done := make(chan answer, 1)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		path, chosen, err := showWindow(req)
		done <- answer{path, chosen, err}
	}()
	select {
	case a := <-done:
		return a.path, a.chosen, a.err
	case <-ctx.Done():
		return "", false, ctx.Err()
	}
}

func showWindow(req dialogRequest) (path string, chosen bool, err error) {
	if err := windows.CoInitializeEx(0, windows.COINIT_APARTMENTTHREADED); err != nil {
		return "", false, err
	}
	defer windows.CoUninitialize()

	class := &clsidFileOpenDialog
	if req.kind == saveFile {
		class = &clsidFileSaveDialog
	}
	var raw unsafe.Pointer
	if hr, _, _ := procCoCreateInstance.Call(uintptr(unsafe.Pointer(class)), 0, windows.CLSCTX_INPROC_SERVER, uintptr(unsafe.Pointer(&iidFileDialog)), uintptr(unsafe.Pointer(&raw))); failed(hr) {
		return "", false, hresult("cannot create the file window", hr)
	}
	dialog := (*comDialog)(raw)
	defer dialog.release()

	if err := dialog.configure(req); err != nil {
		return "", false, err
	}
	hr, _, _ := syscall.SyscallN(dialog.table[dialogShow], uintptr(unsafe.Pointer(dialog)), 0)
	if uint32(hr) == errCancelled {
		return "", false, nil
	}
	if failed(hr) {
		return "", false, hresult("cannot show the file window", hr)
	}

	var item *shellItem
	if hr, _, _ := syscall.SyscallN(dialog.table[dialogGetResult], uintptr(unsafe.Pointer(dialog)), uintptr(unsafe.Pointer(&item))); failed(hr) {
		return "", false, hresult("cannot read what was chosen", hr)
	}
	defer item.release()

	var name *uint16
	if hr, _, _ := syscall.SyscallN(item.table[itemGetDisplayName], uintptr(unsafe.Pointer(item)), sigdnFileSystemPath, uintptr(unsafe.Pointer(&name))); failed(hr) {
		return "", false, hresult("cannot read the name of what was chosen", hr)
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(name))

	return windows.UTF16PtrToString(name), true, nil
}

// configure gives the window its kind of choice, its title, its place to start and, for a
// file, the files it shows first.
func (d *comDialog) configure(req dialogRequest) error {
	var options uint32
	if hr, _, _ := syscall.SyscallN(d.table[dialogGetOptions], uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(&options))); failed(hr) {
		return hresult("cannot read the options of the file window", hr)
	}
	switch req.kind {
	case pickFolder:
		options |= fosPickFolders | fosForceFileSystem | fosPathMustExist
	case openFile:
		options |= fosForceFileSystem | fosFileMustExist | fosPathMustExist
	case saveFile:
		options |= fosForceFileSystem | fosOverwritePrompt | fosPathMustExist
	}
	_, _, _ = syscall.SyscallN(d.table[dialogSetOptions], uintptr(unsafe.Pointer(d)), uintptr(options))

	title, err := windows.UTF16PtrFromString(req.title())
	if err != nil {
		return err
	}
	_, _, _ = syscall.SyscallN(d.table[dialogSetTitle], uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(title)))

	if info, err := os.Stat(req.start); req.start != "" && err == nil && info.IsDir() {
		if folder := itemFromPath(req.start); folder != nil {
			_, _, _ = syscall.SyscallN(d.table[dialogSetFolder], uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(folder)))
			folder.release()
		}
	}
	if req.kind == pickFolder {
		return nil
	}
	return d.configureFiles(req)
}

// configureFiles says which files an open or save window shows: those of the program's kind
// first, then all of them; and, for saving, the name offered and the extension it gets.
func (d *comDialog) configureFiles(req dialogRequest) error {
	pattern := "*." + req.extension
	names := [2]*uint16{}
	specs := [2]*uint16{}
	for i, s := range []string{req.label + " (" + pattern + ")", pattern} {
		p, err := windows.UTF16PtrFromString(s)
		if err != nil {
			return err
		}
		names[i], specs[i] = p, p
	}
	allName, err := windows.UTF16PtrFromString("All files (*.*)")
	if err != nil {
		return err
	}
	allSpec, err := windows.UTF16PtrFromString("*.*")
	if err != nil {
		return err
	}
	// COMDLG_FILTERSPEC is two pointers: a name and a pattern.
	filters := [2][2]*uint16{{names[0], specs[1]}, {allName, allSpec}}
	_, _, _ = syscall.SyscallN(d.table[dialogSetFileTypes], uintptr(unsafe.Pointer(d)), uintptr(len(filters)), uintptr(unsafe.Pointer(&filters[0])))
	_, _, _ = syscall.SyscallN(d.table[dialogSetFileTypeIndex], uintptr(unsafe.Pointer(d)), 1)

	if req.kind != saveFile {
		return nil
	}
	if req.name != "" {
		name, err := windows.UTF16PtrFromString(req.name)
		if err != nil {
			return err
		}
		_, _, _ = syscall.SyscallN(d.table[dialogSetFileName], uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(name)))
	}
	ext, err := windows.UTF16PtrFromString(req.extension)
	if err != nil {
		return err
	}
	_, _, _ = syscall.SyscallN(d.table[dialogSetDefaultExtension], uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(ext)))
	runtime.KeepAlive(filters)

	return nil
}

// itemFromPath is the shell's own object for a folder, or nil when it cannot be made.
func itemFromPath(path string) *shellItem {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil
	}
	var item *shellItem
	hr, _, _ := procSHCreateItemFromParsingN.Call(uintptr(unsafe.Pointer(p)), 0, uintptr(unsafe.Pointer(&iidShellItem)), uintptr(unsafe.Pointer(&item)))
	if failed(hr) {
		return nil
	}

	return item
}

