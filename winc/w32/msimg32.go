package w32

import (
	"syscall"
)

var (
	modmsimg32 = syscall.NewLazyDLL("msimg32.dll")

	procAlphaBlend = modmsimg32.NewProc("AlphaBlend")
)

func AlphaBlend(dcdest HDC, xoriginDest int32, yoriginDest int32, wDest int32, hDest int32, dcsrc HDC, xoriginSrc int32, yoriginSrc int32, wsrc int32, hsrc int32, ftn uintptr) (err error) {
	r1, _, e1 := syscall.SyscallN(
		procAlphaBlend.Addr(),
		uintptr(dcdest),
		uintptr(xoriginDest),
		uintptr(yoriginDest),
		uintptr(wDest),
		uintptr(hDest),
		uintptr(dcsrc),
		uintptr(xoriginSrc),
		uintptr(yoriginSrc),
		uintptr(wsrc),
		uintptr(hsrc),
		ftn)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}
