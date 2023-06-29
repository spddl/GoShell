package main

import (
	"syscall"

	"github.com/leaanthony/winc"
	"github.com/leaanthony/winc/w32"
)

func GetProcesses() []uintptr {
	var tasklist []uintptr
	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
		tasklist = append(tasklist, uintptr(h))
		return 1
	})
	w32.EnumWindows(cb, 0)
	return tasklist
}

func WindowTitle(hWnd uintptr) string {
	b := make([]uint16, syscall.MAX_PATH)
	_, err := w32.GetWindowTextW(syscall.Handle(hWnd), &b[0], int32(len(b)))
	if err == nil {
		return syscall.UTF16ToString(b)
	}
	return ""
}

var (
	ICON_SMALL  = 0
	ICON_BIG    = 1
	ICON_SMALL2 = 2

	GCL_HICON   = -14
	GCL_HICONSM = -34
)

func GetAppIcon(hwnd uintptr) *winc.Icon {
	// https://stackoverflow.com/a/24052117
	var iconHandle uintptr

	if iconHandle = w32.SendMessage(hwnd, w32.WM_GETICON, uintptr(ICON_SMALL2), 0); iconHandle != 0 {
		return winc.NewIcon(iconHandle)
	}

	if iconHandle = w32.SendMessage(hwnd, w32.WM_GETICON, uintptr(ICON_SMALL), 0); iconHandle != 0 {
		return winc.NewIcon(iconHandle)
	}

	if iconHandle = w32.SendMessage(hwnd, w32.WM_GETICON, uintptr(ICON_BIG), 0); iconHandle != 0 {
		return winc.NewIcon(iconHandle)
	}

	if iconHandle = w32.GetClassLongPtr(uintptr(hwnd), GCL_HICON); iconHandle != 0 {
		return winc.NewIcon(iconHandle)
	}

	if iconHandle = w32.GetClassLongPtr(uintptr(hwnd), GCL_HICONSM); iconHandle != 0 {
		return winc.NewIcon(iconHandle)
	}

	r := winc.Icon{}
	return &r
}

// NewIconFromHICONForDPI returns a new Icon at given DPI, using the specified w32.HICON as source.
// func NewIconFromHICONForDPI(hIcon w32.HICON, dpi int) (ic *Icon, err error) {
// 	s, err := sizeFromHICON(hIcon)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return newIconFromHICONAndSize(hIcon, SizeTo96DPI(s, dpi), dpi), nil
// }

// sizeFromHICON returns icon size in native pixels.
// func sizeFromHICON(hIcon w32.HICON) (Size, error) {
// 	var ii w32.ICONINFO
// 	var bi w32.BITMAPINFO

// 	if !w32.GetIconInfo(hIcon, &ii) {
// 		return Size{}, lastError("GetIconInfo")
// 	}
// 	defer w32.DeleteObject(w32.HGDIOBJ(ii.HbmMask))

// 	var hBmp w32.HBITMAP
// 	if ii.HbmColor != 0 {
// 		hBmp = ii.HbmColor

// 		defer w32.DeleteObject(w32.HGDIOBJ(ii.HbmColor))
// 	} else {
// 		hBmp = ii.HbmMask
// 	}

// 	if 0 == w32.GetObject(w32.HGDIOBJ(hBmp), unsafe.Sizeof(bi), unsafe.Pointer(&bi)) {
// 		return Size{}, newError("GetObject")
// 	}

// 	return Size{int(bi.BmiHeader.BiWidth), int(bi.BmiHeader.BiHeight)}, nil
// }

// func newIconFromHICONAndSize(hIcon w32.HICON, size Size, dpi int) *Icon {
// 	return &Icon{dpi2hIcon: map[int]w32.HICON{dpi: hIcon}, size96dpi: size}
// }
