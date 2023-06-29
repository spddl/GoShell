/*
 * Copyright (C) 2019 The Winc Authors. All Rights Reserved.
 * Copyright (C) 2010-2013 Allen Dang. All Rights Reserved.
 */

package winc

import (
	"errors"
	"fmt"
	"math"
	"unsafe"

	"github.com/leaanthony/winc/w32"
)

type Bitmap struct {
	handle        w32.HBITMAP
	width, height int

	// hBmp               w32.HBITMAP
	hPackedDIB w32.HGLOBAL
	// size               Size // in native pixels
	dpi                int
	transparencyStatus transparencyStatus
}

type transparencyStatus byte

const (
	transparencyUnknown transparencyStatus = iota
	transparencyOpaque
	transparencyTransparent
)

func assembleBitmapFromHBITMAP(hbitmap w32.HBITMAP) (*Bitmap, error) {
	var dib w32.DIBSECTION
	if w32.GetObject(w32.HGDIOBJ(hbitmap), unsafe.Sizeof(dib), unsafe.Pointer(&dib)) == 0 {
		return nil, errors.New("GetObject for HBITMAP failed")
	}

	return &Bitmap{
		handle: hbitmap,
		width:  int(dib.DsBmih.BiWidth),
		height: int(dib.DsBmih.BiHeight),
	}, nil
}

func NewBitmapFromFile(filepath string, background Color) (*Bitmap, error) {
	var gpBitmap *uintptr
	var err error

	gpBitmap, err = w32.GdipCreateBitmapFromFile(filepath)
	if err != nil {
		return nil, err
	}
	defer w32.GdipDisposeImage(gpBitmap)

	var hbitmap w32.HBITMAP
	// Reverse RGB to BGR to satisfy gdiplus color schema.
	hbitmap, err = w32.GdipCreateHBITMAPFromBitmap(gpBitmap, uint32(RGB(background.B(), background.G(), background.R())))
	if err != nil {
		return nil, err
	}

	return assembleBitmapFromHBITMAP(hbitmap)
}

func NewBitmapFromResource(instance w32.HINSTANCE, resName *uint16, resType *uint16, background Color) (*Bitmap, error) {
	var gpBitmap *uintptr
	var err error
	var hRes w32.HRSRC

	hRes, err = w32.FindResource(w32.HMODULE(instance), resName, resType)
	if err != nil {
		return nil, err
	}
	resSize := w32.SizeofResource(w32.HMODULE(instance), hRes)
	pResData := w32.LockResource(w32.LoadResource(w32.HMODULE(instance), hRes))
	resBuffer := w32.GlobalAlloc(w32.GMEM_MOVEABLE, resSize)
	pResBuffer := w32.GlobalLock(resBuffer)
	w32.MoveMemory(pResBuffer, pResData, resSize)

	stream := w32.CreateStreamOnHGlobal(resBuffer, false)

	gpBitmap, err = w32.GdipCreateBitmapFromStream(stream)
	if err != nil {
		return nil, err
	}
	defer stream.Release()
	defer w32.GlobalUnlock(resBuffer)
	defer w32.GlobalFree(resBuffer)
	defer w32.GdipDisposeImage(gpBitmap)

	var hbitmap w32.HBITMAP
	// Reverse gform.RGB to BGR to satisfy gdiplus color schema.
	hbitmap, err = w32.GdipCreateHBITMAPFromBitmap(gpBitmap, uint32(RGB(background.B(), background.G(), background.R())))
	if err != nil {
		return nil, err
	}

	return assembleBitmapFromHBITMAP(hbitmap)
}

func (bm *Bitmap) Dispose() {
	if bm.handle != 0 {
		w32.DeleteObject(w32.HGDIOBJ(bm.handle))
		bm.handle = 0
	}
}

func (bm *Bitmap) GetHBITMAP() w32.HBITMAP {
	return bm.handle
}

func (bm *Bitmap) Size() (int, int) {
	return bm.width, bm.height
}

func (bm *Bitmap) Height() int {
	return bm.height
}

func (bm *Bitmap) Width() int {
	return bm.width
}

// func (bm *Bitmap) handle() int {
// 	return bm.handle
// }

func NewBitmapFromIcon(iconHandle uintptr, x, y int32) *Bitmap {
	hDC := w32.GetDC(0)
	hMemDC := w32.CreateCompatibleDC(hDC)
	hMemBmp := w32.CreateCompatibleBitmap(hDC, x, y)
	hOrgOBJ := w32.SelectObject(hMemDC, hMemBmp)
	w32.DrawIconEx(hMemDC, 0, 0, iconHandle, x, y, 0, 0, w32.DI_NORMAL)

	w32.SelectObject(hMemDC, hOrgOBJ)
	w32.DeleteDC(hMemDC)
	w32.ReleaseDC(0, hDC)
	w32.DestroyIcon(iconHandle)

	return &Bitmap{
		handle: hMemBmp,
		width:  int(x),
		height: int(y),
	}
}

// func x() error {
// 	hdc := w32.CreateCompatibleDC(0)
// 	if hdc == 0 {
// 		return fmt.Errorf("createCompatibleDC failed")
// 	}
// 	defer w32.DeleteDC(hdc)

// 	hBmpOld := w32.SelectObject(hdc, w32.HGDIOBJ(bmp.hBmp))
// 	if hBmpOld == 0 {
// 		return fmt.Errorf("SelectObject failed")
// 	}
// 	defer w32.SelectObject(hdc, hBmpOld)

// }

const inchesPerMeter float64 = 39.37007874

// newBitmap creates a bitmap with given size in native pixels and DPI.
func newBitmap(size w32.Size, transparent bool, dpi int) (bmp *Bitmap, err error) {
	hdc := w32.CreateCompatibleDC(0)
	if hdc == 0 {
		return nil, fmt.Errorf("createCompatibleDC failed")
	}
	defer w32.DeleteDC(hdc)

	bufSize := int(size.Width * size.Height * 4)

	var hdr w32.BITMAPINFOHEADER
	hdr.BiSize = uint32(unsafe.Sizeof(hdr))
	hdr.BiBitCount = 32
	hdr.BiCompression = w32.BI_RGB
	hdr.BiPlanes = 1
	hdr.BiWidth = int32(size.Width)
	hdr.BiHeight = int32(size.Height)
	hdr.BiSizeImage = uint32(bufSize)
	dpm := int32(math.Round(float64(dpi) * inchesPerMeter))
	hdr.BiXPelsPerMeter = dpm
	hdr.BiYPelsPerMeter = dpm

	var bitsPtr unsafe.Pointer

	hBmp := w32.CreateDIBSection(hdc, &hdr, w32.DIB_RGB_COLORS, &bitsPtr, 0, 0)
	switch hBmp {
	case 0, w32.ERROR_INVALID_PARAMETER:
		return nil, fmt.Errorf("createDIBSection failed")
	}

	if transparent {
		w32.GdiFlush()

		bits := (*[1 << 24]byte)(bitsPtr)

		for i := 0; i < bufSize; i += 4 {
			// Mark pixel as not drawn to by GDI.
			bits[i+3] = 0x01
		}
	}

	bmp, err = newBitmapFromHBITMAP(hBmp, dpi)
	return bmp, err
}

// newBitmapFromHBITMAP creates Bitmap from win.HBITMAP.
//
// The BiXPelsPerMeter and BiYPelsPerMeter fields of win.BITMAPINFOHEADER are unreliable (for
// loaded PNG they are both unset). Therefore, we require caller to specify DPI explicitly.
func newBitmapFromHBITMAP(hBmp w32.HBITMAP, dpi int) (bmp *Bitmap, err error) {
	var dib w32.DIBSECTION
	if w32.GetObject(w32.HGDIOBJ(hBmp), unsafe.Sizeof(dib), unsafe.Pointer(&dib)) == 0 {
		return nil, fmt.Errorf("GetObject failed")
	}

	bmih := &dib.DsBmih

	bmihSize := uintptr(unsafe.Sizeof(*bmih))
	pixelsSize := uintptr(int32(bmih.BiBitCount)*bmih.BiWidth*bmih.BiHeight) / 8

	totalSize := uintptr(bmihSize + pixelsSize)

	hPackedDIB := w32.GlobalAlloc(w32.GHND, uint32(totalSize))
	dest := w32.GlobalLock(hPackedDIB)
	defer w32.GlobalUnlock(hPackedDIB)

	src := unsafe.Pointer(&dib.DsBmih)

	w32.MoveMemory(dest, src, uint32(bmihSize))

	dest = unsafe.Pointer(uintptr(dest) + bmihSize)
	src = dib.DsBm.BmBits

	w32.MoveMemory(dest, src, uint32(pixelsSize))

	return &Bitmap{
		handle:     hBmp,
		hPackedDIB: hPackedDIB,
		width:      int(bmih.BiWidth),
		height:     int(bmih.BiHeight),
		// hBmp:       hBmp,
		// size: Size{
		// 	int(bmih.BiWidth),
		// 	int(bmih.BiHeight),
		// },
		dpi: dpi,
	}, nil
}

// hBitmapFromIcon creates a new win.HBITMAP with given size in native pixels and DPI, and paints
// the icon on it stretched.
func hBitmapFromIcon(icon *Icon, size w32.Size, dpi int) (w32.HBITMAP, error) {
	hdc := w32.GetDC(0)
	defer w32.ReleaseDC(0, hdc)

	hdcMem := w32.CreateCompatibleDC(hdc)
	if hdcMem == 0 {
		return 0, fmt.Errorf("CreateCompatibleDC failed")
	}
	defer w32.DeleteDC(hdcMem)

	var bi w32.BITMAPV5HEADER
	bi.BiSize = uint32(unsafe.Sizeof(bi))
	bi.BiWidth = int32(size.Width)
	bi.BiHeight = int32(size.Height)
	bi.BiPlanes = 1
	bi.BiBitCount = 32
	bi.BiCompression = w32.BI_RGB
	dpm := int32(math.Round(float64(dpi) * inchesPerMeter))
	bi.BiXPelsPerMeter = dpm
	bi.BiYPelsPerMeter = dpm
	// The following mask specification specifies a supported 32 BPP
	// alpha format for Windows XP.
	bi.BV4RedMask = 0x00FF0000
	bi.BV4GreenMask = 0x0000FF00
	bi.BV4BlueMask = 0x000000FF
	bi.BV4AlphaMask = 0xFF000000

	hBmp := w32.CreateDIBSection(hdcMem, &bi.BITMAPINFOHEADER, w32.DIB_RGB_COLORS, nil, 0, 0)
	switch hBmp {
	case 0, w32.ERROR_INVALID_PARAMETER:
		return 0, fmt.Errorf("CreateDIBSection failed")
	}

	hOld := w32.SelectObject(hdcMem, w32.HGDIOBJ(hBmp))
	defer w32.SelectObject(hdcMem, hOld)

	r := NewRect(0, 0, size.Width, size.Height)
	err := icon.drawStretched(hdcMem, *r)
	// err := icon.drawStretched(hdcMem, w32.RECT{Width: size.Width, Height: size.Height})
	if err != nil {
		return 0, err
	}

	return hBmp, nil
}

// NewBitmapFromIconForDPI creates a new bitmap with given size in native pixels and DPI and paints
// the icon on it.
func NewBitmapFromIconForDPI(icon *Icon, size w32.Size, dpi int) (*Bitmap, error) {
	hBmp, err := hBitmapFromIcon(icon, size, dpi)
	if err != nil {
		return nil, err
	}

	return newBitmapFromHBITMAP(hBmp, dpi)
}
