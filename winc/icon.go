/*
 * Copyright (C) 2019 The Winc Authors. All Rights Reserved.
 * Copyright (C) 2010-2013 Allen Dang. All Rights Reserved.
 */

package winc

import (
	"fmt"
	"math"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/leaanthony/winc/w32"
	"golang.org/x/sys/windows"
)

type Icon struct {
	handle w32.HICON

	filePath  string
	index     int
	res       *uint16
	dpi2hIcon map[int]w32.HICON
	size96dpi w32.Size
	isStock   bool
	hasIndex  bool
}

func NewIcon(handle uintptr) *Icon {
	return &Icon{
		handle: handle,
	}
}

func NewIconFromFile(path string) (*Icon, error) {
	ico := new(Icon)
	var err error
	p, _ := syscall.UTF16PtrFromString(path)
	if ico.handle = w32.LoadIcon(0, p); ico.handle == 0 {
		err = fmt.Errorf("cannot load icon from %s", path)
	}
	return ico, err
}

func NewIconFromResource(instance w32.HINSTANCE, resId uint16) (*Icon, error) {
	ico := new(Icon)
	var err error
	if ico.handle = w32.LoadIcon(instance, w32.MakeIntResource(resId)); ico.handle == 0 {
		err = fmt.Errorf("cannot load icon from resource with id %v", resId)
	}
	return ico, err
}

func ExtractIcon(fileName string, index int) (*Icon, error) {
	ico := new(Icon)
	var err error
	if ico.handle = w32.ExtractIcon(fileName, index); ico.handle == 0 || ico.handle == 1 {
		err = fmt.Errorf("cannot extract icon from %s at index %v", fileName, index)
	}
	return ico, err
}

func (ic *Icon) Destroy() bool {
	return w32.DestroyIcon(ic.handle)
}

func (ic *Icon) Handle() w32.HICON {
	return ic.handle
}

func ExtractIconToBitmap(fileName string, index int) *Bitmap {
	ico, err := ExtractIcon(fileName, index)
	if err != nil {
		return new(Bitmap)
	}

	hBmp, err := NewBitmapFromIconForDPI(ico, w32.Size{Width: 16, Height: 16}, 96)
	if err != nil {
		return new(Bitmap)
	}
	return hBmp
}

// const IDI_ICON1 = 100

// // https://stackoverflow.com/a/47481569
// func LoadIconFromResource(newSize, dpi int) *Icon {
// 	// ic := new(Icon)
// 	// ic.handle = handle
// 	// return ic

// 	hdc := w32.GetDC(0)
// 	logpixy := w32.GetDeviceCaps(hdc, w32.LOGPIXELSY)
// 	// log.Println("logpixy", logpixy)
// 	size := w32.MulDiv(newSize, logpixy, dpi)
// 	// log.Println("size", size)
// 	// hInst := w32.GetModuleHandle("")
// 	// h := w32.LoadIconWithScaleDown(0, nil, int32(size), int32(size), &ic.handle)
// 	// h := w32.LoadImage(ic.handle, w32.MakeIntResource(IDI_ICON1), w32.IMAGE_ICON, int32(size), int32(size), w32.LR_DEFAULTCOLOR)
// 	// log.Println("w32.MakeIntResource(IDI_ICON1)", w32.MakeIntResource(IDI_ICON1))
// 	hIcon := w32.LoadImage(
// 		0,
// 		w32.MakeIntResource(IDI_ICON1),
// 		w32.IMAGE_ICON,
// 		int32(size), // width
// 		int32(size), // height
// 		w32.LR_DEFAULTSIZE,
// 	)

// 	// log.Println("hIcon", hIcon)
// 	// return ic
// 	return &Icon{
// 		handle: hIcon,
// 	}

// 	// if ico.handle = w32.LoadIcon(instance, w32.MakeIntResource(resId)); ico.handle == 0 {
// 	// 	err = fmt.Errorf("cannot load icon from resource with id %v", resId)
// 	// }

// 	// hicon = (HICON)LoadImage(
// 	// 	AfxGetApp()->m_hInstance,
// 	// 			MAKEINTRESOURCE(IDI_ICON1),
// 	// 			IMAGE_ICON, size, size,
// 	// 			LR_DEFAULTCOLOR);
// 	// button.SetIcon(hicon);

// 	// return ico
// }

// sizeFromHICON returns icon size in native pixels.
func sizeFromHICON(hIcon w32.HICON) (w32.Size, error) {
	var ii w32.ICONINFO
	var bi w32.BITMAPINFO

	if !w32.GetIconInfo(hIcon, &ii) {
		return w32.Size{}, fmt.Errorf("GetIconInfo")
	}
	defer w32.DeleteObject(w32.HGDIOBJ(ii.HbmMask))

	var hBmp w32.HBITMAP
	if ii.HbmColor != 0 {
		hBmp = ii.HbmColor

		defer w32.DeleteObject(w32.HGDIOBJ(ii.HbmColor))
	} else {
		hBmp = ii.HbmMask
	}

	if w32.GetObject(w32.HGDIOBJ(hBmp), unsafe.Sizeof(bi), unsafe.Pointer(&bi)) == 0 {
		return w32.Size{}, fmt.Errorf("GetObject")
	}

	return w32.Size{Width: int(bi.BmiHeader.BiWidth), Height: int(bi.BmiHeader.BiHeight)}, nil
}

// NewIconFromHICONForDPI returns a new Icon at given DPI, using the specified win.HICON as source.
func NewIconFromHICONForDPI(hIcon w32.HICON, dpi int) (ic *Icon, err error) {
	// s, err := sizeFromHICON(hIcon)
	// if err != nil {
	// 	return nil, err
	// }
	return NewIcon(hIcon), nil
	// return newIconFromHICONAndSize(hIcon, SizeTo96DPI(s, dpi), dpi), nil
}

// func newIconFromHICONAndSize(hIcon w32.HICON, size w32.Size, dpi int) *Icon {
// 	// return &Icon{dpi2hIcon: map[int]win.HICON{dpi: hIcon}, size96dpi: size}
// 	return &Icon{dpi2hIcon: map[int]win.HICON{dpi: hIcon}, size96dpi: size}
// }

func (i *Icon) handleForDPIWithError(dpi int) (w32.HICON, error) {
	if i.dpi2hIcon == nil {
		i.dpi2hIcon = make(map[int]w32.HICON)
	} else if handle, ok := i.dpi2hIcon[dpi]; ok {
		return handle, nil
	}

	var hInst w32.HINSTANCE
	var name *uint16
	if i.filePath != "" {
		absFilePath, err := filepath.Abs(i.filePath)
		if err != nil {
			return 0, err
		}
		name, _ = syscall.UTF16PtrFromString(absFilePath)

	} else {
		if !i.isStock {
			if hInst = w32.GetModuleHandle(""); hInst == 0 {
				return 0, fmt.Errorf("GetModuleHandle")
			}
		}

		name = i.res
	}

	var size w32.Size
	if i.size96dpi.Width == 0 || i.size96dpi.Height == 0 {
		size = SizeFrom96DPI(defaultIconSize(), dpi)
	} else {
		size = SizeFrom96DPI(i.size96dpi, dpi)
	}

	var hIcon w32.HICON

	if i.hasIndex {
		w32.SHDefExtractIcon(
			name,
			int32(i.index),
			0,
			nil,
			&hIcon,
			w32.MAKELONG(0, uint16(size.Width)))
		if hIcon == 0 {
			return 0, fmt.Errorf("SHDefExtractIcon")
		}
	} else {
		hr := w32.HICON(w32.LoadIconWithScaleDown(
			hInst,
			name,
			int32(size.Width),
			int32(size.Height),
			&hIcon))
		if hr < 0 || hIcon == 0 {
			return 0, fmt.Errorf("loadIconWithScaleDown")
		}
	}

	i.dpi2hIcon[dpi] = hIcon

	return hIcon, nil
}

func (i *Icon) drawStretched(hdc w32.HDC, bounds Rect) error {

	// dpi := int(float64(bounds.Width()) / float64(i.size96dpi.Width) * 96.0)
	// dpi := 96

	// hIcon, _ := i.handleForDPIWithError(dpi)
	// if hIcon == 0 {
	// 	var dpiAvailMax int
	// 	for dpiAvail, handle := range i.dpi2hIcon {
	// 		if dpiAvail > dpiAvailMax {
	// 			hIcon = handle
	// 			dpiAvailMax = dpiAvail
	// 		}
	// 		if dpiAvail > dpi {
	// 			break
	// 		}
	// 	}
	// }

	if !w32.DrawIconEx(hdc, 0, 0, i.handle, int32(bounds.Width()), int32(bounds.Height()), 0, 0, w32.DI_NORMAL) {
		return fmt.Errorf("DrawIconEx")
	}

	return nil
}

func scaleInt(value int, scale float64) int {
	return int(math.Round(float64(value) * scale))
}

// SizeFrom96DPI converts from 1/96" units to native pixels.
func SizeFrom96DPI(value w32.Size, dpi int) w32.Size {
	return scaleSize(value, float64(dpi)/96.0)
}

// SizeTo96DPI converts from native pixels to 1/96" units.
func SizeTo96DPI(value w32.Size, dpi int) w32.Size {
	return scaleSize(value, 96.0/float64(dpi))
}

func scaleSize(value w32.Size, scale float64) w32.Size {
	return w32.Size{
		Width:  scaleInt(value.Width, scale),
		Height: scaleInt(value.Height, scale),
	}
}

// defaultIconSize returns default small icon size in 1/92" units.
func defaultIconSize() w32.Size {
	return w32.Size{Width: int(w32.GetSystemMetricsForDpi(w32.SM_CXSMICON, 96)), Height: int(w32.GetSystemMetricsForDpi(w32.SM_CYSMICON, 96))}
}

// NewIconFromSysDLL returns a new Icon, as identified by index of
// size 16x16 from the system DLL identified by dllBaseName.
func NewIconFromSysDLL(dllBaseName string, index int) (*Icon, error) {
	return NewIconFromSysDLLWithSize(dllBaseName, index, 16)
}

// NewIconFromSysDLLWithSize returns a new Icon, as identified by
// index of the desired size from the system DLL identified by dllBaseName.
func NewIconFromSysDLLWithSize(dllBaseName string, index, size int) (*Icon, error) {
	system32, err := windows.GetSystemDirectory()
	if err != nil {
		return nil, err
	}

	return checkNewIcon(&Icon{filePath: filepath.Join(system32, dllBaseName+".dll"), index: index, hasIndex: true, size96dpi: w32.Size{Width: size, Height: size}})
}

func checkNewIcon(icon *Icon) (*Icon, error) {
	if _, err := icon.handleForDPIWithError(96); err != nil {
		return nil, err
	}

	return icon, nil
}
