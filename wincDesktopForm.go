package main

import (
	"github.com/leaanthony/winc"
	"github.com/leaanthony/winc/w32"
)

type DesktopForm struct {
	winc.Form
}

var WM_SHELLHOOK uint32
var WM_TASKBARCREATEDMESSAGE uint32
var TASKBARBUTTONCREATEDMESSAGE uint32

func NewDesktopForm(parent winc.Controller) *DesktopForm {
	dlg := new(DesktopForm)
	dlg.SetIsForm(true)

	winc.RegClassOnlyOnce("DesktopForm")

	dlg.SetHandle(winc.CreateWindow("DesktopForm", parent,
		w32.WS_EX_TOOLWINDOW|w32.WS_EX_NOACTIVATE, // EX_STYLE => https://learn.microsoft.com/en-us/windows/win32/winmsg/extended-window-styles
		w32.WS_POPUP)) // STYLE => https://learn.microsoft.com/en-us/windows/win32/winmsg/window-styles
	dlg.SetParent(parent)

	// Dlg forces display of focus rectangles, as soon as the user starts to type.
	w32.SendMessage(dlg.Handle(), w32.WM_CHANGEUISTATE, w32.UIS_INITIALIZE, 0)
	winc.RegMsgHandler(dlg)

	dlg.SetFont(winc.DefaultFont)
	dlg.SetText("Desktop")

	return dlg
}

func (dlg *DesktopForm) WndProc(msg uint32, wparam, lparam uintptr) uintptr {
	switch msg {
	case w32.WM_DISPLAYCHANGE:
		// https://learn.microsoft.com/en-us/windows/win32/gdi/wm-displaychange
		SM_CXSCREEN = int(w32.GetSystemMetrics(w32.SM_CXSCREEN)) // Desktop width
		SM_CYSCREEN = int(w32.GetSystemMetrics(w32.SM_CYSCREEN)) // Desktop height
		SM_XVIRTUALSCREEN = int(w32.GetSystemMetrics(w32.SM_XVIRTUALSCREEN))
		SM_YVIRTUALSCREEN = int(w32.GetSystemMetrics(w32.SM_YVIRTUALSCREEN))
		SM_CXVIRTUALSCREEN = int(w32.GetSystemMetrics(w32.SM_CXVIRTUALSCREEN))
		SM_CYVIRTUALSCREEN = int(w32.GetSystemMetrics(w32.SM_CYVIRTUALSCREEN))

		w32.SetWindowPos(dlg.Handle(), w32.HWND_BOTTOM, SM_XVIRTUALSCREEN, SM_YVIRTUALSCREEN, SM_CXVIRTUALSCREEN, SM_CYVIRTUALSCREEN, w32.SWP_SHOWWINDOW)

	case w32.WM_CLOSE:
		dlg.Close()
	case w32.WM_DESTROY:
		if dlg.Parent() == nil {
			winc.Exit()
		}
	case w32.WM_HOTKEY:
		i := wparam

		switch {
		case config.Hotkey[i].ShellExecute != "":
			shellExecute(ResolveVariables(config.Hotkey[i].ShellExecute), config.Hotkey[i].Args, config.Hotkey[i].Hidden)
		case config.Hotkey[i].CreateProcess != "":
			createProcess(ResolveVariables(config.Hotkey[i].CreateProcess), config.Hotkey[i].Args, config.Hotkey[i].Hidden)
		case config.Hotkey[i].OpenProcess != "":
			openProcess(ResolveVariables(config.Hotkey[i].OpenProcess), config.Hotkey[i].Args, config.Hotkey[i].Hidden)
		}
	default:
		// log.Printf("DesktopForm WndProc (%d, 0x%x)\n", msg, msg)
	}
	return w32.DefWindowProc(dlg.Handle(), msg, wparam, lparam)
}
