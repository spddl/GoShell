package main

import (
	"github.com/leaanthony/winc"
	"github.com/leaanthony/winc/w32"
)

type TaskItem struct {
	winc.Button
	Icon   uintptr
	Bitmap *winc.Bitmap
	hWnd   uintptr
}

func NewTaskItem(parent winc.Controller) *TaskItem {
	pb := new(TaskItem)

	winc.RegClassOnlyOnce("TaskItem")
	pb.InitControl("TaskItem", parent, w32.WS_EX_NOACTIVATE, w32.WS_CHILD|w32.WS_CLIPSIBLINGS|w32.WS_VISIBLE|w32.WS_TABSTOP)

	winc.RegMsgHandler(pb)

	pb.SetFont(winc.DefaultFont)
	pb.SetText("TaskItem")
	pb.SetSize(config.Taskbar.Button.Size.Width, config.Taskbar.Button.Size.Height)

	return pb
}

// SetIcon sets icon on the button. Recommended icons are 32x32 with 32bit color depth.
func (bt *TaskItem) SetIcon(ico *winc.Icon) {
	w32.SendMessage(bt.Handle(), w32.BM_SETIMAGE, w32.IMAGE_ICON, ico.Handle())
}

func (bt *TaskItem) SetBitmap(bmp *winc.Bitmap) {
	w32.SendMessage(bt.Handle(), w32.BM_SETIMAGE, w32.IMAGE_BITMAP, bmp.GetHBITMAP())
}

func (bt *TaskItem) WndProc(msg uint32, wparam, lparam uintptr) uintptr {
	return w32.DefWindowProc(bt.Handle(), msg, wparam, lparam)
}
