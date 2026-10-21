package run

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func alive(pid int) bool { return syscall.Kill(pid, 0) == nil }

func readPids(t *testing.T, path string, n int) []int {
	t.Helper()
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		b, _ := os.ReadFile(path)
		if f := strings.Fields(string(b)); len(f) >= n {
			pids := make([]int, n)
			for i := range pids {
				pids[i], _ = strconv.Atoi(f[i])
			}
			return pids
		}
	}
	t.Fatalf("no %d pids in %s", n, path)
	return nil
}

func TestLocalStatus(t *testing.T) { // C3
	for cmd, want := range map[string]int{"exit 0": 0, "exit 7": 7, "exit 255": 255, "kill -TERM $$": 143} {
		rc, err := Local(context.Background(), []string{"sh", "-c", cmd}, nil)
		if err != nil || rc != want {
			t.Errorf("%q → %d, %v; want %d", cmd, rc, err, want)
		}
	}
}

func TestCommandOutlivesOpenStdin(t *testing.T) { // X1: only EOF or the leader's exit end the run
	rc, err := Local(context.Background(), []string{"sh", "-c", "sleep 1.5; exit 4"}, nil)
	if err != nil || rc != 4 {
		t.Fatalf("rc %d err %v; want 4 (a teardown before EOF reads as 143)", rc, err)
	}
}

func TestLocalArgvIsNotShellParsed(t *testing.T) { // X2
	out := filepath.Join(t.TempDir(), "out")
	rc, err := Local(context.Background(), []string{"sh", "-c", `printf '%s' "$1" > "$2"`, "_", "$HOME `x` 'q'", out}, nil)
	if b, _ := os.ReadFile(out); err != nil || rc != 0 || string(b) != "$HOME `x` 'q'" {
		t.Fatalf("rc %d err %v out %q", rc, err, b)
	}
}

func TestCancelEndsWholeSession(t *testing.T) { // X1 + X3: setpgid child, TERM-ignoring leader
	pids := filepath.Join(t.TempDir(), "pids")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int)
	go func() {
		rc, _ := Local(ctx, []string{"bash", "-c", `trap "" TERM
perl -e 'setpgrp(0,0); $SIG{TERM}="IGNORE"; sleep 300' & echo $$ $! > ` + pids + `; sleep 300`}, nil)
		done <- rc
	}()
	p := readPids(t, pids, 2)
	time.Sleep(3 * time.Second) // nothing may end before the cancel
	for _, pid := range p {
		if !alive(pid) {
			t.Fatalf("pid %d ended before offload went away", pid)
		}
	}
	t0 := time.Now()
	cancel()
	select {
	case rc := <-done:
		if rc != 137 || time.Since(t0) < 9*time.Second {
			t.Errorf("rc %d after %s; want 137 after the 10 s grace", rc, time.Since(t0))
		}
	case <-time.After(30 * time.Second):
		t.Fatal("watcher did not finish")
	}
	for _, pid := range p {
		if alive(pid) {
			t.Errorf("pid %d survived", pid)
		}
	}
}

func TestStragglersEndWithLeader(t *testing.T) { // X1
	pids := filepath.Join(t.TempDir(), "pids")
	rc, err := Local(context.Background(), []string{"bash", "-c", "sleep 300 & echo $! > " + pids + "; exit 3"}, nil)
	p := readPids(t, pids, 1)
	if err != nil || rc != 3 || alive(p[0]) {
		t.Fatalf("rc %d err %v straggler alive %v", rc, err, alive(p[0]))
	}
}

func TestLockReachesWatcherNotCommand(t *testing.T) { // S3 + S5
	lock, err := os.Create(filepath.Join(t.TempDir(), "slot1"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Close() }()
	out := filepath.Join(t.TempDir(), "fds")
	rc, err := Local(context.Background(), []string{"bash", "-c", "ls -l /proc/$$/fd/ > " + out}, lock)
	b, _ := os.ReadFile(out)
	if err != nil || rc != 0 || strings.Contains(string(b), "slot1") {
		t.Fatalf("rc %d err %v; command fds:\n%s", rc, err, b)
	}
}

// remote.bash, driven directly with a temp HOME as the "host".
type host struct {
	t    *testing.T
	home string
}

func newHost(t *testing.T) host {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	return host{t, t.TempDir()}
}

func (h host) path(p string) string { return filepath.Join(h.home, p) }

func (h host) write(p, body string) {
	h.t.Helper()
	if err := os.MkdirAll(filepath.Dir(h.path(p)), 0o755); err != nil {
		h.t.Fatal(err)
	}
	if err := os.WriteFile(h.path(p), []byte(body), 0o644); err != nil {
		h.t.Fatal(err)
	}
}

func (h host) exists(p string) bool { _, err := os.Lstat(h.path(p)); return err == nil }

// run executes the remote script with stdin = offload's pipe: the framed list, then held open until
// the script exits (as offload does); eof = close right after the list (offload died mid-list).
func (h host) run(mode, nonce string, count int, stdin []byte, cmd string, eof bool) (int, string) {
	h.t.Helper()
	c := exec.Command("bash", "-c", watchScript+remoteScript, "offload", "offload/w", "offload/.meta/w", ".",
		mode, nonce, strconv.Itoa(count), "", cmd)
	c.Env = append(os.Environ(), "HOME="+h.home)
	in, err := c.StdinPipe()
	if err != nil {
		h.t.Fatal(err)
	}
	var stderr bytes.Buffer
	c.Stderr = &stderr
	if err = c.Start(); err != nil {
		h.t.Fatal(err)
	}
	_, _ = in.Write(stdin)
	if eof {
		_ = in.Close()
	}
	err = c.Wait()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode(), stderr.String()
	}
	if err != nil {
		h.t.Fatal(err)
	}
	return 0, stderr.String()
}

func frame(nonce string, paths ...string) []byte {
	return []byte(nonce + "\x00" + strings.Join(append(paths, ""), "\x00"))
}

func (h host) gitWorkdir() {
	h.t.Helper()
	w := h.path("offload/w")
	if err := os.MkdirAll(w, 0o755); err != nil {
		h.t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", w, "init", "-q").CombinedOutput(); err != nil {
		h.t.Fatalf("%v %s", err, out)
	}
}

func TestRemotePrune(t *testing.T) { // Y5
	h := newHost(t)
	h.gitWorkdir()
	h.write("offload/w/.gitignore", "cache/\n")
	h.write("offload/w/keep.txt", "k")
	h.write("offload/w/stale.txt", "s")
	h.write("offload/w/deep/stale.txt", "s")
	h.write("offload/w/cache/venv.bin", "v")
	h.write("offload/victim/file", "precious")
	if err := os.Symlink("../victim", h.path("offload/w/link")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(h.path("offload/w/emptydir"), 0o755); err != nil {
		t.Fatal(err)
	}
	// link/file is listed by git through the symlink only if tracked: track it via the index.
	if out, err := exec.Command("git", "-C", h.path("offload/w"), "update-index", "--add", "--cacheinfo",
		"100644,e69de29bb2d1d6434b8b29ae775ad8c2e48c5391,link/file").CombinedOutput(); err != nil {
		t.Fatalf("%v %s", err, out)
	}

	rc, stderr := h.run("git", "N1", 3, frame("N1", ".gitignore", "keep.txt", "link"), "true", false)
	if rc != 0 {
		t.Fatalf("rc %d: %s", rc, stderr)
	}
	for p, want := range map[string]bool{
		"offload/w/keep.txt": true, "offload/w/stale.txt": false, "offload/w/deep/stale.txt": false,
		"offload/w/deep": true, "offload/w/cache/venv.bin": true, "offload/victim/file": true,
		"offload/w/emptydir": true, "offload/w/link": true,
	} {
		if h.exists(p) != want {
			t.Errorf("%s exists = %v, want %v", p, !want, want)
		}
	}
	status, _ := os.ReadFile(h.path("offload/.meta/w/status"))
	if string(status) != "N1 0\n" {
		t.Errorf("status record %q", status)
	}
}

func TestRemoteFramingMismatchPrunesNothing(t *testing.T) { // Y5
	for name, tc := range map[string]struct {
		nonce string
		count int
		in    []byte
	}{
		"wrong nonce":  {"N2", 1, frame("XX", "keep.txt")},
		"short list":   {"N2", 3, frame("N2", "keep.txt")},
		"empty stdin":  {"N2", 1, nil},
		"unterminated": {"N2", 1, []byte("N2\x00keep.txt")},
	} {
		t.Run(name, func(t *testing.T) {
			h := newHost(t)
			h.gitWorkdir()
			h.write("offload/w/keep.txt", "k")
			h.write("offload/w/stale.txt", "s")
			rc, stderr := h.run("git", tc.nonce, tc.count, tc.in, "touch ran", true)
			if rc != 125 || !strings.Contains(stderr, "framing mismatch") {
				t.Fatalf("rc %d stderr %q", rc, stderr)
			}
			if !h.exists("offload/w/stale.txt") || h.exists("offload/w/ran") || h.exists("offload/.meta/w/status") {
				t.Fatal("pruned, ran or recorded a status after a framing mismatch")
			}
		})
	}
}

func TestRemoteRefusesMissingGit(t *testing.T) { // Y5
	h := newHost(t)
	h.write("offload/w/stale.txt", "s")
	rc, stderr := h.run("git", "N3", 0, frame("N3"), "touch ran", false)
	if rc != 125 || !strings.Contains(stderr, "no .git") || !h.exists("offload/w/stale.txt") || h.exists("offload/w/ran") {
		t.Fatalf("rc %d stderr %q", rc, stderr)
	}
}

func TestRemoteDownList(t *testing.T) { // Y6: existing, in-tree, non-ignored paths only
	h := newHost(t)
	h.gitWorkdir()
	h.write("offload/w/.gitignore", "*.log\n")
	h.write("offload/victim/file", "v")
	if err := os.Symlink("../victim", h.path("offload/w/link")); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", h.path("offload/w"), "update-index", "--add", "--cacheinfo",
		"100644,e69de29bb2d1d6434b8b29ae775ad8c2e48c5391,gone.txt").CombinedOutput(); err != nil {
		t.Fatalf("%v %s", err, out)
	}
	rc, stderr := h.run("git", "N4", 2, frame("N4", ".gitignore", "link"), "echo o > out.txt; echo l > run.log; exit 9", false)
	if rc != 9 {
		t.Fatalf("rc %d: %s", rc, stderr)
	}
	b, _ := os.ReadFile(h.path("offload/.meta/w/down"))
	got := strings.Split(strings.TrimSuffix(string(b), "\x00"), "\x00")
	slices.Sort(got)
	if want := []string{".gitignore", "link", "out.txt"}; !slices.Equal(got, want) {
		t.Fatalf("down = %q, want %q", got, want)
	}
	if status, _ := os.ReadFile(h.path("offload/.meta/w/status")); string(status) != "N4 9\n" {
		t.Fatalf("status %q", status)
	}
}

func TestRemoteLockSerializes(t *testing.T) { // Y11
	h := newHost(t)
	if err := os.MkdirAll(h.path("offload/.meta/w"), 0o755); err != nil {
		t.Fatal(err)
	}
	holder := exec.Command("flock", h.path("offload/.meta/w/lock"), "sleep", "2")
	if err := holder.Start(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	t0 := time.Now()
	rc, stderr := h.run("nosync", "N5", 0, nil, "true", false)
	_ = holder.Wait()
	if rc != 0 || time.Since(t0) < 1500*time.Millisecond {
		t.Fatalf("rc %d after %s (%s); want a wait for the lock", rc, time.Since(t0), stderr)
	}
}
