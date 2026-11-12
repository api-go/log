package log

import (
	"golang.org/x/sys/windows"
	"os"
)

func lockLogFile(fp *os.File) error {
	return windows.LockFileEx(windows.Handle(fp.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &windows.Overlapped{})
}
