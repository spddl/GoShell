package main

import (
	"log"
	"sync"

	"golang.org/x/sys/windows"
)

// https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-setpriorityclass
type powersaver struct {
	mu      sync.RWMutex
	enabled bool
}

var Powersaver powersaver

func (ps *powersaver) Try(states bool) {
	if states {
		// Begin background processing mode. The system lowers the resource scheduling priorities of the process (and its threads) so that it can perform background work without significantly affecting activity in the foreground.
		// This value can be specified only if hProcess is a handle to the current process. The function fails if the process is already in background processing mode.

		// Windows Server 2003 and Windows XP:  This value is not supported.
		if !ps.Get() {
			if err := windows.SetPriorityClass(windows.CurrentProcess(), windows.PROCESS_MODE_BACKGROUND_BEGIN); err != nil {
				log.Println(err)
			}
			ps.Set(true)
		}
	} else {
		// End background processing mode. The system restores the resource scheduling priorities of the process (and its threads) as they were before the process entered background processing mode.
		// This value can be specified only if hProcess is a handle to the current process. The function fails if the process is not in background processing mode.

		// Windows Server 2003 and Windows XP:  This value is not supported.
		if ps.Get() {
			if err := windows.SetPriorityClass(windows.CurrentProcess(), windows.PROCESS_MODE_BACKGROUND_END); err != nil {
				log.Println(err)
			}
			ps.Set(false)
		}
	}
}

func (ps *powersaver) Get() bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.enabled
}

func (ps *powersaver) Set(value bool) {
	ps.mu.Lock()
	ps.enabled = value
	ps.mu.Unlock()
}
