package system

import (
	"os"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

func init() {
	var (
		mu          sync.RWMutex
		skipDevices = make(map[uint64]struct{})
	)

	fileCreatedAt = func(relPath string, info os.FileInfo) time.Time {
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok || stat == nil {
			return time.Time{}
		}

		mu.RLock()
		_, skip := skipDevices[stat.Dev]
		mu.RUnlock()
		if skip {
			// We've determined the device this file belongs to doesn't support birth time.
			// Bail out early to avoid an extra syscall.
			return time.Time{}
		}

		// Make a statx syscall with STATX_BTIME to request the birth time.
		var statx unix.Statx_t
		if err := unix.Statx(
			unix.AT_FDCWD,
			relPath,
			unix.AT_STATX_SYNC_AS_STAT,
			unix.STATX_BTIME,
			&statx,
		); err != nil {
			return time.Time{}
		}

		// Determine if the device actually supports birth time from the statx response.
		sec, nsec := statx.Btime.Sec, int64(statx.Btime.Nsec)
		if statx.Mask&unix.STATX_BTIME == 0 || (sec == 0 && nsec == 0) {
			// Assume the device doesn't support birth time and skip future statx calls for all other files on the device.
			mu.Lock()
			skipDevices[stat.Dev] = struct{}{}
			mu.Unlock()
			return time.Time{}
		}

		return time.Unix(sec, nsec)
	}
}
