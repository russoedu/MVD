//go:build windows

package notification

import (
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32              = windows.NewLazySystemDLL("user32.dll")
	shell32             = windows.NewLazySystemDLL("shell32.dll")
	procRegisterClass   = user32.NewProc("RegisterClassExW")
	procCreateWindow    = user32.NewProc("CreateWindowExW")
	procDestroyWindow   = user32.NewProc("DestroyWindow")
	procDefWindowProc   = user32.NewProc("DefWindowProcW")
	procSetTimer        = user32.NewProc("SetTimer")
	procGetMessage      = user32.NewProc("GetMessageW")
	procDispatchMessage = user32.NewProc("DispatchMessageW")
	procPostQuit        = user32.NewProc("PostQuitMessage")
	procLoadIcon        = user32.NewProc("LoadIconW")
	procNotifyIcon      = shell32.NewProc("Shell_NotifyIconW")
)

const (
	className = "MVDNotice"
	// How long the icon stays, in milliseconds, so the balloon can be read.
	balloonStay = 11000

	wmTimer = 0x0113
	// hwndMessage is HWND_MESSAGE (-3): a window that is never shown.
	hwndMessage = ^uintptr(2)

	nimAdd    = 0
	nimModify = 1
	nimDelete = 2

	nifIcon = 0x2
	nifTip  = 0x4
	nifInfo = 0x10

	niifInfo  = 0x1
	niifError = 0x3

	idiInformation = 32516
)

// windowClass is WNDCLASSEXW.
type windowClass struct {
	size        uint32
	style       uint32
	proc        uintptr
	classExtra  int32
	windowExtra int32
	instance    windows.Handle
	icon        windows.Handle
	cursor      windows.Handle
	background  windows.Handle
	menuName    *uint16
	className   *uint16
	smallIcon   windows.Handle
}

// message is MSG.
type message struct {
	window  windows.Handle
	id      uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pointX  int32
	pointY  int32
	private uint32
}

// iconData is NOTIFYICONDATAW.
type iconData struct {
	size            uint32
	window          windows.Handle
	id              uint32
	flags           uint32
	callbackMessage uint32
	icon            windows.Handle
	tip             [128]uint16
	state           uint32
	stateMask       uint32
	info            [256]uint16
	version         uint32
	infoTitle       [64]uint16
	infoFlags       uint32
	guid            windows.GUID
	balloonIcon     windows.Handle
}

// showBalloon shows n as a balloon from a temporary notification-area icon, through the
// system's own interfaces and with no program started, and returns at once: the icon goes
// away by itself once the balloon could be read.
func showBalloon(n Notice) bool {
	go func() {
		// The window and its messages belong to the thread that made them.
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		balloon(n)
	}()

	return true
}

func balloon(n Notice) {
	name, err := windows.UTF16PtrFromString(className)
	if err != nil {
		return
	}
	proc := windows.NewCallback(func(window, id, wParam, lParam uintptr) uintptr {
		if id == wmTimer {
			_, _, _ = procPostQuit.Call(0)

			return 0
		}
		result, _, _ := procDefWindowProc.Call(window, id, wParam, lParam)

		return result
	})
	class := windowClass{proc: proc, className: name}
	class.size = uint32(unsafe.Sizeof(class))
	// Registering twice answers that the class exists, which is what is wanted.
	_, _, _ = procRegisterClass.Call(uintptr(unsafe.Pointer(&class)))

	window, _, _ := procCreateWindow.Call(0, uintptr(unsafe.Pointer(name)), 0, 0, 0, 0, 0, 0, hwndMessage, 0, 0, 0)
	if window == 0 {
		return
	}
	defer func() { _, _, _ = procDestroyWindow.Call(window) }()

	icon, _, _ := procLoadIcon.Call(0, idiInformation)
	data := iconData{window: windows.Handle(window), id: 1, flags: nifIcon | nifTip, icon: windows.Handle(icon)}
	data.size = uint32(unsafe.Sizeof(data))
	copy(data.tip[:len(data.tip)-1], windows.StringToUTF16("MVD"))
	if added, _, _ := procNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(&data))); added == 0 {
		return
	}
	defer func() { _, _, _ = procNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(&data))) }()

	data.flags = nifInfo
	data.infoFlags = niifInfo
	if n.Failure {
		data.infoFlags = niifError
	}
	copy(data.infoTitle[:len(data.infoTitle)-1], windows.StringToUTF16(clip(n.Title, noticeTitleLimit)))
	copy(data.info[:len(data.info)-1], windows.StringToUTF16(clip(n.Text, noticeTextLimit)))
	_, _, _ = procNotifyIcon.Call(nimModify, uintptr(unsafe.Pointer(&data)))

	_, _, _ = procSetTimer.Call(window, 1, balloonStay, 0)
	var m message
	for {
		if got, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0); int32(got) <= 0 {
			return
		}
		_, _, _ = procDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
}
