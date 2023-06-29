package main

import (
	"log"

	"github.com/leaanthony/winc/w32"
)

func NewCanvasFromImage(canvashdc uintptr) {
	hdc := w32.CreateCompatibleDC(canvashdc)
	if hdc == 0 {
		log.Println("CreateCompatibleDC failed")
		return
	}
	defer w32.DeleteDC(hdc)

	hbmp := w32.CreateCompatibleBitmap(hdc, 16, 16)
	if hbmp == 0 {
		log.Println("CreateCompatibleBitmap failed")
		return
	}
	defer w32.DeleteObject(w32.HGDIOBJ(hbmp))

	oldbmp := w32.SelectObject(canvashdc, w32.HGDIOBJ(hbmp))
	if oldbmp == 0 {
		log.Println("SelectObject failed")
		return
	}
	defer w32.SelectObject(hdc, oldbmp)

	w32.SetViewportOrgEx(hdc, -int32(16), -int32(16), nil)
	w32.SetBrushOrgEx(hdc, -int(16), -int(16), nil)
}

// this ignore workRect
func SetPos(hwnd uintptr, x, y int) {
	w32.SetWindowPos(hwnd, w32.HWND_TOP, x, y, 0, 0, w32.SWP_NOSIZE)
}
