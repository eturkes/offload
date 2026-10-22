package e2e_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestReviewV6ProbeFailureIsNotAbsence(t *testing.T) {
	if os.Getenv("OFFLOAD_REVIEW_V6_CHILD") == "1" {
		dir := t.TempDir()
		marker := fmt.Sprintf("review-v6-%d", os.Getpid())
		ready := filepath.Join(dir, "ready")
		child := exec.Command(helperPath, "ignore", marker, ready)
		must(t, child.Start())
		defer func() {
			_ = child.Process.Kill()
			_ = child.Wait()
		}()
		deadline := time.Now().Add(5 * time.Second)
		for !exists(ready) {
			if time.Now().After(deadline) {
				t.Fatal("positive-control helper did not start")
			}
			time.Sleep(time.Millisecond)
		}
		out, err := exec.Command("pgrep", "-f", "ctrhelper ignore "+marker).Output()
		must(t, err)
		if string(out) != strconv.Itoa(child.Process.Pid)+"\n" {
			t.Fatalf("positive control: expected exactly helper pid %d; got %q", child.Process.Pid, out)
		}
		must(t, os.WriteFile(filepath.Join(dir, "pgrep"), []byte("#!/bin/sh\necho 'pgrep: injected read failure' >&2\nexit 2\n"), 0o755))
		t.Setenv("PATH", dir)
		var probeErr *exec.ExitError
		if err := exec.Command("pgrep", "-f", "ctrhelper ignore "+marker).Run(); !errors.As(err, &probeErr) || probeErr.ExitCode() != 2 {
			t.Fatalf("injected probe returned %v, want rc 2", err)
		}
		fmt.Println("probe setup complete: expected visible count=1; injected probe rc=2")
		markerGone(t, marker, time.Second)
		must(t, child.Process.Signal(syscall.Signal(0)))
		fmt.Printf("returned-as-gone: helper pid %d is live; positive count=1; probe rc=2\n", child.Process.Pid)
		return
	}
	child := exec.Command(os.Args[0], "-test.run=^TestReviewV6ProbeFailureIsNotAbsence$", "-test.v")
	child.Env = append(os.Environ(), "OFFLOAD_REVIEW_V6_CHILD=1")
	out, err := child.CombinedOutput()
	if !strings.Contains(string(out), "probe setup complete: expected visible count=1; injected probe rc=2") {
		t.Fatalf("child did not reach the controlled probe: err=%v output=%s", err, out)
	}
	if err == nil || strings.Contains(string(out), "returned-as-gone:") {
		t.Fatalf("absence assertion accepted a failed process probe:\n%s", out)
	}
}
