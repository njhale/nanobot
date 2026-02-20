package system

import (
	"os"
	"syscall"
	"time"
)

func init() {
	fileCreatedAt = func(_ string, info os.FileInfo) time.Time {
		data, ok := info.Sys().(*syscall.Win32FileAttributeData)
		if !ok || data == nil {
			return time.Time{}
		}

		return time.Unix(0, data.CreationTime.Nanoseconds()).UTC()
	}
}
