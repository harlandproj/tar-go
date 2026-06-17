//go:build windows

package ops

import "os"

func isSameDevice(a, b os.FileInfo) bool {
	return true
}
