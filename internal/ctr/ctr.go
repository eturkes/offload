// Package ctr runs a command in the project's own image (`offload ctr`, K1–K11): one image per
// host, built on first use from `.devcontainer/Containerfile` and tagged by its content, with an
// init of offload's own as PID 1 so the command dies with `offload ctr` whatever kills it.
package ctr

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/eturkes/offload/internal/mirror"
	"github.com/eturkes/offload/internal/remote"
	"github.com/eturkes/offload/internal/slots"
)

// Definition = the image definition, relative to the project root (K3).
const Definition = ".devcontainer/Containerfile"

// Paths inside the container (K6): root-level, so no project root on any host sits under them.
const (
	initPath = "/.offload/init"
	homePath = "/.offload/home"
)

// Status values offload owns (K3, K7).
const (
	ExitUsage   = 2
	exitFailure = 125
)

// UsageError = an invocation or setup problem found before any work (rc 2).
type UsageError struct{ error }

// Project = the root a ctr run mounts + the image it runs in.
type Project struct {
	Root       string // symlink-resolved git toplevel, else cwd
	Cwd        string // symlink-resolved
	CommonDir  string // git common dir when it lies outside Root (a linked worktree's), else ""
	Definition string // <Root>/.devcontainer/Containerfile
	Context    string // the definition's directory
	Binary     string // this offload binary: the container's init
}

// Find resolves the project around cwd (K3).
func Find(ctx context.Context, cwd string) (Project, error) {
	var p Project
	var err error
	if p.Cwd, err = filepath.EvalSymlinks(cwd); err != nil {
		return p, err
	}
	l, err := mirror.Locate(ctx, p.Cwd)
	if err != nil {
		return p, err
	}
	if p.Root, err = filepath.EvalSymlinks(l.Root); err != nil {
		return p, err
	}
	if l.Git {
		var common string
		if common, err = filepath.EvalSymlinks(l.CommonDir); err != nil {
			return p, err
		}
		if !within(p.Root, common) {
			p.CommonDir = common
		}
	}
	p.Definition = filepath.Join(p.Root, Definition)
	p.Context = filepath.Dir(p.Definition)
	// Lstat both: a symlinked .devcontainer would hash empty (Tag never walks into a symlinked root).
	dir, derr := os.Lstat(p.Context)
	def, ferr := os.Lstat(p.Definition)
	if derr != nil || ferr != nil || !dir.IsDir() || !def.Mode().IsRegular() {
		return p, UsageError{fmt.Errorf("no image definition at %s (a regular file in a real directory, no symlinks)", p.Definition)}
	}
	if p.Binary, err = Executable(); err != nil {
		return p, err
	}
	return p, nil
}

// Executable = this binary's path; a replaced one (its old inode unlinked) → error.
func Executable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if exe, err = filepath.EvalSymlinks(exe); err != nil {
		return "", fmt.Errorf("the offload binary was replaced while it ran; run the command again: %w", err)
	}
	return exe, nil
}

func within(root, path string) bool {
	return path == root || strings.HasPrefix(path, root+string(filepath.Separator))
}

// Ensure builds the project's image unless its tag exists (K5) and returns the tag.
func Ensure(ctx context.Context, p Project, getenv func(string) string, stderr io.Writer) (string, error) {
	tag, err := Tag(p.Context)
	if err != nil {
		return "", err
	}
	if ok, xerr := exists(ctx, tag); ok || xerr != nil {
		return tag, xerr
	}
	lock, err := slots.WaitLock(ctx, filepath.Join(slots.RuntimeDir(getenv), "offload", "ctr", strings.TrimPrefix(tag, Repo+":")+".lock"))
	if err != nil {
		return "", err
	}
	defer func() { _ = lock.Close() }()
	if ok, xerr := exists(ctx, tag); ok || xerr != nil { // another run built it while we waited
		return tag, xerr
	}
	fmt.Fprintf(stderr, "offload: building %s from %s\n", tag, p.Definition)
	cmd := remote.Command(ctx, "podman", "build", "--timestamp", "0", "-t", tag, "-f", p.Definition, p.Context)
	var log bytes.Buffer
	cmd.Stdout, cmd.Stderr = &log, &log
	cmd.ExtraFiles = []*os.File{lock} // the lock lives as long as the build, even past offload's death
	if err := cmd.Run(); err != nil {
		if ctx.Err() == nil {
			_, _ = stderr.Write(log.Bytes())
		}
		return "", fmt.Errorf("podman build: %w", err)
	}
	return tag, nil
}

// exists = `podman image exists`: 0 present, 1 absent, anything else an error (K5).
func exists(ctx context.Context, tag string) (bool, error) {
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "podman", "image", "exists", tag) //nolint:gosec // G204: tag = Repo + hex
	cmd.Stderr = &stderr
	err := cmd.Run()
	var ee *exec.ExitError
	switch {
	case err == nil:
		return true, nil
	case errors.As(err, &ee) && ee.ExitCode() == 1:
		return false, nil
	}
	return false, fmt.Errorf("podman image exists: %w %s", err, strings.TrimSpace(stderr.String()))
}

// Layers = the image's RootFS layer digests (K9).
func Layers(ctx context.Context, tag string) ([]string, error) {
	out, err := exec.CommandContext(ctx, "podman", "image", "inspect", "--format", //nolint:gosec // G204: tag = Repo + hex
		"{{range .RootFS.Layers}}{{println .}}{{end}}", tag).Output()
	if err != nil {
		return nil, fmt.Errorf("podman image inspect %s: %w", tag, err)
	}
	return strings.Fields(string(out)), nil
}

// Home = the persistent per-root HOME directory on this host (K6).
func Home(getenv func(string) string, root string) (string, error) {
	cache := getenv("XDG_CACHE_HOME")
	if cache == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		cache = filepath.Join(home, ".cache")
	}
	return filepath.Join(cache, "offload", "ctr", mirror.Name(root), "home"), nil
}

// RunArgs = the podman argv (without "podman") running argv in tag (K6).
func RunArgs(p Project, tag, home string, argv []string) []string {
	args := []string{
		// --http-proxy=false: podman otherwise copies the host's proxy variables in (K6: host env stays out).
		"run", "--rm", "--pull=never", "--pid=private", "--userns=keep-id", "--log-driver=none", "--http-proxy=false",
		"-i", "--preserve-fds=1",
		"--entrypoint", initPath, "-v", p.Binary + ":" + initPath + ":ro", "-v", p.Root + ":" + p.Root,
	}
	if p.CommonDir != "" {
		args = append(args, "-v", p.CommonDir+":"+p.CommonDir)
	}
	args = append(args, "-v", home+":"+homePath, "-e", "HOME="+homePath,
		"--passwd-entry", "$USERNAME:*:$UID:$GID:$NAME:"+homePath+":/bin/sh", "-w", p.Cwd, tag, "__ctr-init")
	return append(args, argv...)
}

// Run executes argv in the image of the project around cwd and returns CMD's status (K3–K8). sigs
// carries every Forwarded signal offload received: before the container starts the first one ends
// the run (128+n); afterwards each goes to CMD over the control socket.
func Run(cwd string, argv []string, getenv func(string) string, sigs <-chan os.Signal, stderr io.Writer) int {
	return run(func(ctx context.Context) (Project, error) { return Find(ctx, cwd) }, argv, getenv, sigs, stderr)
}

// run = Run with the project lookup injected (tests hand it a fixed Project).
func run(find func(context.Context) (Project, error), argv []string, getenv func(string) string, sigs <-chan os.Signal, stderr io.Writer) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var mu sync.Mutex
	var ctl *os.File
	var got syscall.Signal
	go func() {
		for s := range sigs {
			n, _ := s.(syscall.Signal)
			mu.Lock()
			if ctl != nil {
				_, _ = ctl.Write([]byte{byte(n)}) //nolint:gosec // G115: Forwarded signals are < 32
			} else if got == 0 {
				got = n
				cancel()
			}
			mu.Unlock()
		}
	}()
	fail := func(err error) int {
		mu.Lock()
		defer mu.Unlock()
		if got != 0 {
			return 128 + int(got)
		}
		fmt.Fprintf(stderr, "offload: %v\n", err)
		return exitFailure
	}

	p, err := find(ctx)
	var ue UsageError
	if errors.As(err, &ue) {
		fmt.Fprintf(stderr, "offload: %v\n", err)
		return ExitUsage
	}
	if err != nil {
		return fail(err)
	}
	tag, err := Ensure(ctx, p, getenv, stderr)
	if err != nil {
		return fail(err)
	}
	home, err := Home(getenv, p.Root)
	if err == nil {
		err = os.MkdirAll(home, 0o700)
	}
	if err == nil {
		err = mountable(p.Binary, p.Root, p.CommonDir, home)
	}
	if err == nil {
		err = running(p.Binary)
	}
	if err != nil {
		return fail(err)
	}
	mine, theirs, err := controlPair()
	if err != nil {
		return fail(err)
	}
	defer func() { _ = mine.Close() }()

	cmd := podmanCommand(ctx, getenv, RunArgs(p, tag, home, argv))
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, stderr
	cmd.ExtraFiles = []*os.File{theirs}
	// Own session: a session-wide or terminal signal reaches offload alone, which forwards each once;
	// no controlling terminal ⇒ podman reads a terminal stdin without stalling, as a background
	// process group of that terminal would (SIGTTIN).
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	mu.Lock()
	if got != 0 {
		mu.Unlock()
		_ = theirs.Close()
		return 128 + int(got)
	}
	err = cmd.Start()
	if err == nil {
		ctl = mine
	}
	mu.Unlock()
	_ = theirs.Close()
	if err != nil {
		return fail(err)
	}
	waitErr := cmd.Wait()
	mu.Lock()
	ctl = nil
	mu.Unlock()

	if status, ok := readRecord(mine); ok {
		return status
	}
	return fail(fmt.Errorf("the container ended without its command's status (podman: %w)", waitErr))
}

// podmanCommand = podman with args, in a transient systemd scope of its own where the user's
// manager starts one (K8): podman dies on TERM (always on a second, at times on a first: 4.9.3,
// 5.8.8), so no signal aimed at its caller's cgroup — X1's scope TERM + KILL — may reach it.
// The caller's INVOCATION_ID never reaches podman (with it podman keeps conmon in the caller's
// cgroup); in the scope, systemd-run sets the scope's own.
func podmanCommand(ctx context.Context, getenv func(string) string, args []string) *exec.Cmd {
	env := slices.DeleteFunc(os.Environ(), func(kv string) bool { return strings.HasPrefix(kv, "INVOCATION_ID=") })
	if rt := scopeManager(ctx, getenv); rt != "" {
		args = append([]string{"--user", "--scope", "--quiet", "--collect", "--expand-environment=no", "--", "podman"}, args...)
		cmd := exec.Command("systemd-run", args...)  //nolint:gosec // G204: podman + RunArgs
		cmd.Env = append(env, "XDG_RUNTIME_DIR="+rt) // the manager's; the last duplicate wins
		return cmd
	}
	cmd := exec.Command("podman", args...) //nolint:gosec // G204: RunArgs, the caller's command as given
	cmd.Env = env
	return cmd
}

// scopeManager = the runtime dir whose systemd user manager starts a transient scope (cgroup v2,
// probed with `true` under ctx, ≤ scopeProbe), else "" (podman then runs in offload's cgroup).
func scopeManager(ctx context.Context, getenv func(string) string) string {
	if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err != nil {
		return ""
	}
	for _, d := range []string{getenv("XDG_RUNTIME_DIR"), "/run/user/" + strconv.Itoa(os.Getuid())} {
		if fi, err := os.Stat(filepath.Join(d, "systemd", "private")); d == "" || err != nil || fi.Mode()&os.ModeSocket == 0 {
			continue
		}
		pctx, cancel := context.WithTimeout(ctx, scopeProbe)
		defer cancel()
		probe := exec.CommandContext(pctx, "systemd-run", "--user", "--scope", "--quiet", "--collect", "--expand-environment=no", "--", "true")
		probe.Env = append(os.Environ(), "XDG_RUNTIME_DIR="+d)
		if probe.Run() == nil {
			return d
		}
		return ""
	}
	return ""
}

// scopeProbe bounds the scope probe: a manager that stalls longer gets no scope (podman runs plainly).
const scopeProbe = 5 * time.Second

// controlPair = the control socket (K8): mine for offload, theirs for the container (fd 3).
func controlPair() (mine, theirs *os.File, err error) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM|syscall.SOCK_CLOEXEC, 0)
	if err != nil {
		return nil, nil, err
	}
	// Nonblocking ⇒ os.NewFile puts it on Go's poller, the one place SetReadDeadline works: a blocking
	// fd ignores the deadline, and readRecord would hang while conmon holds the container's end.
	if err = syscall.SetNonblock(fds[0], true); err != nil {
		_ = syscall.Close(fds[0])
		_ = syscall.Close(fds[1])
		return nil, nil, err
	}
	return os.NewFile(uintptr(fds[0]), "control"), os.NewFile(uintptr(fds[1]), "control"), nil
}

// mountable refuses paths podman's `-v src:dst` cannot carry (K3).
func mountable(paths ...string) error {
	for _, path := range paths {
		if strings.Contains(path, ":") {
			return fmt.Errorf("cannot mount %s into the container: the path holds ':'", path)
		}
	}
	return nil
}

// running = path is still this process's binary (K6): an install that replaced it since Find would
// put another version in the container as its init.
func running(path string) error {
	self, err := os.Stat("/proc/self/exe")
	if err != nil {
		return err
	}
	if cur, err := os.Stat(path); err != nil || !os.SameFile(self, cur) {
		return errors.New("the offload binary was replaced while it ran; run the command again")
	}
	return nil
}

// readRecord reads the init's `<status>\n`. It is written before the init exits, so it is buffered
// by the time podman returns; the deadline only bounds a missing one (podman's copies of the
// container's end may delay EOF).
func readRecord(f *os.File) (int, bool) {
	_ = f.SetReadDeadline(time.Now().Add(time.Second))
	var rec []byte
	b := make([]byte, 16)
	for len(rec) < 16 {
		n, err := f.Read(b)
		rec = append(rec, b[:n]...)
		if line, _, found := bytes.Cut(rec, []byte("\n")); found {
			status, aerr := strconv.Atoi(string(line))
			return status, aerr == nil && status >= 0
		}
		if err != nil {
			break
		}
	}
	return 0, false
}

// Notify = a channel of every Forwarded signal, for Run.
func Notify() <-chan os.Signal {
	ch := make(chan os.Signal, 16)
	signal.Notify(ch, Forwarded...)
	return ch
}
