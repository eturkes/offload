package ctr

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/eturkes/offload/internal/shell"
)

// Forwarded = the signals `offload ctr` hands to CMD (K8): over the control socket, one byte each.
var Forwarded = []os.Signal{
	syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2,
}

// grace = KILL this long after the teardown TERM (X1's 10 s).
const grace = 10 * time.Second

// Init is the container's PID 1 (K8): argv = CMD, fd 3 = the control socket. It runs CMD in its
// own process group, delivers each signal byte from the socket to that group, reaps every process,
// TERMs the rest once CMD exits (KILL after grace = its own exit), records CMD's status on the
// socket and exits with it. Socket EOF = `offload ctr` is gone: exit at once, which KILLs the pid
// namespace.
func Init(argv []string, stderr io.Writer) int {
	ctl := os.NewFile(3, "control")
	// PID 1 alone: the teardown's kill(-1) outside a pid namespace would hit every process of the user.
	if os.Getpid() != 1 || ctl == nil || len(argv) == 0 {
		fmt.Fprintln(stderr, "offload: __ctr-init runs inside offload ctr's container only")
		return 125
	}
	syscall.CloseOnExec(3) // podman passes it inheritable; CMD must not write the status channel
	// Handlers, not signal.Ignore: an ignored disposition survives exec, so CMD would start deaf.
	// Signals arriving as signals (podman's sig-proxy, a scope-wide kill) are dropped: the socket
	// carries each one exactly once.
	signal.Notify(make(chan os.Signal, 1), Forwarded...)
	chld := make(chan os.Signal, 1)
	signal.Notify(chld, syscall.SIGCHLD)

	pid, status := start(argv, stderr)
	if pid == 0 {
		return record(ctl, status)
	}
	fwd := make(chan syscall.Signal, 8)
	go func() {
		b := make([]byte, 16)
		for {
			n, err := ctl.Read(b)
			for _, s := range b[:n] {
				fwd <- syscall.Signal(s)
			}
			if err != nil {
				os.Exit(137) // liveness lost (K8)
			}
		}
	}()

	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	var deadline time.Time
	for {
		select {
		case <-chld:
		case <-tick.C:
		case s := <-fwd:
			if deadline.IsZero() {
				_ = syscall.Kill(-pid, s)
			}
			continue
		}
		done, empty := reap(pid)
		if done >= 0 && deadline.IsZero() {
			status = done
			_ = syscall.Kill(-1, syscall.SIGTERM) // every process but PID 1
			deadline = time.Now().Add(grace)
		}
		if !deadline.IsZero() && (empty || time.Now().After(deadline)) {
			return record(ctl, status)
		}
	}
}

// start forks CMD as its own process group leader; pid 0 = it never ran, status = the shell's
// 127 (not found) or 126 (not executable).
func start(argv []string, stderr io.Writer) (pid, status int) {
	path, status, err := shell.Look(argv[0])
	if status == shell.NotFound {
		fmt.Fprintf(stderr, "offload: %s: command not found\n", argv[0])
		return 0, status
	}
	if status != 0 {
		fmt.Fprintf(stderr, "offload: %s: %v\n", path, err)
		return 0, status
	}
	pid, err = syscall.ForkExec(path, argv, &syscall.ProcAttr{ //nolint:gosec // G204: CMD = the caller's command, as given
		Env: os.Environ(), Files: []uintptr{0, 1, 2}, Sys: &syscall.SysProcAttr{Setpgid: true},
	})
	if err != nil {
		fmt.Fprintf(stderr, "offload: %s: %v\n", argv[0], err)
		return 0, 126
	}
	return pid, 0
}

// reap collects every exited child: done = CMD's shell-style status once reaped (else -1), empty =
// no child is left.
func reap(cmd int) (done int, empty bool) {
	done = -1
	for {
		var ws syscall.WaitStatus
		pid, err := syscall.Wait4(-1, &ws, syscall.WNOHANG, nil)
		switch {
		case errors.Is(err, syscall.EINTR):
			continue
		case err != nil: // ECHILD
			return done, true
		case pid == 0:
			return done, false
		case pid == cmd:
			done = ws.ExitStatus()
			if ws.Signaled() {
				done = 128 + int(ws.Signal())
			}
		}
	}
}

func record(ctl *os.File, status int) int {
	_, _ = ctl.WriteString(strconv.Itoa(status) + "\n")
	return status
}
