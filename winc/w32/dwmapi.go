package w32

import (
	"syscall"
	"unsafe"
)

var (
	moddwmapi = syscall.NewLazyDLL("dwmapi.dll")

	procDwmGetWindowAttribute = moddwmapi.NewProc("DwmGetWindowAttribute")
	procDwmSetWindowAttribute = moddwmapi.NewProc("DwmSetWindowAttribute")
)

func DwmGetWindowAttribute(hWnd HWND, dwAttribute DWMWINDOWATTRIBUTE, pvAttribute *DWMNCRENDERINGPOLICY, cbAttribute uint32) int32 {
	ret, _, _ := procDwmGetWindowAttribute.Call(
		uintptr(hWnd),
		uintptr(dwAttribute),
		uintptr(unsafe.Pointer(pvAttribute)),
		uintptr(cbAttribute))
	return int32(ret)
}
func DwmSetWindowAttribute(hwnd HWND, dwAttribute DWMWINDOWATTRIBUTE, pvAttribute LPCVOID, cbAttribute uint32) HRESULT {
	ret, _, _ := procDwmSetWindowAttribute.Call(
		hwnd,
		uintptr(dwAttribute),
		uintptr(pvAttribute),
		uintptr(cbAttribute))
	return HRESULT(ret)
}
