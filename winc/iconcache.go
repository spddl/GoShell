package winc

import (
	"fmt"
	"log"
	"syscall"
	"unsafe"

	"github.com/leaanthony/winc/w32"
)

var iconCache *IconCache

type IconCache struct {
	Bitmap map[string]*Bitmap
	Icon   map[string]uintptr
}

func init() {
	iconCache = NewIconCache()
}

func NewIconCache() *IconCache {
	return &IconCache{
		Bitmap: make(map[string]*Bitmap),
		Icon:   make(map[string]uintptr),
	}
}

// func (ic *IconCache) Clear() {
// 	for key, bmp := range ic.imageAndDPI2Bitmap {
// 		bmp.Dispose()
// 		delete(ic.imageAndDPI2Bitmap, key)
// 	}
// 	for key, ico := range ic.imageAndDPI2Icon {
// 		ico.Destroy()
// 		delete(ic.imageAndDPI2Icon, key)
// 	}
// }

func GetBitmap(filename string, index int) *Bitmap {
	// log.Println("GetBitmap", filename, index)
	path := fmt.Sprintf("%s%d", filename, index)
	val, ok := iconCache.Bitmap[path]
	if ok {
		return val
	}
	bmp := ExtractIconToBitmap(filename, index)
	iconCache.Bitmap[path] = bmp
	return bmp
}

func GetBitmapFromIcon(hIcon uintptr, size w32.Size, dpi int) *Bitmap {
	// log.Println("GetBitmapFromIcon", hIcon, size, dpi)

	bmp, err := NewBitmapFromIconForDPI(NewIcon(hIcon), size, dpi)
	if err != nil {
		log.Println(err)
	}
	return bmp
}

func GetIcon(filePath string) uintptr {
	// log.Println("GetIcon", filePath)

	val, ok := iconCache.Icon[filePath]
	if ok {
		return val
	}
	hIcon := hIconForFilePath(filePath)
	iconCache.Icon[filePath] = hIcon

	return hIcon
}

func hIconForFilePath(filePath string) w32.HICON {
	fPptr, _ := syscall.UTF16PtrFromString(filePath)
	// https://learn.microsoft.com/en-us/windows/win32/api/shellapi/ns-shellapi-shfileinfow
	var shfi w32.SHFILEINFO
	hIml := w32.HIMAGELIST(w32.SHGetFileInfo( // https://docs.microsoft.com/en-us/windows/win32/api/shellapi/nf-shellapi-shgetfileinfow
		fPptr,
		w32.FILE_ATTRIBUTE_NORMAL,
		&shfi,
		uint32(unsafe.Sizeof(shfi)),
		// w32.SHGFI_ICON|w32.SHGFI_SMALLICON,
		w32.SHGFI_USEFILEATTRIBUTES|w32.SHGFI_ICON|w32.SHGFI_SMALLICON|w32.SHGFI_DISPLAYNAME,
	))
	if hIml != 0 {
		// log.Printf("%s %s - %#v %#v\n", filePath, syscall.UTF16ToString(shfi.SzDisplayName[:]), shfi.HIcon, shfi.IIcon)
		return shfi.HIcon
	}
	return 0
}

// func hIconForFilePath(filePath string) (string, w32.HICON) {
// 	fPptr, _ := syscall.UTF16PtrFromString(filePath)

// 	var shfi w32.SHFILEINFO
// 	hIml := w32.HIMAGELIST(win.SHGetFileInfo( // https://docs.microsoft.com/en-us/windows/win32/api/shellapi/nf-shellapi-shgetfileinfow
// 		fPptr,
// 		w32.FILE_ATTRIBUTE_NORMAL,
// 		// 0x80, // FILE_ATTRIBUTE_NORMAL
// 		&shfi,
// 		uint32(unsafe.Sizeof(shfi)),
// 		win.SHGFI_USEFILEATTRIBUTES|win.SHGFI_ICON|win.SHGFI_SMALLICON|win.SHGFI_DISPLAYNAME))
// 	if hIml != 0 {
// 		return syscall.UTF16ToString(shfi.SzDisplayName[:]), shfi.HIcon
// 	}
// 	return "", 0
// }
