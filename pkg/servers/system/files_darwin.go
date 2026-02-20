package system

import (
	"os"
	"syscall"
	"time"
)

func init() {
	fileCreatedAt = func(_ string, info os.FileInfo) time.Time {
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok || stat == nil {
			return time.Time{}
		}
		return time.Unix(stat.Birthtimespec.Sec, stat.Birthtimespec.Nsec)
	}
}
