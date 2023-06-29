/*
 * Copyright (C) 2019 Tad Vizbaras. All Rights Reserved.
 * Copyright (C) 2010-2012 The W32 Authors. All Rights Reserved.
 */

package w32

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

type CSIDL uint32

const (
	CSIDL_DESKTOP                 = 0x00
	CSIDL_INTERNET                = 0x01
	CSIDL_PROGRAMS                = 0x02
	CSIDL_CONTROLS                = 0x03
	CSIDL_PRINTERS                = 0x04
	CSIDL_PERSONAL                = 0x05
	CSIDL_FAVORITES               = 0x06
	CSIDL_STARTUP                 = 0x07
	CSIDL_RECENT                  = 0x08
	CSIDL_SENDTO                  = 0x09
	CSIDL_BITBUCKET               = 0x0A
	CSIDL_STARTMENU               = 0x0B
	CSIDL_MYDOCUMENTS             = 0x0C
	CSIDL_MYMUSIC                 = 0x0D
	CSIDL_MYVIDEO                 = 0x0E
	CSIDL_DESKTOPDIRECTORY        = 0x10
	CSIDL_DRIVES                  = 0x11
	CSIDL_NETWORK                 = 0x12
	CSIDL_NETHOOD                 = 0x13
	CSIDL_FONTS                   = 0x14
	CSIDL_TEMPLATES               = 0x15
	CSIDL_COMMON_STARTMENU        = 0x16
	CSIDL_COMMON_PROGRAMS         = 0x17
	CSIDL_COMMON_STARTUP          = 0x18
	CSIDL_COMMON_DESKTOPDIRECTORY = 0x19
	CSIDL_APPDATA                 = 0x1A
	CSIDL_PRINTHOOD               = 0x1B
	CSIDL_LOCAL_APPDATA           = 0x1C
	CSIDL_ALTSTARTUP              = 0x1D
	CSIDL_COMMON_ALTSTARTUP       = 0x1E
	CSIDL_COMMON_FAVORITES        = 0x1F
	CSIDL_INTERNET_CACHE          = 0x20
	CSIDL_COOKIES                 = 0x21
	CSIDL_HISTORY                 = 0x22
	CSIDL_COMMON_APPDATA          = 0x23
	CSIDL_WINDOWS                 = 0x24
	CSIDL_SYSTEM                  = 0x25
	CSIDL_PROGRAM_FILES           = 0x26
	CSIDL_MYPICTURES              = 0x27
	CSIDL_PROFILE                 = 0x28
	CSIDL_SYSTEMX86               = 0x29
	CSIDL_PROGRAM_FILESX86        = 0x2A
	CSIDL_PROGRAM_FILES_COMMON    = 0x2B
	CSIDL_PROGRAM_FILES_COMMONX86 = 0x2C
	CSIDL_COMMON_TEMPLATES        = 0x2D
	CSIDL_COMMON_DOCUMENTS        = 0x2E
	CSIDL_COMMON_ADMINTOOLS       = 0x2F
	CSIDL_ADMINTOOLS              = 0x30
	CSIDL_CONNECTIONS             = 0x31
	CSIDL_COMMON_MUSIC            = 0x35
	CSIDL_COMMON_PICTURES         = 0x36
	CSIDL_COMMON_VIDEO            = 0x37
	CSIDL_RESOURCES               = 0x38
	CSIDL_RESOURCES_LOCALIZED     = 0x39
	CSIDL_COMMON_OEM_LINKS        = 0x3A
	CSIDL_CDBURN_AREA             = 0x3B
	CSIDL_COMPUTERSNEARME         = 0x3D
	CSIDL_FLAG_CREATE             = 0x8000
	CSIDL_FLAG_DONT_VERIFY        = 0x4000
	CSIDL_FLAG_NO_ALIAS           = 0x1000
	CSIDL_FLAG_PER_USER_INIT      = 0x8000
	CSIDL_FLAG_MASK               = 0xFF00
)

var (
	modshell32 = syscall.NewLazyDLL("shell32.dll")

	procSHBrowseForFolder    = modshell32.NewProc("SHBrowseForFolderW")
	procSHGetPathFromIDList  = modshell32.NewProc("SHGetPathFromIDListW")
	procDragAcceptFiles      = modshell32.NewProc("DragAcceptFiles")
	procDragQueryFile        = modshell32.NewProc("DragQueryFileW")
	procDragQueryPoint       = modshell32.NewProc("DragQueryPoint")
	procDragFinish           = modshell32.NewProc("DragFinish")
	procShellExecute         = modshell32.NewProc("ShellExecuteW")
	procExtractIcon          = modshell32.NewProc("ExtractIconW")
	procGetSpecialFolderPath = modshell32.NewProc("SHGetSpecialFolderPathW")
	procShDefExtractIcon     = modshell32.NewProc("SHDefExtractIconW")
	procShAppBarMessage      = modshell32.NewProc("SHAppBarMessage")
	procShRestricted         = modshell32.NewProc("SHRestricted")
	procShGetDesktopFolder   = modshell32.NewProc("SHGetDesktopFolder")
	procILFindLastID         = modshell32.NewProc("ILFindLastID")
	procILClone              = modshell32.NewProc("ILClone")
	procILRemoveLastID       = modshell32.NewProc("ILRemoveLastID")
	procGetUIObjectOf        = modshell32.NewProc("GetUIObjectOf")
	procShParseDisplayName   = modshell32.NewProc("SHParseDisplayName")
	procShBindToParent       = modshell32.NewProc("SHBindToParent")
	shGetFileInfo            = modshell32.NewProc("SHGetFileInfoW")
)

func SHBrowseForFolder(bi *BROWSEINFO) uintptr {
	ret, _, _ := procSHBrowseForFolder.Call(uintptr(unsafe.Pointer(bi)))

	return ret
}

func SHGetPathFromIDList(idl uintptr) string {
	buf := make([]uint16, 1024)
	procSHGetPathFromIDList.Call(
		idl,
		uintptr(unsafe.Pointer(&buf[0])))

	return syscall.UTF16ToString(buf)
}

func DragAcceptFiles(hwnd HWND, accept bool) {
	procDragAcceptFiles.Call(
		uintptr(hwnd),
		uintptr(BoolToBOOL(accept)))
}

func DragQueryFile(hDrop HDROP, iFile uint) (fileName string, fileCount uint) {
	ret, _, _ := procDragQueryFile.Call(
		uintptr(hDrop),
		uintptr(iFile),
		0,
		0)

	fileCount = uint(ret)

	if iFile != 0xFFFFFFFF {
		buf := make([]uint16, fileCount+1)

		ret, _, _ := procDragQueryFile.Call(
			uintptr(hDrop),
			uintptr(iFile),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(fileCount+1))

		if ret == 0 {
			panic("Invoke DragQueryFile error.")
		}

		fileName = syscall.UTF16ToString(buf)
	}

	return
}

func DragQueryPoint(hDrop HDROP) (x, y int, isClientArea bool) {
	var pt POINT
	ret, _, _ := procDragQueryPoint.Call(
		uintptr(hDrop),
		uintptr(unsafe.Pointer(&pt)))

	return int(pt.X), int(pt.Y), (ret == 1)
}

func DragFinish(hDrop HDROP) {
	procDragFinish.Call(uintptr(hDrop))
}

func ShellExecute(hwnd HWND, lpOperation, lpFile, lpParameters, lpDirectory string, nShowCmd int) error {
	var op, param, directory uintptr
	if len(lpOperation) != 0 {
		op = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpOperation)))
	}
	if len(lpParameters) != 0 {
		param = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpParameters)))
	}
	if len(lpDirectory) != 0 {
		directory = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpDirectory)))
	}

	ret, _, _ := procShellExecute.Call(
		uintptr(hwnd),
		op,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpFile))),
		param,
		directory,
		uintptr(nShowCmd))

	errorMsg := ""
	if ret != 0 && ret <= 32 {
		switch int(ret) {
		case ERROR_FILE_NOT_FOUND:
			errorMsg = "The specified file was not found."
		case ERROR_PATH_NOT_FOUND:
			errorMsg = "The specified path was not found."
		case ERROR_BAD_FORMAT:
			errorMsg = "The .exe file is invalid (non-Win32 .exe or error in .exe image)."
		case SE_ERR_ACCESSDENIED:
			errorMsg = "The operating system denied access to the specified file."
		case SE_ERR_ASSOCINCOMPLETE:
			errorMsg = "The file name association is incomplete or invalid."
		case SE_ERR_DDEBUSY:
			errorMsg = "The DDE transaction could not be completed because other DDE transactions were being processed."
		case SE_ERR_DDEFAIL:
			errorMsg = "The DDE transaction failed."
		case SE_ERR_DDETIMEOUT:
			errorMsg = "The DDE transaction could not be completed because the request timed out."
		case SE_ERR_DLLNOTFOUND:
			errorMsg = "The specified DLL was not found."
		case SE_ERR_NOASSOC:
			errorMsg = "There is no application associated with the given file name extension. This error will also be returned if you attempt to print a file that is not printable."
		case SE_ERR_OOM:
			errorMsg = "There was not enough memory to complete the operation."
		case SE_ERR_SHARE:
			errorMsg = "A sharing violation occurred."
		default:
			errorMsg = fmt.Sprintf("Unknown error occurred with error code %v", ret)
		}
	} else {
		return nil
	}

	return errors.New(errorMsg)
}

func ExtractIcon(lpszExeFileName string, nIconIndex int) HICON {
	ret, _, _ := procExtractIcon.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpszExeFileName))),
		uintptr(nIconIndex))

	return HICON(ret)
}

func SHGetSpecialFolderPath(hwndOwner HWND, lpszPath *uint16, csidl CSIDL, fCreate bool) bool {
	ret, _, _ := procGetSpecialFolderPath.Call(
		uintptr(hwndOwner),
		uintptr(unsafe.Pointer(lpszPath)),
		uintptr(csidl),
		uintptr(BoolToBOOL(fCreate)),
		0,
		0)

	return ret != 0
}

func SHDefExtractIcon(pszIconFile *uint16, iIndex int32, uFlags uint32, phiconLarge, phiconSmall *HICON, nIconSize uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(procShDefExtractIcon.Addr(),
		uintptr(unsafe.Pointer(pszIconFile)),
		uintptr(iIndex),
		uintptr(uFlags),
		uintptr(unsafe.Pointer(phiconLarge)),
		uintptr(unsafe.Pointer(phiconSmall)),
		uintptr(nIconSize))

	return HRESULT(ret)
}

func SHAppBarMessage(dwMessage uint32, pData *APPBARDATA) uintptr {
	ret, _, _ := syscall.SyscallN(procShAppBarMessage.Addr(), uintptr(dwMessage), uintptr(unsafe.Pointer(pData)))
	return ret
}

func SHRestricted(policy RESTRICTIONS) DWORD {
	ret1, _, _ := syscall.SyscallN(procShRestricted.Addr(), uintptr(policy))
	return DWORD(ret1)
}

// func SHGetDesktopFolder(psf **IShellFolder) HRESULT {
// 	ret1, _, _ := syscall.SyscallN(procShGetDesktopFolder.Addr(),
// 		uintptr(unsafe.Pointer(psf)))
// 	return HRESULT(ret1)
// }

// func (i *IContextMenu) QueryContextMenu(hMenu syscall.Handle, indexMenu uint32, idCmdFirst uint32, idCmdLast uint32, uFlags uint32) int32 {
// 	ret, _, _ := syscall.Syscall6(i.LpVtbl.QueryContextMenu, 6,
// 		uintptr(unsafe.Pointer(i)),
// 		uintptr(hMenu),
// 		uintptr(indexMenu),
// 		uintptr(idCmdFirst),
// 		uintptr(idCmdLast),
// 		uintptr(uFlags))

// 	return int32(ret)
// }

// func (i *IContextMenu) InvokeCommand(pici uintptr) {
// 	syscall.Syscall(i.LpVtbl.InvokeCommand, 2,
// 		uintptr(unsafe.Pointer(i)),
// 		pici,
// 		0)
// }

// func (i *IContextMenu) GetCommandString(idCmd uint32, uType uint32, pReserved uintptr, pszName *uint16, cchMax uint32) int32 {
// 	ret, _, _ := syscall.Syscall6(i.LpVtbl.GetCommandString, 6,
// 		uintptr(unsafe.Pointer(i)),
// 		uintptr(idCmd),
// 		uintptr(uType),
// 		pReserved,
// 		uintptr(unsafe.Pointer(pszName)),
// 		uintptr(cchMax))

// 	return int32(ret)
// }

// func (i *IContextMenu2) QueryContextMenu(hMenu syscall.Handle, indexMenu uint32, idCmdFirst uint32, idCmdLast uint32, uFlags uint32) int32 {
// 	ret, _, _ := syscall.SyscallN(i.LpVtbl.QueryContextMenu,
// 		uintptr(unsafe.Pointer(i)),
// 		uintptr(hMenu),
// 		uintptr(indexMenu),
// 		uintptr(idCmdFirst),
// 		uintptr(idCmdLast),
// 		uintptr(uFlags))

// 	return int32(ret)
// }

// func (i *IContextMenu2) InvokeCommand(pici uintptr) {
// 	syscall.SyscallN(i.LpVtbl.InvokeCommand,
// 		uintptr(unsafe.Pointer(i)),
// 		pici)
// }

// func (i *IContextMenu2) GetCommandString(idCmd uint32, uType uint32, pReserved uintptr, pszName *uint16, cchMax uint32) int32 {
// 	ret, _, _ := syscall.SyscallN(i.LpVtbl.GetCommandString,
// 		uintptr(unsafe.Pointer(i)),
// 		uintptr(idCmd),
// 		uintptr(uType),
// 		pReserved,
// 		uintptr(unsafe.Pointer(pszName)),
// 		uintptr(cchMax))

// 	return int32(ret)
// }

// func (i *IContextMenu3) QueryContextMenu(hMenu syscall.Handle, indexMenu uint32, idCmdFirst uint32, idCmdLast uint32, uFlags uint32) int32 {
// 	ret, _, _ := syscall.Syscall6(i.LpVtbl.QueryContextMenu, 6,
// 		uintptr(unsafe.Pointer(i)),
// 		uintptr(hMenu),
// 		uintptr(indexMenu),
// 		uintptr(idCmdFirst),
// 		uintptr(idCmdLast),
// 		uintptr(uFlags))

// 	return int32(ret)
// }

// func (i *IContextMenu3) InvokeCommand(pici uintptr) {
// 	syscall.Syscall(i.LpVtbl.InvokeCommand, 2,
// 		uintptr(unsafe.Pointer(i)),
// 		pici,
// 		0)
// }

// func (i *IContextMenu3) GetCommandString(idCmd uint32, uType uint32, pReserved uintptr, pszName *uint16, cchMax uint32) int32 {
// 	ret, _, _ := syscall.Syscall6(i.LpVtbl.GetCommandString, 6,
// 		uintptr(unsafe.Pointer(i)),
// 		uintptr(idCmd),
// 		uintptr(uType),
// 		pReserved,
// 		uintptr(unsafe.Pointer(pszName)),
// 		uintptr(cchMax))

// 	return int32(ret)
// }

// func (i *IContextMenu3) HandleMenuMsg2(uMsg uint32, wParam uintptr, lParam uintptr, plResult *int32) int32 {
// 	ret, _, _ := syscall.SyscallN(i.LpVtbl.HandleMenuMsg2,
// 		uintptr(unsafe.Pointer(i)),
// 		uintptr(uMsg),
// 		wParam,
// 		lParam,
// 		uintptr(unsafe.Pointer(plResult)))

// 	return int32(ret)
// }

// func (sf *IShellFolder) QueryInterface(iid *ole.GUID, ppvObj interface{}) uintptr {
// 	ret, _, _ := syscall.SyscallN(
// 		sf.lpVtbl.QueryInterface,
// 		uintptr(unsafe.Pointer(sf)),
// 		uintptr(unsafe.Pointer(iid)),
// 		uintptr(unsafe.Pointer(&ppvObj)),
// 	)
// 	return ret
// }

// func (sf *IShellFolder) AddRef() uintptr {
// 	ret, _, _ := syscall.SyscallN(
// 		sf.lpVtbl.AddRef,
// 		uintptr(unsafe.Pointer(sf)),
// 	)
// 	return ret
// }

// func (sf *IShellFolder) Release() uintptr {
// 	ret, _, _ := syscall.SyscallN(
// 		sf.lpVtbl.Release,
// 		uintptr(unsafe.Pointer(sf)),
// 	)
// 	return ret
// }

// // https://learn.microsoft.com/en-us/windows/win32/api/shobjidl_core/nf-shobjidl_core-ishellfolder-parsedisplayname
// func (sf *IShellFolder) ParseDisplayName(hwndOwner uintptr, pbcReserved uintptr, pszDisplayName string, pchEaten *uint32, ppidl *uintptr, pdwAttributes *uint32) uintptr {
// 	ret, _, _ := syscall.SyscallN(
// 		sf.lpVtbl.ParseDisplayName,
// 		uintptr(unsafe.Pointer(sf)),
// 		hwndOwner,
// 		pbcReserved,
// 		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(pszDisplayName))),
// 		uintptr(unsafe.Pointer(pchEaten)),
// 		uintptr(unsafe.Pointer(ppidl)),
// 		uintptr(unsafe.Pointer(pdwAttributes)),
// 	)
// 	return ret
// }

// func (sf *IShellFolder) BindToObject(pidl uintptr, pbcReserved uintptr, riid *ole.GUID, ppvOut interface{}) uintptr {
// 	ret, _, _ := syscall.SyscallN(
// 		sf.lpVtbl.BindToObject,
// 		uintptr(unsafe.Pointer(sf)),
// 		pidl,
// 		pbcReserved,
// 		uintptr(unsafe.Pointer(riid)),
// 		uintptr(unsafe.Pointer(&ppvOut)),
// 	)
// 	return ret
// }

// func ILFindLastID(pidl uintptr) uintptr {
// 	ret, _, _ := procILFindLastID.Call(pidl)
// 	return ret
// }

// func ILClone(pidl uintptr) uintptr {
// 	ret, _, _ := procILClone.Call(pidl)
// 	return ret
// }

// func ILRemoveLastID(pidl uintptr) bool {
// 	ret, _, _ := procILRemoveLastID.Call(pidl)
// 	return ret != 0
// }

// func GetUIObjectOf(shellFolder *IShellFolder, hwnd uintptr, pidl uintptr, riid *ole.GUID) (uintptr, error) {
// 	var ppvObj uintptr
// 	ret, _, _ := procGetUIObjectOf.Call(
// 		uintptr(unsafe.Pointer(shellFolder)),
// 		hwnd,
// 		pidl,
// 		uintptr(unsafe.Pointer(riid)),
// 		uintptr(unsafe.Pointer(&ppvObj)),
// 	)
// 	if ret != 0 {
// 		return ppvObj, ole.NewError(ret)
// 	}
// 	return ppvObj, nil
// }

// // https://learn.microsoft.com/en-us/windows/win32/api/shlobj_core/nf-shlobj_core-shparsedisplayname
// func SHParseDisplayName(pszName *uint16, pbc uintptr, ppidl *ITEMIDLIST, sfgaoIn uint32, psfgaoOut *SFGAOF) HRESULT {
// 	ret, _, _ := syscall.SyscallN(procShParseDisplayName.Addr(),
// 		uintptr(unsafe.Pointer(pszName)),
// 		pbc,
// 		uintptr(unsafe.Pointer(ppidl)),
// 		0,
// 		uintptr(unsafe.Pointer(psfgaoOut)))

// 	return HRESULT(ret)
// }

// // https://learn.microsoft.com/en-us/windows/win32/api/shlobj_core/nf-shlobj_core-shbindtoparent
// func SHBindToParent(pidl /*const*/ LPCITEMIDLIST, riid REFIID, ppv *uintptr, ppidlLast *LPCITEMIDLIST) HRESULT {
// 	ret, _, _ := syscall.SyscallN(procShBindToParent.Addr(),
// 		uintptr(unsafe.Pointer(pidl)),
// 		uintptr(unsafe.Pointer(riid)),
// 		uintptr(unsafe.Pointer(ppv)),
// 		uintptr(unsafe.Pointer(ppidlLast)))
// 	return HRESULT(ret)
// }

func SHGetFileInfo(pszPath *uint16, dwFileAttributes uint32, psfi *SHFILEINFO, cbFileInfo, uFlags uint32) uintptr {
	ret, _, _ := syscall.SyscallN(shGetFileInfo.Addr(),
		uintptr(unsafe.Pointer(pszPath)),
		uintptr(dwFileAttributes),
		uintptr(unsafe.Pointer(psfi)),
		uintptr(cbFileInfo),
		uintptr(uFlags))

	return ret
}
