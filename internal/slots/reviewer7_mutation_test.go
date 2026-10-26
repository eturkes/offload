package slots

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// R2: a live FIFO writer does not make a pipe into a slot owner record or stall status.
func TestR7HoldersIgnoreFIFOWithConnectedWriter(t *testing.T) {
	if dir := os.Getenv("OFFLOAD_R7_OWNER_FIFO"); dir != "" {
		got := (Dir{Path: dir, Total: 1}).Holders()
		if len(got) != 1 || got[0] != "slot 1: " {
			t.Fatalf("FIFO contents became holder text: %q", got)
		}
		return
	}
	dir := t.TempDir()
	slot, err := os.OpenFile(filepath.Join(dir, "slot1"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = slot.Close() }()
	if err = syscall.Flock(int(slot.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	owner := filepath.Join(dir, "slot1.owner")
	if err = syscall.Mkfifo(owner, 0o600); err != nil {
		t.Fatal(err)
	}
	writer, err := os.OpenFile(owner, os.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = writer.Close() }()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(self, "-test.run=^TestR7HoldersIgnoreFIFOWithConnectedWriter$")
	cmd.Env = append(os.Environ(), "OFFLOAD_R7_OWNER_FIFO="+dir)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	defer func() { _ = cmd.Process.Kill() }()
	deadline := time.After(20 * time.Second)
	for {
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("status reader: %v\n%s", err, &out)
			}
			return
		case <-deadline:
			_ = cmd.Process.Kill()
			if err := <-done; err != nil {
				t.Log(err)
			}
			t.Fatal("status reader never completed")
		default:
		}
		fds, _ := os.ReadDir(fmt.Sprintf("/proc/%d/fd", cmd.Process.Pid))
		for _, fd := range fds {
			link, _ := os.Readlink(fmt.Sprintf("/proc/%d/fd/%s", cmd.Process.Pid, fd.Name()))
			if link == owner {
				// The reader is open: send data, then EOF. The correct reader has already
				// refused the nonregular file; the mutant returns these bytes as owner text.
				_, _ = writer.WriteString("not-an-owner-line\n")
				_ = writer.Close()
				break
			}
		}
		time.Sleep(time.Millisecond)
	}
}
