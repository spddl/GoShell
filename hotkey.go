package main

import (
	"strings"

	"github.com/leaanthony/winc"
	"github.com/leaanthony/winc/w32"
)

func SetupHotkeys(hWnd uintptr) (keyboardHook uintptr) {
	for i, hk := range config.Hotkey {
		keys := strings.ToUpper(hk.Buttons)
		arr := strings.Split(keys, "+")
		var hotkey string
		var fsModifiers uint32
		for _, v := range arr {
			switch v {
			case "WIN":
				fsModifiers |= w32.MOD_WIN

			case "ALT":
				fsModifiers |= w32.MOD_ALT

			case "CTRL":
			case "STRG":
				fsModifiers |= w32.MOD_CONTROL

			case "SHIFT":
				fsModifiers |= w32.MOD_SHIFT
			default:
				hotkey = v
			}
		}
		w32.RegisterHotKey(hWnd, i, fsModifiers, uint32(winc.String2key[hotkey]))
	}
	return
}
