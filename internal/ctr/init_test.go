package ctr

import (
	"io"
	"runtime"
	"syscall"
	"testing"
)

// K8: a refused init (not PID 1) leaves fd 3 alone — it once wrapped fd 3 before the check, and the
// *os.File's finalizer then closed whatever fd 3 was (here: the runtime's epoll fd → fatal netpoll).
func TestK8RefusedInitLeavesFd3Open(t *testing.T) {
	var before syscall.Stat_t
	if err := syscall.Fstat(3, &before); err != nil {
		t.Skipf("fd 3 not open in this process: %v", err)
	}
	if rc := Init([]string{"true"}, io.Discard); rc != 125 {
		t.Fatalf("Init outside a container = %d, want 125", rc)
	}
	for range 3 {
		runtime.GC()
	}
	var after syscall.Stat_t
	if err := syscall.Fstat(3, &after); err != nil || after.Dev != before.Dev || after.Ino != before.Ino {
		t.Fatalf("fd 3 changed after a refused Init + GC: %v (dev/ino %d/%d → %d/%d)", err, before.Dev, before.Ino, after.Dev, after.Ino)
	}
}
