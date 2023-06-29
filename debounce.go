package main

import (
	"sync"
	"time"
)

var msgpool sync.Map
var debounceDelay = time.Duration(time.Millisecond * 300)

func (dlg *TaskbarForm) debounce(hWnd uintptr) {
	if hWnd == 0 {
		return
	}
	if _, ok := msgpool.Load(hWnd); !ok {
		dlg.CheckFullscreen(hWnd)
		msgpool.Store(hWnd, time.AfterFunc(debounceDelay, func() {
			msgpool.Delete(hWnd)
		}))
	} else {
		msgpool.Store(hWnd, time.AfterFunc(debounceDelay, func() {
			msgpool.Delete(hWnd)
			dlg.CheckFullscreen(hWnd)
		}))
	}
}
