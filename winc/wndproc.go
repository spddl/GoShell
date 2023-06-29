/*
 * Copyright (C) 2019 The Winc Authors. All Rights Reserved.
 * Copyright (C) 2010-2013 Allen Dang. All Rights Reserved.
 */

package winc

import (
	"unsafe"

	"github.com/leaanthony/winc/w32"
)

var wmInvokeCallback uint32

func init() {
	wmInvokeCallback = RegisterWindowMessage("WincV0.InvokeCallback")
}

func genPoint(p uintptr) (x, y int) {
	x = int(w32.GET_X_LPARAM(p))
	y = int(w32.GET_Y_LPARAM(p))
	return
}

func genMouseEventArg(wparam, lparam uintptr) *MouseEventData {
	var data MouseEventData
	data.Button = int(wparam)
	data.X, data.Y = genPoint(lparam)

	return &data
}

func genDropFilesEventArg(wparam uintptr) *DropFilesEventData {
	hDrop := w32.HDROP(wparam)

	var data DropFilesEventData
	_, fileCount := w32.DragQueryFile(hDrop, 0xFFFFFFFF)
	data.Files = make([]string, fileCount)

	var i uint
	for i = 0; i < fileCount; i++ {
		data.Files[i], _ = w32.DragQueryFile(hDrop, i)
	}

	data.X, data.Y, _ = w32.DragQueryPoint(hDrop)
	w32.DragFinish(hDrop)
	return &data
}

var lastMiddleMenu uintptr

func generalWndProc(hwnd w32.HWND, msg uint32, wparam, lparam uintptr) uintptr {
	// switch msg {
	// case w32.WM_HSCROLL:
	// 	//println("case w32.WM_HSCROLL")

	// case w32.WM_VSCROLL:
	// 	//println("case w32.WM_VSCROLL")
	// }

	if controller := GetMsgHandler(hwnd); controller != nil {
		ret := controller.WndProc(msg, wparam, lparam)

		// switch msg {
		// case 32, 132, 512, 289, 134, 70, 537:
		// case w32.WM_CAPTURECHANGED:
		// 	log.Println("WM_CAPTURECHANGED", wparam, lparam)
		// case w32.WM_UNINITMENUPOPUP:
		// 	log.Println("WM_UNINITMENUPOPUP", wparam, lparam)
		// case w32.WM_MENUSELECT:
		// 	log.Println("WM_MENUSELECT", wparam, lparam)
		// case w32.WM_EXITMENULOOP:
		// 	log.Println("WM_EXITMENULOOP", wparam, lparam)
		// case w32.WM_MENURBUTTONUP:
		// 	log.Println("WM_MENURBUTTONUP", wparam, lparam)
		// case w32.WM_CONTEXTMENU:
		// 	log.Println("WM_CONTEXTMENU", wparam, lparam)
		// case w32.WM_RBUTTONUP:
		// 	log.Println("WM_RBUTTONUP", wparam, lparam)

		// default:
		// 	// log.Printf("%d, 0x%x => %x, %x\n", msg, msg, wparam, lparam)
		// 	// 49192
		// 	// c028
		// }

		switch msg {
		case w32.WM_ERASEBKGND:
			return 1 // important
		case w32.WM_NOTIFY: //Reflect notification to control
			nm := (*w32.NMHDR)(unsafe.Pointer(lparam))
			if controller := GetMsgHandler(nm.HwndFrom); controller != nil {
				ret := controller.WndProc(msg, wparam, lparam)
				if ret != 0 {
					w32.SetWindowLong(hwnd, w32.DWL_MSGRESULT, uint32(ret))
					return w32.TRUE
				}
			}
		case w32.WM_COMMAND:
			if lparam != 0 { //Reflect message to control
				h := w32.HWND(lparam)
				if controller := GetMsgHandler(h); controller != nil {
					ret := controller.WndProc(msg, wparam, lparam)
					if ret != 0 {
						w32.SetWindowLong(hwnd, w32.DWL_MSGRESULT, uint32(ret))
						return w32.TRUE
					}
				}
			}
		case w32.WM_CLOSE:
			controller.OnClose().Fire(NewEvent(controller, nil))
		case w32.WM_KILLFOCUS:
			controller.OnKillFocus().Fire(NewEvent(controller, nil))
		case w32.WM_SETFOCUS:
			controller.OnSetFocus().Fire(NewEvent(controller, nil))
		case w32.WM_DROPFILES:
			controller.OnDropFiles().Fire(NewEvent(controller, genDropFilesEventArg(wparam)))
		case w32.WM_CONTEXTMENU:
			// log.Println("WM_CONTEXTMENU", wparam, lparam)
			if wparam != 0 { //Reflect message to control
				h := w32.HWND(wparam)
				if controller := GetMsgHandler(h); controller != nil {
					contextMenu := controller.ContextMenu()
					if contextMenu != nil {
						var x, y int32
						if lparam == 0 {
							_x, _y, _ := w32.GetCursorPos()
							x = int32(_x)
							y = int32(_y)
						} else {
							x = w32.GET_X_LPARAM(lparam)
							y = w32.GET_Y_LPARAM(lparam)
						}

						id := w32.TrackPopupMenuEx(
							contextMenu.hMenu,
							w32.TPM_NOANIMATION|w32.TPM_RETURNCMD, // |w32.TPM_RECURSE,
							x,
							y,
							controller.Handle(),
							nil)

						item := findMenuItemByID(int(id))
						// log.Printf("%#v - %d\n", item, id)
						if item != nil {
							item.OnClick().Fire(NewEvent(controller, &MouseContextData{Mouse: genMouseEventArg(wparam, lparam), Item: item}))
						}
						return 0
					}
				}
			}

		case w32.WM_MENURBUTTONUP:
			// log.Println("WM_MENURBUTTONUP", wparam, lparam)
			// https://devblogs.microsoft.com/oldnewthing/20120104-00/?p=8703
			// https://pastebin.com/raw/4ku9Ts5R
			// log.Println("wparam", wparam)                  // The zero-based index of the menu item on which the right mouse button was released.
			// log.Printf("lparam %d 0x%x\n", lparam, lparam) // A handle to the menu containing the item.

			// item := findMenuItemByID(int(lparam))
			// log.Printf("%#v\n", item)

			// xPos := w32.GET_X_LPARAM(lparam) // Wrong
			// yPos := w32.GET_Y_LPARAM(lparam)
			// log.Println("xPos", xPos)
			// log.Println("yPos", yPos)

		case w32.WM_MENUSELECT:
			// log.Printf("WM_MENUSELECT: %d 0x%x, %d 0x%x\n", wparam, wparam, lparam, lparam)
			// controller.OnLBDown().Fire(NewEvent(controller, genMouseEventArg(wparam, lparam)))
		case w32.WM_LBUTTONDOWN:
			controller.OnLBDown().Fire(NewEvent(controller, genMouseEventArg(wparam, lparam)))
		case w32.WM_LBUTTONUP:
			controller.OnLBUp().Fire(NewEvent(controller, genMouseEventArg(wparam, lparam)))
		case w32.WM_LBUTTONDBLCLK:
			controller.OnLBDbl().Fire(NewEvent(controller, genMouseEventArg(wparam, lparam)))
		case w32.WM_MBUTTONDOWN:
			// https://learn.microsoft.com/en-us/windows/win32/inputdev/wm-mbuttondown
			// log.Println("WM_MBUTTONDOWN", wparam, lparam)
			// controller.OnMBDown().Fire(NewEvent(controller, genMouseEventArg(wparam, lparam)))

			if wparam != 0 { // Reflect message to control
				contextMenu := controller.MiddleMenu()

				x, y, ok := w32.GetCursorPos()
				if !ok {
					x, y = genPoint(lparam)
				}

				if contextMenu != nil {
					// https://learn.microsoft.com/de-de/windows/win32/winmsg/wm-cancelmode
					w32.SendMessage(hwnd, w32.WM_CANCELMODE, 0, 0)

					id := w32.TrackPopupMenuEx(
						contextMenu.hMenu,
						w32.TPM_NOANIMATION|w32.TPM_RETURNCMD,
						int32(x),
						int32(y),
						controller.Handle(),
						nil)

					item := findMenuItemByID(int(id))
					// log.Printf("%#v - %d\n", item, id)
					if item != nil {
						item.OnMClick().Fire(NewEvent(controller, &MouseContextData{Mouse: genMouseEventArg(wparam, lparam), Item: item}))
					}

					return 0
				}
			}
		case w32.WM_MBUTTONUP:
			controller.OnMBUp().Fire(NewEvent(controller, genMouseEventArg(wparam, lparam)))
		case w32.WM_RBUTTONDOWN:
			controller.OnRBDown().Fire(NewEvent(controller, genMouseEventArg(wparam, lparam)))
		case w32.WM_RBUTTONUP:
			controller.OnRBUp().Fire(NewEvent(controller, genMouseEventArg(wparam, lparam)))
		case w32.WM_RBUTTONDBLCLK:
			controller.OnRBDbl().Fire(NewEvent(controller, genMouseEventArg(wparam, lparam)))
		case w32.WM_MOUSEMOVE:
			controller.OnMouseMove().Fire(NewEvent(controller, genMouseEventArg(wparam, lparam)))
		case w32.WM_PAINT:
			canvas := NewCanvasFromHwnd(hwnd)
			defer canvas.Dispose()
			controller.OnPaint().Fire(NewEvent(controller, &PaintEventData{Canvas: canvas}))
		case w32.WM_KEYUP:
			controller.OnKeyUp().Fire(NewEvent(controller, &KeyUpEventData{int(wparam), int(lparam)}))
		case w32.WM_SIZE:
			x, y := genPoint(lparam)
			controller.OnSize().Fire(NewEvent(controller, &SizeEventData{uint(wparam), x, y}))
		case wmInvokeCallback:
			controller.invokeCallbacks()
		}
		return ret
	}

	return w32.DefWindowProc(hwnd, uint32(msg), wparam, lparam)
}
