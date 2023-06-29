package main

import (
	"log"
	"unsafe"

	"github.com/leaanthony/winc"
	"github.com/leaanthony/winc/w32"
)

type taskList struct {
	PushButtonList []*TaskItem
	centered       bool
}

func ContextMenuTask() *winc.MenuItem {
	popupMn := winc.NewContextMenu()

	dwmToggleContextMenu := popupMn.AddItem("DWM Toggle", winc.NoShortcut)
	dwmToggleContextMenu.OnClick().Bind(func(e *winc.Event) {
		hwnd := e.Sender.(*TaskItem).hWnd
		var policyParameter w32.DWMNCRENDERINGPOLICY
		if getDWMactive(hwnd) {
			// https://docs.microsoft.com/en-us/windows/win32/api/dwmapi/nf-dwmapi-dwmsetwindowattribute
			policyParameter = w32.DWMNCRP_DISABLED
		} else {
			// https://docs.microsoft.com/en-us/windows/win32/api/dwmapi/nf-dwmapi-dwmsetwindowattribute
			policyParameter = w32.DWMNCRP_ENABLED
		}

		// Use with DwmSetWindowAttribute. Sets the non-client rendering policy. The pvAttribute parameter points to a value from the DWMNCRENDERINGPOLICY enumeration.
		w32.DwmSetWindowAttribute(hwnd, w32.DWMWA_NCRENDERING_POLICY, unsafe.Pointer(&policyParameter), uint32(unsafe.Sizeof(policyParameter)))

		// Use with DwmSetWindowAttribute. Forces the window to display an iconic thumbnail or peek representation (a static bitmap), even if a live or snapshot representation of the window is available. This value is normally set during a window's creation, and not changed throughout the window's lifetime. Some scenarios, however, might require the value to change over time. The pvAttribute parameter points to a value of type BOOL. TRUE to require a iconic thumbnail or peek representation; otherwise, FALSE.
		w32.DwmSetWindowAttribute(hwnd, w32.DWMWA_FORCE_ICONIC_REPRESENTATION, unsafe.Pointer(&policyParameter), uint32(unsafe.Sizeof(policyParameter)))
	})

	exitContextMenu := popupMn.AddItem("End Task", winc.NoShortcut)
	exitContextMenu.OnClick().Bind(func(e *winc.Event) {
		// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-endtask
		w32.SendMessage(e.Sender.(*TaskItem).hWnd, w32.WM_CLOSE, 0, 0)
	})

	return popupMn
}

func getDWMactive(hWnd uintptr) bool {
	var isNCRenderingEnabled = w32.DWMNCRP_USEWINDOWSTYLE
	if w32.DwmGetWindowAttribute(hWnd, w32.DWMWA_NCRENDERING_ENABLED, &isNCRenderingEnabled, uint32(unsafe.Sizeof(isNCRenderingEnabled))) != 0 {
		log.Println("getDWMactive Err", hWnd)
	}
	return isNCRenderingEnabled == 1
}

func (tl *taskList) Add(parent winc.Controller, hWnd uintptr) {
	if !FilterhWnd(hWnd) {
		return
	}

	windowTitle := WindowTitle(hWnd)
	// if windowTitle == "" {
	// 	return // Firefox Bug
	// }

	btn := NewTaskItem(parent)
	btn.hWnd = hWnd

	btn.SetText(windowTitle)

	if ico := GetAppIcon(hWnd); ico != nil {
		btn.Icon = ico.Handle()
	}

	btn.OnPaint().Bind(func(arg *winc.Event) {
		t, _ := arg.Sender.(*TaskItem)

		if p, ok := arg.Data.(*winc.PaintEventData); ok {
			// TODO: more options Border, Background

			p.Canvas.DrawFillRect(
				winc.NewRect(0, 0, config.Taskbar.Button.Size.Width, config.Taskbar.Button.Size.Height),
				winc.NewPen(w32.PS_GEOMETRIC, 0, winc.NewSolidColorBrush(winc.RGB(24, 24, 24))),
				winc.NewSolidColorBrush(winc.RGB(byte(config.Taskbar.Button.Bgcolor.R), byte(config.Taskbar.Button.Bgcolor.G), byte(config.Taskbar.Button.Bgcolor.B))),
			)

			// Icon
			left := 2
			if t.Icon != 0 {
				left = 34
				var iconSize = 24
				p.Canvas.DrawIconEx(winc.NewIcon(t.Icon), int32((config.Taskbar.Height-iconSize)/2), int32((config.Taskbar.Height-iconSize)/2), int32(iconSize), int32(iconSize), 0, 0, w32.DI_NORMAL)
			}

			// Text
			text := arg.Sender.Text()
			rc := winc.NewRect(left, 0, config.Taskbar.Button.Size.Width, config.Taskbar.Button.Size.Height)
			color := winc.RGB(byte(config.Taskbar.Button.Textcolor.R), byte(config.Taskbar.Button.Textcolor.G), byte(config.Taskbar.Button.Textcolor.B))

			logfont, err := winc.NewLogFont(winc.LogFontDesc{Name: config.Taskbar.FontFamily, Height: config.Taskbar.FontSize})
			if err != nil {
				log.Println(err)
			}

			// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-drawtext
			p.Canvas.DrawText(text, rc, uint(w32.DT_LEFT|w32.DT_NOCLIP|w32.DT_VCENTER|w32.DT_SINGLELINE|w32.DT_END_ELLIPSIS|w32.DT_NOPREFIX), logfont.GetFONT(), color)

			// TODO: Try GDI+ for an AA Font
			// gdiplus.New
			// bmp := CreateDPIAwareFont()
			// p.Canvas.DrawBitmap()
		}
	})

	btn.OnLBDown().Bind(func(arg *winc.Event) {
		if b, ok := arg.Sender.(*TaskItem); ok {
			// https://github.com/dremin/RetroBar/blob/eb3683d49b8431e2c6e99eb72ea10813eea0d29d/RetroBar/Controls/TaskButton.xaml.cs#L159-L160
			// BUG: something is still odd but for now this is ok
			if w32.IsWindowVisible(b.hWnd) {
				if w32.IsIconic(b.hWnd) {
					w32.ShowWindow(w32.HWND(b.hWnd), w32.SW_RESTORE)
					w32.SetForegroundWindow(w32.HWND(w32.GetLastActivePopup(hWnd)))
				} else {
					w32.ShowWindow(w32.HWND(b.hWnd), w32.SW_MINIMIZE)
				}
			} else {
				w32.ShowWindow(w32.HWND(hWnd), w32.SW_SHOW)
				w32.SetForegroundWindow(w32.HWND(w32.GetLastActivePopup(hWnd)))
			}
		}
	})

	btn.SetContextMenu(ContextMenuTask())
	tl.PushButtonList = append(tl.PushButtonList, btn)
}

func (tl *taskList) Remove(hwnd uintptr) {
	for i, v := range tl.PushButtonList {
		if v.hWnd == hwnd {
			tl.PushButtonList = append(tl.PushButtonList[:i], tl.PushButtonList[i+1:]...)
			v.Close()
			return
		}
	}
}

func (tl *taskList) Contains(hwnd uintptr) int {
	for i, v := range tl.PushButtonList {
		if v.Handle() == hwnd {
			return i
		}
	}
	return -1
}

func (tl *taskList) Refresh(parent winc.Controller, refresh bool) {
	var lastOffset int
	if tl.centered {
		lastOffset = (SM_CXSCREEN / 2) - len(tl.PushButtonList)*config.Taskbar.Button.Size.Width/2
	}

	var topOffset = 0
	var TaskbarButtonSizeWidth = config.Taskbar.Button.Size.Width
	var TaskbarButtonSizeHeight = config.Taskbar.Button.Size.Height

	var lastHandle uintptr
	if len(tl.PushButtonList) != 0 {
		if len(tl.PushButtonList)*TaskbarButtonSizeWidth > SM_CXSCREEN {
			// BUG: the items overlap each other
			TaskbarButtonSizeWidth = SM_CXSCREEN / len(tl.PushButtonList)
		}
	}
	for _, btn := range tl.PushButtonList {
		if refresh {
			var repaint = false
			if windowTitle := WindowTitle(btn.hWnd); windowTitle != "" && windowTitle != btn.Text() {
				btn.SetText(windowTitle)
				repaint = true
			}
			if ico := GetAppIcon(btn.hWnd); ico != nil && btn.Icon != ico.Handle() {
				btn.Icon = ico.Handle()
				repaint = true
			}
			if repaint {
				btn.Invalidate(true)
			}
		}

		// https://learn.microsoft.com/de-de/windows/win32/api/winuser/nf-winuser-setwindowpos
		w32.SetWindowPos(btn.Handle(), lastHandle, lastOffset, topOffset, TaskbarButtonSizeWidth, TaskbarButtonSizeHeight, w32.SWP_SHOWWINDOW)

		lastHandle = btn.Button.Handle()
		lastOffset += TaskbarButtonSizeWidth
	}
}

func FilterhWnd(hWnd uintptr) bool {
	// https://github.com/cairoshell/ManagedShell/blob/c6349cf2db8e656fd16e3f58c0cea016ed267cd9/src/ManagedShell.WindowsTasks/ApplicationWindow.cs#L311
	if !w32.IsWindow(hWnd) {
		return false
	}
	if !w32.IsWindowVisible(hWnd) {
		return false
	}

	// https://stackoverflow.com/questions/210504/enumerate-windows-like-alt-tab-does/210519#210519
	classname := w32.GetClassName(hWnd)
	if classname == "ApplicationFrameWindow" || classname == "Windows.UI.Core.CoreWindow" {
		return false
	}

	extendedWindowStyles := w32.GetWindowLongPtr(hWnd, w32.GWL_EXSTYLE)
	isToolWindow := (extendedWindowStyles & w32.WS_EX_TOOLWINDOW) != 0
	isAppWindow := (extendedWindowStyles & w32.WS_EX_APPWINDOW) != 0
	isNoActivate := (extendedWindowStyles & w32.WS_EX_NOACTIVATE) != 0
	ownerWin := w32.GetWindow(hWnd, w32.GW_OWNER)

	return (ownerWin == 0 || isAppWindow) && (!isNoActivate || isAppWindow) && !isToolWindow
}

// var magicDWord uintptr = 0x49474541

// func NewFilterhWnd(hWnd uintptr) bool {
// 	// https://github.com/lsdev/litestep-/blob/bbc1182d8abcd660e1c91c2207aa27b940898c9b/lsapi/bangs.cpp#L192
// 	/* Based off of Jugg's task.dll */
// 	if w32.IsWindow(hWnd) {
// 		if (w32.GetWindowLongPtr(hWnd, w32.GWLP_USERDATA) != magicDWord) && (w32.GetWindowLong(hWnd, w32.GWL_STYLE)&w32.WS_CHILD != 0) && (w32.GetWindowLong(hWnd, w32.GWL_STYLE)&w32.WS_VISIBLE == 0) && (w32.GetWindowLong(hWnd, w32.GWL_EXSTYLE)&w32.WS_EX_TOOLWINDOW != 0) {
// 			return true
// 		}
// 	}
// 	return false
// }
