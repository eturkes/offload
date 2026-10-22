package ctr

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

// reviewDenySyscalls injects errno into selected syscalls on this subprocess's calling thread.
// The parent test + every other process retain their ordinary syscall behavior.
func reviewDenySyscalls(t *testing.T, errno syscall.Errno, numbers ...uint32) {
	t.Helper()
	runtime.LockOSThread()
	filters := []syscall.SockFilter{{Code: 0x20, K: 0}}
	for _, number := range numbers {
		filters = append(filters, syscall.SockFilter{Code: 0x15, Jf: 1, K: number}, syscall.SockFilter{Code: 0x06, K: 0x00050000 | uint32(errno)})
	}
	filters = append(filters, syscall.SockFilter{Code: 0x06, K: 0x7fff0000})
	program := syscall.SockFprog{Len: uint16(len(filters)), Filter: &filters[0]}
	_, _, err := syscall.RawSyscall6(syscall.SYS_PRCTL, 38, 1, 0, 0, 0, 0)
	must(t, errErrno(err))
	_, _, err = syscall.RawSyscall6(syscall.SYS_PRCTL, 22, 2, uintptr(unsafe.Pointer(&program)), 0, 0, 0)
	must(t, errErrno(err))
}

func errErrno(errno syscall.Errno) error {
	if errno == 0 {
		return nil
	}
	return errno
}

func TestK6ReviewUnavailableSelfIdentity(t *testing.T) {
	if os.Getenv("OFFLOAD_REVIEW_STAT_ERROR") == "1" {
		reviewDenySyscalls(t, syscall.EPERM, syscall.SYS_NEWFSTATAT, syscall.SYS_STAT)
		if _, err := os.Stat("/proc/self/exe"); !errors.Is(err, syscall.EPERM) {
			t.Fatalf("positive control: self stat = %v, want injected EPERM", err)
		}
		if err := running("/proc/self/exe"); err == nil {
			fmt.Fprintln(os.Stderr, "accepted binary without self identity")
			os.Exit(93)
		}
		fmt.Fprintln(os.Stdout, "identity unavailable: refused")
		os.Exit(0)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestK6ReviewUnavailableSelfIdentity$")
	cmd.Env = append(os.Environ(), "OFFLOAD_REVIEW_STAT_ERROR=1")
	out, err := cmd.CombinedOutput()
	if err != nil || string(out) != "identity unavailable: refused\n" {
		t.Fatalf("unreadable self identity must refuse: err=%v output=%q", err, out)
	}
}

func TestK3ReviewFindFailures(t *testing.T) {
	for _, mode := range []string{"cwd", "git", "root", "common"} {
		t.Run(mode, func(t *testing.T) {
			p := reviewProject(t)
			must(t, os.Mkdir(filepath.Join(p.Root, ".git"), 0o700))
			cwd, root, common := p.Root, p.Root, filepath.Join(p.Root, ".git")
			switch mode {
			case "cwd":
				cwd += "/absent"
			case "root":
				root += "/absent"
			case "common":
				common += "/absent"
			}
			dir := t.TempDir()
			script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n%%s\\n%%s\\n\\n' %q %q %q\n", root, common, common)
			if mode == "git" {
				script = "#!/bin/sh\necho 'fatal: injected git failure' >&2\nexit 1\n"
			}
			must(t, os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0o755))
			t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
			_, err := Find(context.Background(), cwd)
			var usage UsageError
			if err == nil || errors.As(err, &usage) {
				t.Fatalf("%s discovery failure = %v; want non-usage error (K3 rc 125)", mode, err)
			}
		})
	}
}

func TestK7ReviewSocketpairFailure(t *testing.T) {
	if root := os.Getenv("OFFLOAD_REVIEW_SOCKET_ERROR"); root != "" {
		reviewDenySyscalls(t, syscall.EMFILE, syscall.SYS_SOCKETPAIR)
		binary, err := Executable()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(91)
		}
		p := Project{Root: root, Cwd: root, Context: filepath.Join(root, ".devcontainer"), Definition: filepath.Join(root, Definition), Binary: binary}
		getenv := func(k string) string {
			if k == "XDG_CACHE_HOME" {
				return filepath.Join(root, "cache")
			}
			return ""
		}
		os.Exit(run(fixedProject(p), []string{"unused-command"}, getenv, nil, os.Stderr))
	}
	p := reviewProject(t)
	reviewPodman(t, "[ \"$1 $2\" = 'image exists' ] && exit 0\ntouch \"$OFFLOAD_REVIEW_SOCKET_ERROR/ran\"\nprintf '7\\n' >&3\nexit 42\n")
	input, err := os.CreateTemp(t.TempDir(), "stdin-")
	must(t, err)
	defer func() { _ = input.Close() }()
	_, err = input.WriteString("7\n7\n")
	must(t, err)
	_, err = input.Seek(0, 0)
	must(t, err)
	cmd := exec.Command(os.Args[0], "-test.run=^TestK7ReviewSocketpairFailure$")
	cmd.Env = append(os.Environ(), "OFFLOAD_REVIEW_SOCKET_ERROR="+p.Root)
	cmd.Stdin = input
	out, err := cmd.CombinedOutput()
	var ee *exec.ExitError
	if !errors.As(err, &ee) || ee.ExitCode() != 125 || !strings.Contains(string(out), "offload:") {
		t.Fatalf("socketpair failure must stay rc 125: err=%v output=%q", err, out)
	}
	if _, err := os.Stat(filepath.Join(p.Root, "ran")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Podman started without a control socket: %v", err)
	}
}
