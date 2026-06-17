//go:build !windows

package ops

import (
	"os"
	"syscall"
)

func isSameDevice(a, b os.FileInfo) bool {
	aStat, ok := a.Sys().(*syscall.Stat_t)
	if !ok {
		return true
	}
	bStat, ok := b.Sys().(*syscall.Stat_t)
	if !ok {
		return true
	}
	return aStat.Dev == bStat.Dev
}
