package probe

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestReviewR1ProbeDeadlineWithInheritedOutput(t *testing.T) {
	dir := t.TempDir()
	pidfile := filepath.Join(dir, "child.pid")
	ssh := filepath.Join(dir, "ssh")
	body := "#!/bin/sh\nsleep 120 &\nprintf '%s\\n' \"$!\" > \"$PROBE_CHILD_PID\"\nwait\n"
	if err := os.WriteFile(ssh, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	t.Setenv("PROBE_CHILD_PID", pidfile)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := Probe(ctx, "fake-host")
		done <- err
	}()
	finished := false
	t.Cleanup(func() {
		cancel()
		data, err := os.ReadFile(pidfile)
		if err != nil {
			t.Error(err)
			return
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil || pid <= 1 {
			t.Fatalf("invalid fixture child pid: %q", data)
		}
		_ = syscall.Kill(pid, syscall.SIGKILL)
		if !finished {
			select {
			case err := <-done:
				if err == nil {
					t.Error("cancelled probe unexpectedly succeeded")
				}
			case <-time.After(3 * time.Second):
				t.Error("fixture cleanup did not release Probe")
			}
		}
	})
	select {
	case err := <-done:
		finished = true
		if err == nil {
			t.Error("R1: stalled probe must report down")
		}
	case <-time.After(8*time.Second + 500*time.Millisecond): // contract R1: ≤ 8 s end to end (+ scheduling slack)
		t.Errorf("R1: Probe still blocked after 8.5 s; killed ssh's descendant holds stdout/stderr open")
	}
}
