//go:build !windows

package log

import (
	"golang.org/x/sys/unix"
	"os"
)

func lockLogFile(fp *os.File) error { return unix.Flock(int(fp.Fd()), unix.LOCK_EX|unix.LOCK_NB) }
