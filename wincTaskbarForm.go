package main

import (
	"log"
	"sync"
	"unsafe"

	"github.com/leaanthony/winc"
	"github.com/leaanthony/winc/w32"
)

type TaskbarForm struct {
	winc.Form
	tl      *taskList
	ExStyle uint32
	mu      sync.Mutex
}

func NewTaskbarForm(parent winc.Controller, tl *taskList) *TaskbarForm {
	dlg := new(TaskbarForm)
	dlg.tl = tl
	dlg.SetIsForm(true)

	winc.RegClassOnlyOnce("TaskbarForm")

	dlg.SetHandle(winc.CreateWindow("TaskbarForm", parent,
		w32.WS_EX_TOOLWINDOW|w32.WS_EX_TRANSPARENT,      // EX_STYLE => https://learn.microsoft.com/en-us/windows/win32/winmsg/extended-window-styles
		w32.WS_POPUP|w32.WS_CLIPCHILDREN|w32.WS_VISIBLE, // STYLE => https://learn.microsoft.com/en-us/windows/win32/winmsg/window-styles
	))
	dlg.SetParent(parent)

	// Dlg forces display of focus rectangles, as soon as the user starts to type.
	w32.SendMessage(dlg.Handle(), w32.WM_CHANGEUISTATE, w32.UIS_INITIALIZE, 0)
	winc.RegMsgHandler(dlg)

	dlg.SetFont(winc.DefaultFont)
	dlg.SetText("Taskbar")

	// prevent other shells from working properly
	w32.SetTaskmanWindow(dlg.Handle())
	/*
		Note that custom shell applications do not receive WH_SHELL messages.
		Therefore, any application that registers itself as the default shell must call the SystemParametersInfo function
		before it (or any other application) can receive WH_SHELL messages.
		This function must be called with SPI_SETMINIMIZEDMETRICS and a MINIMIZEDMETRICS structure.
		Set the iArrange member of this structure to ARW_HIDE.
	*/
	SetMinimizedMetrics()
	w32.RegisterShellHookWindow(dlg.Handle())

	WM_SHELLHOOK = w32.RegisterWindowMessage(UTF16PtrFromString("SHELLHOOK"))
	WM_TASKBARCREATEDMESSAGE = w32.RegisterWindowMessage(UTF16PtrFromString("TaskbarCreated"))
	TASKBARBUTTONCREATEDMESSAGE = w32.RegisterWindowMessage(UTF16PtrFromString("TaskbarButtonCreated"))

	dlg.ExStyle = uint32(w32.GetWindowLong(dlg.Handle(), w32.GWL_EXSTYLE))

	return dlg
}

func (dlg *TaskbarForm) CheckFullscreen(hWnd uintptr) {
	if ok := dlg.IsFullscreen(hWnd); !ok {
		dlg.mu.Lock()
		if dlg.ExStyle&w32.WS_EX_TOPMOST != 0 {
			dlg.ExStyle &^= w32.WS_EX_TOPMOST
			w32.SetWindowLong(dlg.Handle(), w32.GWL_EXSTYLE, dlg.ExStyle)
		}
		dlg.mu.Unlock()
		Powersaver.Try(false)
	} else {
		dlg.mu.Lock()
		if dlg.ExStyle&w32.WS_EX_TOPMOST == 0 {
			dlg.ExStyle |= w32.WS_EX_TOPMOST
			w32.SetWindowLong(dlg.Handle(), w32.GWL_EXSTYLE, dlg.ExStyle)
		}
		dlg.mu.Unlock()
		Powersaver.Try(true)
	}
}

func (dlg *TaskbarForm) IsFullscreen(hWnd uintptr) bool {
	if dlg.IsPrimaryMonitor(hWnd) {
		// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getwindowplacement
		wndpl := new(w32.WINDOWPLACEMENT)
		wndpl.Length = uint32(unsafe.Sizeof(wndpl))
		if ret := w32.GetWindowPlacement(hWnd, wndpl); ret {
			if wndpl.RcNormalPosition.Top == 0 && wndpl.RcNormalPosition.Left == 0 && int(wndpl.RcNormalPosition.Right) == SM_CXSCREEN && int(wndpl.RcNormalPosition.Bottom) == SM_CYSCREEN {
				return true
			} else {
				return false
			}
		}
	}
	return false
}

func (dlg *TaskbarForm) IsPrimaryMonitor(hWnd uintptr) bool {
	// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-monitorfromwindow
	return w32.MonitorFromWindow(hWnd, w32.MONITOR_DEFAULTTONEAREST) == w32.MonitorFromWindow(dlg.Handle(), w32.MONITOR_DEFAULTTONEAREST)
}

func (dlg *TaskbarForm) WndProc(msg uint32, wparam, lparam uintptr) uintptr {
	switch msg {
	case w32.WM_DISPLAYCHANGE:
		// https://learn.microsoft.com/en-us/windows/win32/gdi/wm-displaychange
		// log.Println("WM_DISPLAYCHANGE TaskbarForm")

		w32.SetWindowPos(dlg.Handle(), w32.HWND_TOPMOST, 0, 0, 0, 0, w32.SWP_NOACTIVATE|w32.SWP_NOSIZE|w32.SWP_NOMOVE)
		// tl.Refresh(dlg, false)

	case WM_SHELLHOOK:

		switch wparam { // https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-registershellhookwindow
		case w32.HSHELL_WINDOWCREATED:
			dlg.tl.Add(dlg, lparam)
			dlg.tl.Refresh(dlg, true)

			// dlg.CheckFullscreen(lparam)
			dlg.debounce(lparam)

			return 1

		case w32.HSHELL_WINDOWDESTROYED:
			dlg.tl.Remove(lparam)
			dlg.tl.Refresh(dlg, false)

			dlg.debounce(lparam)

			return 1
		case 3:
			// log.Println("HSHELL_ACTIVATESHELLWINDOW", lparam) // Not used.
			dlg.debounce(lparam)
		case 0x8004:
			/*
			 * Note: The ShellHook will always set the HighBit when there
			 * is any full screen app on the desktop, even if it does not
			 * have focus.  Because of this, we have no easy way to tell
			 * if the currently activated app is full screen or not.
			 * This is worked around by checking the window's actual size
			 * against the screen size.  The correct behavior for this is
			 * to hide when a full screen app is active, and to show when
			 * a non full screen app is active.
			 */
			if lparam == 0 {
				return 1
			}
			// log.Println("HSHELL_RUDEAPPACTIVATED", lparam, WindowTitle(lparam))
			dlg.debounce(lparam)

		case w32.HSHELL_WINDOWACTIVATED:
			// log.Println("HSHELL_WINDOWACTIVATED", lparam, WindowTitle(lparam))
			dlg.debounce(lparam)

		// case w32.HSHELL_GETMINRECT:
		// 	log.Println("HSHELL_GETMINRECT", lparam)

		// 	pshi := (*w32.SHELLHOOKINFO)(unsafe.Pointer(lparam))
		// 	log.Printf("%#v\n", pshi)

		// case 6:
		// 	log.Println("HSHELL_REDRAW")

		case 7: // HSHELL_TASKMAN
			h := dlg.Parent().Handle()
			w32.SendMessage(h, w32.WM_CONTEXTMENU, h, 0)
			return 1
		// case w32.HSHELL_LANGUAGE:
		// 	log.Println("HSHELL_LANGUAGE")
		// case 9:
		// 	log.Println("HSHELL_SYSMENU")
		case 10:
			log.Println("HSHELL_ENDTASK", lparam)

			// case w32.HSHELL_ACCESSIBILITYSTATE:
			// 	log.Println("HSHELL_ACCESSIBILITYSTATE")
			// case w32.HSHELL_APPCOMMAND:
			// 	log.Println("HSHELL_APPCOMMAND")
			// case w32.HSHELL_WINDOWREPLACED:
			// 	log.Println("HSHELL_WINDOWREPLACED")
			// case 14:
			// 	log.Println("HSHELL_WINDOWREPLACING")
			// case 16:
			// log.Println("HSHELL_MONITORCHANGED")
			// case 0x8000:
			// 	log.Println("HSHELL_HIGHBIT")
			// case 0x8006:
			// log.Println("HSHELL_FLASH")
			// case 0x35:
			// 	log.Println("HSHELL_UNKNOWN_35")
			// case 0x36:
			// 	log.Println("HSHELL_UNKNOWN_36")
			// https://github.com/groboclown/petronia/blob/486338023d19cee989e92f0c5692680f1a37811f/old_stuff/petronia/shell/native/windows_hook_event.py#L252

			// default:
			// log.Printf("default: %d, 0x%x", wparam, wparam)
		}

	case w32.WM_CLOSE:
		dlg.Close()
	case w32.WM_DESTROY:
		if dlg.Parent() == nil {
			winc.Exit()
		}
	}
	return w32.DefWindowProc(dlg.Handle(), msg, wparam, lparam)
}

// Hide minimized windows
func SetMinimizedMetrics() {
	var minMetrics w32.MINIMIZEDMETRICS
	minMetrics.CbSize = uint32(unsafe.Sizeof(minMetrics))

	var SPI_GETMINIMIZEDMETRICS uint = 0x002B
	var SPI_SETMINIMIZEDMETRICS uint = 0x002C
	var ARW_HIDE int32 = 0x8
	var SPIF_SENDCHANGE uint = 0x02
	w32.SystemParametersInfo(SPI_GETMINIMIZEDMETRICS, uint(unsafe.Sizeof(minMetrics)), uintptr(unsafe.Pointer(&minMetrics)), 0)
	minMetrics.IArrange |= ARW_HIDE
	w32.SystemParametersInfo(SPI_SETMINIMIZEDMETRICS, uint(unsafe.Sizeof(minMetrics)), uintptr(unsafe.Pointer(&minMetrics)), SPIF_SENDCHANGE)

	var SPI_SETMENUANIMATION uint = 0x1003
	// Disable menu animation, as it overrides the alpha blending
	w32.SystemParametersInfo(SPI_SETMENUANIMATION, 0, 0, SPIF_SENDCHANGE)
}

// https://stackoverflow.com/a/6267402
func SetWorkspace(rect *winc.Rect) {
	const SPI_SETWORKAREA = 47
	const SPIF_UPDATEINIFILE = 1
	const SPIF_SENDWININICHANGE = 2
	const SPIF_change = SPIF_UPDATEINIFILE | SPIF_SENDWININICHANGE

	w32.SystemParametersInfo(SPI_SETWORKAREA, 0, uintptr(unsafe.Pointer(rect)), SPIF_change)
}

// func isTaskbarPresent() bool {
// 	var abd w32.APPBARDATA
// 	abd.CbSize = uint32(unsafe.Sizeof(abd))
// 	res := w32.SHAppBarMessage(w32.ABM_GETTASKBARPOS, &abd)
// 	log.Println("res", res)
// 	log.Println(prettyPrint(abd))

// 	res = w32.SHAppBarMessage(w32.ABM_REMOVE, &abd)
// 	log.Println("res", res)
// 	log.Println(prettyPrint(abd))
// 	return res != 0
// }

func (tbf *TaskbarForm) SetTaskbarState(abState *w32.APPBARDATA) {
	var abd w32.APPBARDATA
	abd.CbSize = uint32(unsafe.Sizeof(abd))
	abd.HWnd = tbf.Handle()
	abd.LParam = uintptr(unsafe.Pointer(abState))
	w32.SHAppBarMessage(w32.ABM_SETSTATE, &abd)
}

func (tbf *TaskbarForm) GetTaskbarState() w32.APPBARDATA {
	var abd w32.APPBARDATA
	abd.CbSize = uint32(unsafe.Sizeof(abd))
	abd.HWnd = tbf.Handle()
	w32.SHAppBarMessage(w32.ABM_GETSTATE, &abd)
	return abd
}

func (tbf *TaskbarForm) ContextMenu() *winc.MenuItem {
	contextmenu := winc.NewContextMenu()

	hBmp := winc.GetBitmap(ResolveVariables("%SystemRoot%\\system32\\imageres.dll"), -150)

	exitContextMenu := contextmenu.AddItemWithBitmap("Taskmgr", winc.NoShortcut, hBmp)
	exitContextMenu.OnClick().Bind(func(_ *winc.Event) {
		shellExecute(ResolveVariables("%SystemRoot%\\system32\\Taskmgr.exe"), []string{}, false)
	})

	return contextmenu
}
