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

// Closing-review reds (reviewer-3, G-F6 + G-F1), ported onto the shared test storage.

// TestReviewG6TerminalStdinCompletes: offload ctr on a terminal returns CMD's status after a Ctrl-C
// (podman once sat in a background process group of that terminal and never returned).
func TestReviewG6TerminalStdinCompletes(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	log, pidfile := s.path("trap.log"), s.path("offload.pid")
	in, input, err := os.Pipe()
	must(t, err)
	defer func() { _ = in.Close(); _ = input.Close() }()
	out, err := os.Create(s.path("terminal.log"))
	must(t, err)
	defer func() { _ = out.Close() }()
	script := fmt.Sprintf("echo $$ > '%s'; exec '%s' ctr -- ctrhelper trap '%s' 3 5", pidfile, binaryPath, log)
	cmd := exec.Command("script", "-q", "-e", "-c", script, "/dev/null")
	cmd.Dir, cmd.Env, cmd.Stdin, cmd.Stdout, cmd.Stderr = s.cwd, s.env, in, out, out
	must(t, cmd.Start())
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	defer func() {
		if exists(pidfile) {
			_ = syscall.Kill(readPID(t, pidfile), syscall.SIGKILL)
		}
		_ = cmd.Process.Kill()
	}()
	deadline := time.Now().Add(10 * time.Second)
	for !strings.HasPrefix(readOr(log), "ready\n") {
		if time.Now().After(deadline) {
			t.Fatalf("terminal command did not become ready: %s", readOr(out.Name()))
		}
		time.Sleep(20 * time.Millisecond)
	}
	_, err = input.Write([]byte{3}) // the terminal's Ctrl-C
	must(t, err)
	select {
	case err := <-done:
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 5 {
			t.Fatalf("terminal command status = %v, want 5", err)
		}
		if got := readOr(log); got != "ready\nINTERRUPT\n" {
			t.Fatalf("trap log = %q, want exactly one INTERRUPT", got)
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("offload ctr did not return after CMD exited; log=%q", readOr(log))
	}
}

// TestReviewG6PodmanDeathIs125: podman killed while the container runs → 125 promptly, and
// offload ctr's exit ends the container (K7, K8).
func TestReviewG6PodmanDeathIs125(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	marker := fmt.Sprintf("review-g6-%d", time.Now().UnixNano())
	ready := s.path("ready")
	p := s.start("ctr", "--", "ctrhelper", "ignore", marker, ready)
	p.await("container start", func() bool { return exists(ready) })
	kids, err := exec.Command("pgrep", "-P", strconv.Itoa(p.cmd.Process.Pid)).Output()
	must(t, err)
	fields := strings.Fields(string(kids))
	if len(fields) != 1 {
		t.Fatalf("want podman as the one child, got %q", kids)
	}
	pid, err := strconv.Atoi(fields[0])
	must(t, err)
	must(t, syscall.Kill(pid, syscall.SIGKILL))
	select {
	case <-p.done:
		p.wait(time.Second).requireCode(t, 125)
	case <-time.After(5 * time.Second):
		t.Fatal("offload ctr still running 5 s after podman's SIGKILL")
	}
	markerGone(t, marker, 5*time.Second)
}

// TestReviewA3SignalCancelsProjectLookup (reviewer-1 F-A3): a K8 signal during the git lookup of
// setup ends the run with 128+n at once, for run, --layers and --check alike.
func TestReviewA3SignalCancelsProjectLookup(t *testing.T) {
	t.Parallel()
	for name, args := range map[string][]string{
		"run":    {"ctr", "--", "true"},
		"layers": {"ctr", "--layers"},
		"check":  {"ctr", "--check"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s := newSandbox(t, true)
			s.setenv("OFFLOAD_CTR", "1")
			ready := filepath.Join(s.base, "git-ready")
			s.setenv("REVIEW_GIT_READY", ready)
			git := filepath.Join(s.base, "bin", "git")
			s.write(git, "#!/bin/sh\necho $$ > \"$REVIEW_GIT_READY\"\nexec sleep 30\n")
			must(t, os.Chmod(git, 0o755))
			p := s.start(args...)
			p.await("git lookup", func() bool { return exists(ready) })
			t.Cleanup(func() { _ = syscall.Kill(readPID(t, ready), syscall.SIGKILL) })
			p.signal(syscall.SIGTERM)
			select {
			case <-p.done:
				p.wait(time.Second).requireCode(t, 143)
			case <-time.After(2 * time.Second):
				t.Fatal("SIGTERM during the project lookup did not cancel git")
			}
		})
	}
}
