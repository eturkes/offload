package run

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/eturkes/offload/internal/remote"
)

// C3 + J2: installed bash/ssh can still fail pipe/start under a finite descriptor limit.
// Each limit applies to a self-exec child only; no system-wide resource is exhausted.
func TestR7LaunchErrorsAreReported(t *testing.T) {
	if mode := os.Getenv("OFFLOAD_R7_FD_LIMIT"); mode != "" {
		r7LaunchWithNoDescriptors(t, mode)
		return
	}
	t.Parallel()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"pipe", "start", "remote-pipe"} {
		t.Run(mode, func(t *testing.T) {
			cmd := exec.Command(self, "-test.run=^TestR7LaunchErrorsAreReported$")
			cmd.Env = append(os.Environ(), "OFFLOAD_R7_FD_LIMIT="+mode)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("%s child: %v\n%s", mode, err, out)
			}
		})
	}
}

// X3: cancel closes the control pipe even while the transport applies backpressure to preamble.
func TestR7CancellationClosesPipeDuringBlockedPreamble(t *testing.T) {
	if ready := os.Getenv("OFFLOAD_R7_HUP_READY"); ready != "" {
		fd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = syscall.Close(fd) }()
		if err := syscall.EpollCtl(fd, syscall.EPOLL_CTL_ADD, 0, &syscall.EpollEvent{Events: syscall.EPOLLHUP, Fd: 0}); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ready, []byte(strconv.Itoa(os.Getpid())), 0o600); err != nil {
			t.Fatal(err)
		}
		for {
			events := make([]syscall.EpollEvent, 1)
			n, err := syscall.EpollWait(fd, events, -1)
			if errors.Is(err, syscall.EINTR) {
				continue
			}
			if err != nil || n != 1 || events[0].Events&syscall.EPOLLHUP == 0 {
				t.Fatalf("pipe HUP: n=%d events=%v err=%v", n, events, err)
			}
			return
		}
	}
	t.Parallel()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ready := filepath.Join(t.TempDir(), "ready")
	safety, stop := context.WithTimeout(context.Background(), 60*time.Second)
	defer stop()
	cmd := exec.CommandContext(safety, self, "-test.run=^TestR7CancellationClosesPipeDuringBlockedPreamble$")
	cmd.Env = append(os.Environ(), "OFFLOAD_R7_HUP_READY="+ready)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := supervise(ctx, cmd, bytes.Repeat([]byte("x"), 1<<20)); done <- err }()
	deadline := time.Now().Add(20 * time.Second)
	var pid int
	for {
		if data, err := os.ReadFile(ready); err == nil {
			pid, err = strconv.Atoi(string(data))
			if err == nil && pid > 1 {
				break
			}
		}
		if time.Now().After(deadline) {
			cancel()
			t.Fatal("HUP peer did not become ready")
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(20 * time.Second):
		_ = syscall.Kill(pid, syscall.SIGKILL)
		if err := <-done; err != nil {
			t.Log(err)
		}
		t.Fatal("cancellation never closed stdin: the preamble write blocked before the closer started")
	}
}

func r7LaunchWithNoDescriptors(t *testing.T, mode string) {
	t.Helper()
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	if err = os.WriteFile(filepath.Join(bin, "ssh"), []byte("#!/bin/sh\nexit 97\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+":"+os.Getenv("PATH"))
	cmd := remote.Command(context.Background(), bash, "-c", "exit 0")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	var old syscall.Rlimit
	if err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &old); err != nil {
		t.Fatal(err)
	}
	limited := old
	limited.Cur = 64
	if err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limited); err != nil {
		t.Fatal(err)
	}
	var fds []int
	defer func() {
		_ = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &old)
		for _, fd := range fds {
			_ = syscall.Close(fd)
		}
	}()
	for {
		fd, openErr := syscall.Open("/dev/null", syscall.O_RDONLY|syscall.O_CLOEXEC, 0)
		if errors.Is(openErr, syscall.EMFILE) {
			break
		}
		if openErr != nil {
			t.Fatal(openErr)
		}
		fds = append(fds, fd)
	}
	if mode == "start" {
		// StdinPipe gets two descriptors; fork/exec's error pipe then cannot get two.
		for range 3 {
			last := len(fds) - 1
			_ = syscall.Close(fds[last])
			fds = fds[:last]
		}
	}
	var rc int
	if mode == "remote-pipe" {
		r := Remote{Target: remote.Target{Address: "unused"}, Workdir: "offload/w", Meta: "offload/.meta/w", Sub: ".", Mode: "nosync", Argv: []string{"true"}}
		rc, err = r.Run(context.Background())
	} else {
		rc, err = supervise(context.Background(), cmd, nil)
	}
	if !errors.Is(err, syscall.EMFILE) {
		t.Fatalf("%s: rc=%d err=%v; want EMFILE launch failure, not success", mode, rc, err)
	}
}
