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

	"github.com/eturkes/offload/internal/remote"
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

func TestLocalArgvIsNotShellParsed(t *testing.T) { // X2, scope + plain paths
	for _, scoped := range []bool{true, false} {
		t.Run(strconv.FormatBool(scoped), func(t *testing.T) {
			if !scoped {
				failingSystemdRun(t)
			}
			out := filepath.Join(t.TempDir(), "out")
			word := "$HOME `x` 'q' $$ ${HOME} %h" // ${…} + $$: systemd-run's default expansion (scope path)
			rc, err := Local(context.Background(), []string{"sh", "-c", `printf '%s' "$1" > "$2"`, "_", word, out}, nil)
			if b, _ := os.ReadFile(out); err != nil || rc != 0 || string(b) != word {
				t.Fatalf("rc %d err %v out %q", rc, err, b)
			}
		})
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

// requireScope skips where no user systemd manager starts a scope; OFFLOAD_REQUIRE_SCOPE=1 (CI with a
// lingering user manager) turns that absence into a failure, so the scope tests never skip silently.
func requireScope(t *testing.T) {
	t.Helper()
	if scopeWorks() {
		return
	}
	if os.Getenv("OFFLOAD_REQUIRE_SCOPE") != "" {
		t.Fatal("OFFLOAD_REQUIRE_SCOPE is set, but systemd-run --user --scope fails")
	}
	t.Skip("no user systemd manager: scope teardown is untestable here")
}

// scopeWorks finds the user manager the way the watcher does ($XDG_RUNTIME_DIR, else /run/user/<uid>)
// and starts a scope there.
func scopeWorks() bool {
	for _, d := range []string{os.Getenv("XDG_RUNTIME_DIR"), "/run/user/" + strconv.Itoa(os.Getuid())} {
		if fi, err := os.Stat(filepath.Join(d, "systemd", "private")); d == "" || err != nil || fi.Mode()&os.ModeSocket == 0 {
			continue
		}
		c := exec.Command("systemd-run", "--user", "--scope", "--quiet", "--collect", "--expand-environment=no", "--", "true")
		c.Env = append(os.Environ(), "XDG_RUNTIME_DIR="+d)
		return c.Run() == nil
	}
	return false
}

// ownCgroup = the test process's cgroup line: a command left out of a new scope shares it, also when
// the suite itself runs inside an offload scope.
func ownCgroup(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// inNewScope = cg names an offload scope other than the one the test may already run in.
func inNewScope(t *testing.T, cg string) bool {
	t.Helper()
	return strings.Contains(cg, "/offload-") && !strings.HasSuffix(cg, ownCgroup(t))
}

// failingSystemdRun puts a systemd-run that fails noisily first on PATH: a manager that refuses.
func failingSystemdRun(t *testing.T) {
	bin := t.TempDir()
	stub := "#!/bin/sh\necho 'systemd-run: refused' >&2\nexit 1\n"
	if err := os.WriteFile(filepath.Join(bin, "systemd-run"), []byte(stub), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+":"+os.Getenv("PATH"))
}

// daemon = a command that double-forks a setsid daemon (own session: out of pkill -s reach), writes
// the daemon's pid to the file in $1, then exits 5.
const daemon = `setsid -f bash -c 'echo $$ > "$0"; exec sleep 300' "$1"
while [ ! -s "$1" ]; do sleep 0.01; done
exit 5`

func requireGone(t *testing.T, pid int) {
	t.Helper()
	if pid <= 0 {
		t.Fatalf("bad pid %d", pid)
	}
	for deadline := time.Now().Add(3 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if !alive(pid) {
			return
		}
	}
	_ = syscall.Kill(pid, syscall.SIGKILL)
	t.Errorf("pid %d outlived the run", pid)
}

func TestScopeEndsEscapedDaemonWithLeader(t *testing.T) { // X1
	requireScope(t)
	pids := filepath.Join(t.TempDir(), "pids")
	t0 := time.Now()
	rc, err := Local(context.Background(), []string{"bash", "-c", daemon, "_", pids}, nil)
	if err != nil || rc != 5 || time.Since(t0) > 5*time.Second {
		t.Fatalf("rc %d err %v after %s; want 5 at once (TERM reaches the scope)", rc, err, time.Since(t0))
	}
	requireGone(t, readPids(t, pids, 1)[0])
}

func TestScopeEndsEscapedDaemonOnCancel(t *testing.T) { // X1 + X3
	requireScope(t)
	pids := filepath.Join(t.TempDir(), "pids")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int)
	go func() {
		rc, _ := Local(ctx, []string{"bash", "-c", strings.Replace(daemon, "exit 5", "sleep 300", 1), "_", pids}, nil)
		done <- rc
	}()
	p := readPids(t, pids, 1)
	cancel()
	select {
	case rc := <-done:
		if rc != 143 {
			t.Errorf("rc %d; want 143 (leader ended by TERM)", rc)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("watcher did not finish")
	}
	requireGone(t, p[0])
}

func TestRefusedScopeRunsCommandWithoutScope(t *testing.T) { // X1 fallback: CMD starts once, unscoped
	failingSystemdRun(t)
	dir := t.TempDir()
	rc, err := Local(context.Background(), []string{"bash", "-c", `echo run >> "$1/runs"; cat /proc/self/cgroup > "$1/cgroup"; exit 6`, "_", dir}, nil)
	runs, _ := os.ReadFile(filepath.Join(dir, "runs"))
	cg, _ := os.ReadFile(filepath.Join(dir, "cgroup"))
	if err != nil || rc != 6 || string(runs) != "run\n" || string(cg) != ownCgroup(t) {
		t.Fatalf("rc %d err %v runs %q cgroup %q", rc, err, runs, cg)
	}
}

func TestScopeHoldsCommand(t *testing.T) { // X1: CMD itself sits in the offload scope
	requireScope(t)
	out := filepath.Join(t.TempDir(), "cgroup")
	rc, err := Local(context.Background(), []string{"bash", "-c", `cat /proc/self/cgroup > "$1"`, "_", out}, nil)
	cg, _ := os.ReadFile(out)
	if err != nil || rc != 0 || !inNewScope(t, string(cg)) {
		t.Fatalf("rc %d err %v cgroup %q (test's own: %q)", rc, err, cg, ownCgroup(t))
	}
}

func TestScopeKeepsCommandEnvironment(t *testing.T) { // X1 + X2: XDG_RUNTIME_DIR + INVOCATION_ID as CMD had them
	requireScope(t)
	for _, set := range []bool{true, false} {
		out := filepath.Join(t.TempDir(), "out")
		want := "unset unset"
		if set {
			dir := filepath.Join(t.TempDir(), "elsewhere") // no manager socket here
			t.Setenv("XDG_RUNTIME_DIR", dir)
			t.Setenv("INVOCATION_ID", "callers-id")
			want = dir + " callers-id"
		} else {
			for _, k := range []string{"XDG_RUNTIME_DIR", "INVOCATION_ID"} {
				t.Setenv(k, "")
				if err := os.Unsetenv(k); err != nil {
					t.Fatal(err)
				}
			}
		}
		rc, err := Local(context.Background(), []string{"bash", "-c", `echo "${XDG_RUNTIME_DIR-unset} ${INVOCATION_ID-unset}" > "$1"; cat /proc/self/cgroup >> "$1"`, "_", out}, nil)
		b, _ := os.ReadFile(out)
		if err != nil || rc != 0 || !strings.HasPrefix(string(b), want+"\n") || !inNewScope(t, strings.TrimPrefix(string(b), want+"\n")) {
			t.Errorf("set=%v rc %d err %v out %q; want %q first, then an offload scope", set, rc, err, b, want)
		}
	}
}

func TestWatcherKeepsCallerVariables(t *testing.T) { // X2: watcher locals never reach CMD's environment
	names := []string{"p", "c", "d", "who", "rc", "deadline", "eof", "rt", "unit", "mark", "hx", "vx", "hi", "vi", "cg"}
	for _, scoped := range []bool{true, false} {
		t.Run(strconv.FormatBool(scoped), func(t *testing.T) {
			if !scoped {
				failingSystemdRun(t)
			}
			var want strings.Builder
			for _, n := range names {
				t.Setenv(n, "caller-"+n)
				want.WriteString(n + "=caller-" + n + "\n")
			}
			out := filepath.Join(t.TempDir(), "env")
			script := `for n in ` + strings.Join(names, " ") + `; do printf '%s=%s\n' "$n" "$(printenv "$n")"; done > "$1"`
			rc, err := Local(context.Background(), []string{"bash", "-c", script, "_", out}, nil)
			if b, _ := os.ReadFile(out); err != nil || rc != 0 || string(b) != want.String() {
				t.Fatalf("rc %d err %v env:\n%s", rc, err, b)
			}
		})
	}
}

func TestWatcherTermBeforeSetsidRunsNothing(t *testing.T) { // X1: TERM to the watcher = EOF, even before setsid
	sid, err := exec.LookPath("setsid")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir() // a slow setsid widens the window before the leader has a session of its own
	if err = os.WriteFile(filepath.Join(bin, "setsid"), []byte("#!/bin/sh\nsleep 1\nexec "+sid+" \"$@\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+":"+os.Getenv("PATH"))
	ran := filepath.Join(t.TempDir(), "ran")
	c := exec.Command("bash", "-c", watchScript+localScript, "offload", "touch", ran)
	in, err := c.StdinPipe() // held open: only the signal may end the watcher
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = in.Close() }()
	if err = c.Start(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)
	if err = c.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() { _ = c.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		_ = c.Process.Kill()
		t.Fatal("watcher ignored TERM")
	}
	time.Sleep(1500 * time.Millisecond)
	if exists(ran) {
		t.Fatal("CMD ran after its watcher was TERMed before setsid")
	}
}

func TestScopeKillsTermIgnoringDaemonAfterGrace(t *testing.T) { // X1: KILL at 10 s (S5 lock: e2e local-kill-escalation)
	requireScope(t)
	pids := filepath.Join(t.TempDir(), "pids")
	cmd := `setsid -f bash -c 'trap "" TERM; echo $$ > "$0"; while :; do sleep 1; done' "$1"
while [ ! -s "$1" ]; do sleep 0.01; done`
	t0 := time.Now()
	done := make(chan int)
	go func() {
		rc, _ := Local(context.Background(), []string{"bash", "-c", cmd, "_", pids}, nil)
		done <- rc
	}()
	daemonPid := readPids(t, pids, 1)[0]
	select {
	case rc := <-done:
		if rc != 0 || time.Since(t0) < 9*time.Second {
			t.Errorf("rc %d after %s; want 0 after the 10 s grace", rc, time.Since(t0))
		}
	case <-time.After(30 * time.Second):
		_ = syscall.Kill(daemonPid, syscall.SIGKILL)
		t.Fatal("watcher still waiting 30 s after the leader exited")
	}
	requireGone(t, daemonPid)
}

func TestEOFDuringScopeAdmissionRunsNothing(t *testing.T) { // X1 + X3: offload gone while systemd-run starts
	requireScope(t)
	sdRun, err := exec.LookPath("systemd-run")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	stub := "#!/bin/sh\nsleep 1\nexec " + sdRun + " \"$@\"\n"
	if err = os.WriteFile(filepath.Join(bin, "systemd-run"), []byte(stub), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+":"+os.Getenv("PATH"))
	ran := filepath.Join(t.TempDir(), "ran")
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(300*time.Millisecond, cancel)
	rc, err := Local(ctx, []string{"bash", "-c", `touch "$1"`, "_", ran}, nil)
	time.Sleep(2 * time.Second)
	if err != nil || rc != 143 || exists(ran) {
		t.Fatalf("rc %d err %v; CMD ran = %v (want 143, never ran)", rc, err, exists(ran))
	}
}

func exists(path string) bool { _, err := os.Lstat(path); return err == nil }

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

// run executes the ssh command line (boot → run script) with stdin = offload's pipe: the run script,
// the framed list, then held open until the script exits (as offload does); eof = close right after
// the list (offload died mid-list).
func (h host) run(mode, nonce string, count int, list []byte, cmd string, eof bool) (int, string) {
	h.t.Helper()
	return h.boot(boot, append([]byte(runScript), list...), mode, nonce, count, cmd, eof)
}

// boot runs script (a boot variant) as the remote login shell would, with the given stdin.
func (h host) boot(script string, stdin []byte, mode, nonce string, count int, cmd string, eof bool) (int, string) {
	h.t.Helper()
	line := remote.Join([]string{
		"bash", "-c", script, "offload", strconv.Itoa(len(runScript)), runSum,
		"offload/w", "offload/.meta/w", ".", mode, nonce, strconv.Itoa(count), cmd,
	})
	c := exec.Command("bash", "-c", line)
	c.Dir = h.home
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

// frame = NONCE, an empty EXCLUDES record, then paths — git mode's list after the run script.
func frame(nonce string, paths ...string) []byte {
	return []byte(nonce + "\x00\x00" + strings.Join(append(paths, ""), "\x00"))
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

func (h host) runFile() string { return h.path("offload/.meta/w/run-" + runSum[:16] + ".bash") }

func TestBootRefusesShortOrCorruptScript(t *testing.T) { // X8: nothing runs from a bad copy
	corrupt := []byte(runScript)
	corrupt[len(corrupt)/2] ^= 1
	for name, stdin := range map[string][]byte{
		"short": []byte(runScript[:len(runScript)-40]), "corrupt": corrupt, "empty": nil,
	} {
		t.Run(name, func(t *testing.T) {
			h := newHost(t)
			rc, stderr := h.boot(boot, stdin, "nosync", "N", 0, "touch ran", true)
			if rc != 125 || !strings.Contains(stderr, "cannot install the run script") || h.exists("offload/w/ran") || h.exists("offload/.meta/w/run-"+runSum[:16]+".bash") {
				t.Fatalf("rc %d stderr %q ran %v", rc, stderr, h.exists("offload/w/ran"))
			}
			if left, _ := filepath.Glob(h.path("offload/.meta/w/run-*")); len(left) != 0 {
				t.Errorf("temp files left: %v", left)
			}
		})
	}
}

func TestBootReplacesWrongScriptFile(t *testing.T) { // X8: the verified copy wins over whatever sits there
	h := newHost(t)
	h.write("offload/.meta/w/run-"+runSum[:16]+".bash", "touch stale-ran; exit 99\n")
	rc, stderr := h.run("nosync", "N", 0, nil, "exit 7", false)
	b, _ := os.ReadFile(h.runFile())
	if rc != 7 || h.exists("offload/w/stale-ran") || string(b) != runScript {
		t.Fatalf("rc %d stderr %q stale ran %v", rc, stderr, h.exists("offload/w/stale-ran"))
	}
}

func TestBootFailureNeverRunsAStaleScript(t *testing.T) { // X8: install failure → 125, old file untouched + unrun
	h := newHost(t)
	h.write("offload/.meta/w/run-"+runSum[:16]+".bash", "touch stale-ran\n")
	h.write("offload/w/keep", "")
	if err := os.Chmod(h.path("offload/.meta/w"), 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(h.path("offload/.meta/w"), 0o700) })
	rc, stderr := h.run("nosync", "N", 0, nil, "touch ran", false)
	if rc != 125 || h.exists("stale-ran") || h.exists("offload/w/stale-ran") || h.exists("offload/w/ran") {
		t.Fatalf("rc %d stderr %q", rc, stderr)
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
	waited := time.Since(t0) // before joining the holder: its own 2 s must not count
	_ = holder.Wait()
	if rc != 0 || waited < 1500*time.Millisecond {
		t.Fatalf("rc %d after %s (%s); want a wait for the lock", rc, waited, stderr)
	}
}
