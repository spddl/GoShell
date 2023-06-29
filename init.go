package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/leaanthony/winc/w32"
)

var (
	PrimaryFromVirtualX int32 = 0
	PrimaryFromVirtualY int32 = 0
	SM_CXSCREEN         int   = w32.GetSystemMetrics(w32.SM_CXSCREEN)
	SM_CYSCREEN         int   = w32.GetSystemMetrics(w32.SM_CYSCREEN)
	SM_XVIRTUALSCREEN   int   = w32.GetSystemMetrics(w32.SM_XVIRTUALSCREEN)
	SM_YVIRTUALSCREEN   int   = w32.GetSystemMetrics(w32.SM_YVIRTUALSCREEN)
	SM_CXVIRTUALSCREEN  int   = w32.GetSystemMetrics(w32.SM_CXVIRTUALSCREEN)
	SM_CYVIRTUALSCREEN  int   = w32.GetSystemMetrics(w32.SM_CYVIRTUALSCREEN)
)

// func init() {
// 	log.Printf("\nSM_CXSCREEN: %d\nSM_CYSCREEN: %d\nSM_XVIRTUALSCREEN: %d\nSM_YVIRTUALSCREEN: %d\nSM_CXVIRTUALSCREEM: %d\nSM_CYVIRTUALSCREEN: %d\n", SM_CXSCREEN, SM_CYSCREEN, SM_XVIRTUALSCREEN, SM_CXVIRTUALSCREEN, SM_CYVIRTUALSCREEN, SM_CYVIRTUALSCREEN)
// }

func createRegFiles() {
	def := []byte(`Windows Registry Editor Version 5.00

[HKEY_CURRENT_USER\Software\Microsoft\Windows NT\CurrentVersion\Winlogon]
"Shell"=-`)
	if err := os.WriteFile(filepath.Join(exPath, "Set_Windows_Default_ExplorerShell.reg"), def, 0o644); err != nil {
		log.Println(err)
	}
	path := filepath.Join(exPath, "GoShell.exe")
	path = strings.ReplaceAll(path, `\`, `\\`)
	path = strings.ReplaceAll(path, `"`, `\"`)
	goshell := []byte(fmt.Sprintf(`Windows Registry Editor Version 5.00

[HKEY_CURRENT_USER\Software\Microsoft\Windows NT\CurrentVersion\Winlogon]
"Shell"="%s -nofiles -startup"`, path))

	if err := os.WriteFile(filepath.Join(exPath, "Set_GoShell.reg"), goshell, 0o644); err != nil {
		log.Println(err)
	}
}
