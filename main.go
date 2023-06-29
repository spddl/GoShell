package main

import (
	"log"
	"strings"
	"syscall"
	"unsafe"

	"github.com/leaanthony/winc"
	"github.com/leaanthony/winc/w32"
	"golang.org/x/sys/windows/registry"
)

type shell struct {
	mainWindow    *DesktopForm
	TaskbarWindow *TaskbarForm
}

type MonitorRect struct {
	Rect       w32.RECT
	LogPixels  uint16
	BitsPerPel uint32
}

var (
	PrimaryMonitor    MonitorRect
	AdditionalMonitor []MonitorRect
)

func main() {
	if config.Desktop.Contextmenu.DarkMode {
		w32.SetPreferredAppMode(w32.AllowDark)
	}

	// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-enumdisplaydevicesw
	var device *uint16
	var displayDevice w32.DISPLAY_DEVICE
	displayDevice.Cb = uint32(unsafe.Sizeof(displayDevice))

	var i uint32
	for i = 0; ; i++ {
		if w32.EnumDisplayDevices(device, i, &displayDevice, w32.EDD_NONE) { // EDD_GET_DEVICE_INTERFACE_NAME
			if displayDevice.StateFlags&w32.DISPLAY_DEVICE_ACTIVE == 0 {
				continue
			}

			deviceName := syscall.UTF16ToString(displayDevice.DeviceName[:])
			var lpszDeviceName = UTF16PtrFromString(deviceName)
			var dm w32.DEVMODEW
			dm.DmSize = uint16(unsafe.Sizeof(dm))

			if w32.EnumDisplaySettingsEx(lpszDeviceName, w32.ENUM_CURRENT_SETTINGS, &dm, 0) { // w32.EDS_ROTATEDMODE
				if displayDevice.StateFlags&w32.DISPLAY_DEVICE_PRIMARY_DEVICE != 0 {
					PrimaryMonitor = MonitorRect{
						Rect:       w32.RECT{Left: dm.DmPosition.X, Top: dm.DmPosition.Y, Right: int32(dm.DmPelsWidth), Bottom: int32(dm.DmPelsHeight)},
						LogPixels:  dm.DmLogPixels,
						BitsPerPel: dm.DmBitsPerPel,
					}
				} else {
					AdditionalMonitor = append(AdditionalMonitor, MonitorRect{
						Rect:       w32.RECT{Left: dm.DmPosition.X, Top: dm.DmPosition.Y, Right: int32(dm.DmPelsWidth), Bottom: int32(dm.DmPelsHeight)},
						LogPixels:  dm.DmLogPixels,
						BitsPerPel: dm.DmBitsPerPel,
					})

					// https://learn.microsoft.com/en-us/windows/win32/gdi/the-virtual-screen
					if PrimaryFromVirtualX > dm.DmPosition.X {
						PrimaryFromVirtualX = dm.DmPosition.X
					}
					if PrimaryFromVirtualY > dm.DmPosition.Y {
						PrimaryFromVirtualY = dm.DmPosition.Y
					}
				}
			}

		} else {
			break
		}
	}

	s := new(shell)
	s.mainWindow = NewDesktopForm(nil)
	s.mainWindow.SetSize(SM_CXVIRTUALSCREEN, SM_CYVIRTUALSCREEN)

	s.mainWindow.OnPaint().Bind(func(arg *winc.Event) {
		p, ok := arg.Data.(*winc.PaintEventData)
		if ok {
			wallpaperPath, backgroundColor := GetWallpaper()
			switch {
			case wallpaperPath != "":
				p.Canvas.DrawFillRect(
					winc.NewRect(SM_XVIRTUALSCREEN, SM_YVIRTUALSCREEN, SM_CXVIRTUALSCREEN, SM_CYVIRTUALSCREEN),
					winc.NewPen(w32.PS_GEOMETRIC, 0, winc.NewSolidColorBrush(winc.RGB(0, 0, 0))),
					winc.NewSolidColorBrush(winc.RGB(0, 0, 0)),
				)

				bmp, _ := winc.NewBitmapFromFile(wallpaperPath, winc.RGB(0, 0, 0))
				p.Canvas.SetStretchBltMode(w32.HALFTONE)
				p.Canvas.DrawStretchedBitmap(
					bmp,
					winc.NewRect(
						int(PrimaryMonitor.Rect.Left+(PrimaryFromVirtualX*-1)),
						int(PrimaryMonitor.Rect.Top+(PrimaryFromVirtualY*-1)),
						int(PrimaryMonitor.Rect.Right),
						int(PrimaryMonitor.Rect.Bottom),
					),
				)

			case backgroundColor != "":
				color := strings.Split(backgroundColor, " ")
				p.Canvas.DrawFillRect(
					winc.NewRect(SM_XVIRTUALSCREEN, SM_YVIRTUALSCREEN, SM_CXVIRTUALSCREEN, SM_CYVIRTUALSCREEN),
					winc.NewPen(w32.PS_GEOMETRIC, 0, winc.NewSolidColorBrush(winc.RGB(0, 0, 0))),
					winc.NewSolidColorBrush(winc.RGB(byte(StringToUint32(color[0])), byte(StringToUint32(color[1])), byte(StringToUint32(color[2])))),
				)

			default: // Should never happen
				p.Canvas.DrawFillRect(
					winc.NewRect(SM_XVIRTUALSCREEN, SM_YVIRTUALSCREEN, SM_CXVIRTUALSCREEN, SM_CYVIRTUALSCREEN),
					winc.NewPen(w32.PS_GEOMETRIC, 0, winc.NewSolidColorBrush(winc.RGB(0, 0, 0))),
					winc.NewSolidColorBrush(winc.RGB(0, 0, 0)),
				)
			}
		}
	})

	s.mainWindow.SetContextMenu(s.ContextMenu())
	s.mainWindow.SetMiddleMenuFunc(s.MiddleMenu)

	// Taskleiste
	tl := new(taskList)
	if config.Taskbar.IconPosition == "center" {
		tl.centered = true
	}
	s.TaskbarWindow = NewTaskbarForm(s.mainWindow, tl)
	s.TaskbarWindow.SetSize(SM_CXSCREEN, config.Taskbar.Height)
	if config.Taskbar.Position == "top" {
		SetPos(s.TaskbarWindow.Handle(), 0, 0)
		SetWorkspace(winc.NewRect(0, config.Taskbar.Height, int(PrimaryMonitor.Rect.Right), int(PrimaryMonitor.Rect.Bottom)))
	} else {
		SetPos(s.TaskbarWindow.Handle(), 0, int(PrimaryMonitor.Rect.Bottom)-config.Taskbar.Height)
		SetWorkspace(winc.NewRect(0, 0, int(PrimaryMonitor.Rect.Right), int(PrimaryMonitor.Rect.Bottom)-config.Taskbar.Height))
	}

	s.TaskbarWindow.OnPaint().Bind(func(arg *winc.Event) {
		if p, ok := arg.Data.(*winc.PaintEventData); ok {
			p.Canvas.DrawFillRect(
				winc.NewRect(0, 0, SM_CXSCREEN, config.Taskbar.Height),
				winc.NewPen(w32.PS_GEOMETRIC, 0, winc.NewSolidColorBrush(winc.RGB(0, 0, 0))),
				winc.NewSolidColorBrush(winc.RGB(byte(config.Taskbar.Bgcolor.R), byte(config.Taskbar.Bgcolor.G), byte(config.Taskbar.Bgcolor.B))),
			)
		}
	})
	s.TaskbarWindow.SetContextMenu(s.TaskbarWindow.ContextMenu())
	list := GetProcesses()
	for i := 0; i < len(list); i++ {
		tl.Add(s.TaskbarWindow, list[i])
	}

	w32.SetShellWindow(s.mainWindow.Handle())
	keyboardHook := SetupHotkeys(s.mainWindow.Handle())
	defer w32.UnhookWindowsHookEx(keyboardHook)
	w32.SetWindowPos(s.mainWindow.Handle(), w32.HWND_BOTTOM, SM_XVIRTUALSCREEN, SM_YVIRTUALSCREEN, SM_CXVIRTUALSCREEN, SM_CYVIRTUALSCREEN, w32.SWP_SHOWWINDOW)
	w32.SetWindowPos(s.TaskbarWindow.Handle(), w32.HWND_TOPMOST, 0, 0, 0, 0, w32.SWP_NOACTIVATE|w32.SWP_NOSIZE|w32.SWP_NOMOVE)

	tl.Refresh(s.TaskbarWindow, false)

	// s.TaskbarWindow.GetTaskbarState()
	winc.RunMainLoop()
}

// func MakeSticky(hWnd w32.HWND) {
// 	// Set magicDWord to make window sticky (same magicDWord that is used by LiteStep)...
// 	w32.SetWindowLongPtr(hWnd, w32.GWLP_USERDATA, 0x49474541) /* magicDWord https://github.com/search?q=0x49474541&type=code */
// }

// func CheckSticky(hWnd w32.HWND) bool {
// 	return (w32.GetWindowLongPtr(uintptr(hWnd), w32.GWLP_USERDATA) == 0x49474541)
// }

func GetWallpaper() (wallpaperPath, backgroundColor string) {
	w, err := registry.OpenKey(registry.CURRENT_USER, `Control Panel\Desktop`, registry.QUERY_VALUE)
	if err != nil {
		log.Println(err)
		return
	}
	defer w.Close()

	// TODO: read out more options: centered, filled, etc
	wallpaperPath, _, err = w.GetStringValue("Wallpaper")
	if err != nil {
		log.Println(err)
		return
	}
	// log.Printf("wallpaperPath is %q\n", wallpaperPath)

	c, err := registry.OpenKey(registry.CURRENT_USER, `Control Panel\Colors`, registry.QUERY_VALUE)
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close()

	backgroundColor, _, err = c.GetStringValue("Background")
	if err != nil {
		log.Println(err)
		return
	}
	// log.Printf("Background is %q\n", backgroundColor)

	return
}
