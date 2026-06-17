package ops

import (
	"runtime"
	"testing"
)

func skipIfWindows(t *testing.T, reason string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip(reason)
	}
}
