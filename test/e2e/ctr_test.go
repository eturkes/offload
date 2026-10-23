package e2e_test

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// Real-podman tests share one isolated storage (vfs under a temp dir): the user's own images stay
// untouched, and the `FROM scratch` test image needs no registry. OFFLOAD_REQUIRE_PODMAN=1 (CI)
// turns a missing podman into a failure.
var (
	storageOnce sync.Once
	storageConf string
	storageErr  error // podman unusable here: skip (fail under OFFLOAD_REQUIRE_PODMAN)
	prebuildErr error // offload could not build the shared image: always a failure
	storageDir  string
)

func requirePodman(t *testing.T) string {
	t.Helper()
	storageOnce.Do(func() {
		if _, storageErr = exec.LookPath("podman"); storageErr != nil {
			return
		}
		if storageDir, storageErr = os.MkdirTemp("", "offload-e2e-podman-"); storageErr != nil {
			return
		}
		storageConf = filepath.Join(storageDir, "storage.conf")
		// graphroot alone: podman 5+ fails an ID-mapped layer copy on the deprecation warning when
		// rootless_storage_path is also set.
		if storageErr = os.WriteFile(storageConf, fmt.Appendf(nil, "[storage]\ndriver = \"vfs\"\ngraphroot = %q\nrunroot = %q\n",
			filepath.Join(storageDir, "graph"), filepath.Join(storageDir, "run")), 0o600); storageErr != nil {
			return
		}
		// The shared image, built once up front: parallel tests then time their own work alone.
		project := filepath.Join(storageDir, "project")
		if prebuildErr = writeDefinition(project, ""); prebuildErr != nil {
			return
		}
		build := exec.Command(binaryPath, "ctr", "--layers")
		build.Dir = project
		build.Env = append(os.Environ(), "CONTAINERS_STORAGE_CONF="+storageConf, "XDG_RUNTIME_DIR="+storageDir,
			"DBUS_SESSION_BUS_ADDRESS="+busAddress())
		if out, err := build.CombinedOutput(); err != nil {
			prebuildErr = fmt.Errorf("offload ctr --layers (shared test image): %w: %s", err, out)
		}
	})
	if storageErr != nil {
		if os.Getenv("OFFLOAD_REQUIRE_PODMAN") != "" {
			t.Fatalf("OFFLOAD_REQUIRE_PODMAN is set, but podman is unusable: %v", storageErr)
		}
		t.Skipf("no podman: image runs are untestable here (%v)", storageErr)
	}
	if prebuildErr != nil {
		t.Fatal(prebuildErr)
	}
	return storageConf
}

// podmanBus points podman at the real user bus: a sandbox's own XDG_RUNTIME_DIR hides it, and
// without one podman warns + falls back to cgroupfs on every call (CI sets no bus address).
func podmanBus(s *sandbox) {
	if bus := busAddress(); bus != "" {
		s.setenv("DBUS_SESSION_BUS_ADDRESS", bus)
	}
}

func busAddress() string {
	if bus := os.Getenv("DBUS_SESSION_BUS_ADDRESS"); bus != "" {
		return bus
	}
	path := filepath.Join("/run/user", strconv.Itoa(os.Getuid()), "bus")
	if rt := os.Getenv("XDG_RUNTIME_DIR"); rt != "" {
		path = filepath.Join(rt, "bus")
	}
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return "unix:path=" + path
}

// removeStorage deletes the shared storage; its files belong to subordinate ids ⇒ podman unshare.
func removeStorage() error {
	if storageDir == "" {
		return nil
	}
	rm := exec.Command("podman", "unshare", "rm", "-rf", storageDir)
	rm.Env = append(os.Environ(), "CONTAINERS_STORAGE_CONF="+storageConf)
	if out, err := rm.CombinedOutput(); err != nil {
		return fmt.Errorf("remove podman test storage: %w: %s", err, out)
	}
	// podman re-creates the store's lock files (ours, not subordinate) as it exits.
	return os.RemoveAll(storageDir)
}

const testImage = "FROM scratch\nENV IMAGE_VAR=from-image\nCOPY ctrhelper /usr/bin/ctrhelper\n"

// writeDefinition = the e2e image's context under root (extra = more Containerfile lines), with
// the modes ctrSandbox's sandboxes get, so equal extras mean equal tags (K4).
func writeDefinition(root, extra string) error {
	dir := filepath.Join(root, ".devcontainer")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	helper, err := os.ReadFile(helperPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "Containerfile"), []byte(testImage+extra), 0o600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "ctrhelper"), helper, 0o755)
}

// ctrSandbox = a sandbox whose project defines the e2e image (extra = more Containerfile lines,
// for a distinct tag), with the switch on.
func ctrSandbox(t *testing.T, git bool, extra string) *sandbox {
	t.Helper()
	conf := requirePodman(t)
	s := newSandbox(t, git)
	must(t, writeDefinition(s.root, extra))
	s.setenv("CONTAINERS_STORAGE_CONF", conf)
	s.setenv("XDG_CACHE_HOME", filepath.Join(s.base, "cache"))
	s.setenv("OFFLOAD_CTR", "1")
	podmanBus(s)
	// Runs before TempDir's removal (LIFO): a container that outlives offload (killed podman, X1)
	// is cleaned up by conmon, whose crun recreates <runtime>/crun after a too-early RemoveAll.
	t.Cleanup(func() {
		for deadline := time.Now().Add(10 * time.Second); ; time.Sleep(100 * time.Millisecond) {
			r := podman(t, s, "ps", "-a", "--format", "{{.Mounts}}")
			if r.code == 0 && !strings.Contains(r.stdout, s.root) {
				return
			}
			if time.Now().After(deadline) {
				t.Errorf("a container of %s still exists 10 s after the test: %q %s", s.root, r.stdout, r.stderr)
				return
			}
		}
	})
	return s
}

func (s *sandbox) ctr(argv ...string) result {
	s.t.Helper()
	return s.start(append([]string{"ctr", "--"}, argv...)...).wait(60 * time.Second)
}

func TestK1CtrUsage(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	for _, args := range [][]string{{"ctr"}, {"ctr", "--"}, {"ctr", "--layers", "--check"}, {"ctr", "--layers", "x"}, {"ctr", "--check", "--", "x"}, {"ctr", "--bogus", "x"}} {
		r := s.run(args...)
		if r.code != 2 || !strings.Contains(r.stderr, "Usage:") {
			t.Errorf("%q: exit %d, stderr %q; want 2 + usage", args, r.code, r.stderr)
		}
	}
	r := s.run("ctr", "--help")
	r.requireCode(t, 0)
	requireContains(t, r.stdout, "offload ctr [--] COMMAND", "--layers", "--check", "OFFLOAD_CTR=1")
}

func TestK2SwitchOffExecsInPlace(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false) // no definition, no podman needed
	s.write(s.path("bin/podman"), "#!/bin/sh\necho podman-called >&2\nexit 99\n")
	must(t, os.Chmod(s.path("bin/podman"), 0o755))
	s.setenv("PATH", s.path("bin")+":"+filepath.Join(s.base, "bin")+":"+os.Getenv("PATH"))
	for _, v := range []string{"", "0"} {
		s.setenv("OFFLOAD_CTR", v)
		p := s.start("ctr", "--", "bash", "-c", `echo "$$ $HEAVY_SLOTS"; exit 9`)
		r := p.wait(10 * time.Second)
		r.requireCode(t, 9)
		if want := fmt.Sprintf("%d 2\n", p.cmd.Process.Pid); r.stdout != want || r.stderr != "" {
			t.Errorf("OFFLOAD_CTR=%q: stdout %q (want %q: same pid, env kept), stderr %q", v, r.stdout, want, r.stderr)
		}
	}
	s.write("plain-file", "x")
	r := s.run("ctr", "--", "./plain-file")
	r.requireCode(t, 126)
	r = s.run("ctr", "--", "no-such-command-k2")
	r.requireCode(t, 127)
}

func TestK2BadSwitchValue(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	s.setenv("OFFLOAD_CTR", "yes")
	r := s.run("ctr", "--", "true")
	r.requireCode(t, 2)
	requireContains(t, r.stderr, "OFFLOAD_CTR")
}

func TestK3MissingDefinitionIsUsageError(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.setenv("OFFLOAD_CTR", "1")
	r := s.run("ctr", "--", "true")
	r.requireCode(t, 2)
	requireContains(t, r.stderr, filepath.Join(s.root, ".devcontainer/Containerfile"))
	r = s.run("ctr", "--layers")
	r.requireCode(t, 2)
}

func TestK5BuildsOnceThenReusesTag(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, true, "ENV K5=build-once\n")
	r := s.ctr("ctrhelper", "exit", "0")
	r.requireCode(t, 0)
	build := regexp.MustCompile(`offload: building localhost/offload-ctr:[0-9a-f]{16} from ` + regexp.QuoteMeta(s.path(".devcontainer/Containerfile")))
	if !build.MatchString(r.stderr) {
		t.Fatalf("first run: stderr %q lacks the build line", r.stderr)
	}
	r = s.ctr("ctrhelper", "exit", "0")
	r.requireCode(t, 0)
	if strings.Contains(r.stderr, "building") {
		t.Fatalf("second run rebuilt: %q", r.stderr)
	}
	s.write(".devcontainer/Containerfile", testImage+"ENV K5=changed\n") // K4: new content → new tag
	r = s.ctr("ctrhelper", "env", "K5")
	r.requireCode(t, 0)
	if !build.MatchString(r.stderr) || r.stdout != "K5=changed\n" {
		t.Fatalf("after a definition change: stdout %q stderr %q", r.stdout, r.stderr)
	}
}

func TestK5ConcurrentFirstRunsBuildOnce(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, true, "ENV K5=concurrent\n")
	runs := []*running{s.start("ctr", "--", "ctrhelper", "exit", "0"), s.start("ctr", "--", "ctrhelper", "exit", "0")}
	builds := 0
	for _, p := range runs {
		r := p.wait(60 * time.Second)
		r.requireCode(t, 0)
		builds += strings.Count(r.stderr, "offload: building")
	}
	if builds != 1 {
		t.Fatalf("%d builds for one tag, want 1", builds)
	}
}

func TestK5SignalWhileWaitingForBuildLock(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, true, "ENV K5=lock-wait\n")
	tag := tagOf(t, s)
	lockPath := filepath.Join(s.runtime, "offload", "ctr", strings.TrimPrefix(tag, "localhost/offload-ctr:")+".lock")
	must(t, os.MkdirAll(filepath.Dir(lockPath), 0o700))
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_WRONLY, 0o600)
	must(t, err)
	defer func() { _ = lock.Close() }()
	must(t, syscall.Flock(int(lock.Fd()), syscall.LOCK_EX))
	p := s.start("ctr", "--", "ctrhelper", "exit", "0")
	p.await("the build-lock wait", func() bool { return holdsOpen(p.cmd.Process.Pid, lockPath) })
	p.signal(syscall.SIGTERM)
	r := p.wait(10 * time.Second)
	r.requireCode(t, 143)
	if strings.Contains(r.stderr, "building") {
		t.Fatalf("built while another holder had the lock: %q", r.stderr)
	}
	if inspect := podman(t, s, "image", "exists", tag); inspect.code == 0 {
		t.Fatal("image tagged although the run was stopped before its build")
	}
}

// holdsOpen = pid has path open (it is polling that lock).
func holdsOpen(pid int, path string) bool {
	fds, _ := filepath.Glob(fmt.Sprintf("/proc/%d/fd/*", pid))
	for _, fd := range fds {
		if target, err := os.Readlink(fd); err == nil && target == path {
			return true
		}
	}
	return false
}

// tagOf = the tag K4 gives the sandbox's context: the tag printed by --layers' build line.
func tagOf(t *testing.T, s *sandbox) string {
	t.Helper()
	probe := ctrSandbox(t, true, strings.TrimPrefix(s.read(".devcontainer/Containerfile"), testImage))
	r := probe.run("ctr", "--layers")
	r.requireCode(t, 0)
	m := regexp.MustCompile(`localhost/offload-ctr:[0-9a-f]{16}`).FindString(r.stderr)
	if m == "" {
		t.Fatalf("no build line: %q", r.stderr)
	}
	rmi := podman(t, probe, "rmi", m)
	if rmi.code != 0 {
		t.Fatalf("rmi %s: %q", m, rmi.stderr)
	}
	return m
}

func podman(t *testing.T, s *sandbox, args ...string) result {
	t.Helper()
	cmd := exec.Command("podman", args...)
	cmd.Env = s.env
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	r := result{stdout: out.String(), stderr: errOut.String()}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		r.code = ee.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestK6MountsHomeAndOwnership(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, true, "")
	must(t, os.MkdirAll(s.path("sub/dir"), 0o700))
	link := filepath.Join(s.base, "link-to-sub")
	must(t, os.Symlink(s.path("sub/dir"), link))
	s.cwd = link // K3: -w = the resolved cwd
	r := s.ctr("ctrhelper", "pwd")
	r.requireCode(t, 0)
	if r.stdout != s.path("sub/dir")+"\n" {
		t.Fatalf("pwd %q, want %q", r.stdout, s.path("sub/dir"))
	}
	r = s.ctr("ctrhelper", "write", "made-here", "content")
	r.requireCode(t, 0)
	fi, err := os.Stat(s.path("sub/dir/made-here"))
	must(t, err)
	if st := fi.Sys().(*syscall.Stat_t); int(st.Uid) != os.Getuid() || s.read("sub/dir/made-here") != "content" {
		t.Fatalf("file owner %d (want %d), content %q", st.Uid, os.Getuid(), s.read("sub/dir/made-here"))
	}
	s.setenv("http_proxy", "http://proxy.invalid:3128") // podman passes proxy variables by default
	r = s.ctr("ctrhelper", "env", "HOME", "IMAGE_VAR", "HEAVY_SLOTS", "http_proxy")
	r.requireCode(t, 0)
	if r.stdout != "HOME=/.offload/home\nIMAGE_VAR=from-image\nHEAVY_SLOTS unset\nhttp_proxy unset\n" {
		t.Fatalf("env %q: HOME fixed, image ENV applies, host env stays out", r.stdout)
	}
	r = s.ctr("ctrhelper", "write", "/.offload/home/kept", "across runs")
	r.requireCode(t, 0)
	r = s.ctr("ctrhelper", "read", "/.offload/home/kept")
	r.requireCode(t, 0)
	sum := sha1.Sum([]byte(s.root))
	home := filepath.Join(s.base, "cache", "offload", "ctr", fmt.Sprintf("%s-%x", filepath.Base(s.root), sum[:4]), "home")
	if r.stdout != "across runs" || s.read(filepath.Join(home, "kept")) != "across runs" {
		t.Fatalf("HOME not kept at %s: %q", home, r.stdout)
	}
}

func TestK6WorktreeCommonDirMounted(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, true, "")
	s.git("add", ".devcontainer")
	s.git("commit", "-qm", "image")
	wt := filepath.Join(s.base, "wt")
	s.git("worktree", "add", "-q", wt)
	s.cwd = wt
	r := s.ctr("ctrhelper", "exists", filepath.Join(s.root, ".git", "HEAD"))
	r.requireCode(t, 0)
	r = s.ctr("ctrhelper", "exists", filepath.Join(s.root, "tracked")) // the main tree itself stays out
	r.requireCode(t, 1)
}

func TestK6StreamsAndStdinByteExact(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	r := s.ctr("ctrhelper", "bytes")
	r.requireCode(t, 0)
	if r.stdout != "out\x00\xff\r" || !strings.HasSuffix(r.stderr, "err\x01\xfe") {
		t.Fatalf("stdout %q stderr %q", r.stdout, r.stderr)
	}
	cmd := exec.Command(binaryPath, "ctr", "--", "ctrhelper", "cat")
	cmd.Env, cmd.Dir = s.env, s.cwd
	cmd.Stdin = strings.NewReader("in\x00put")
	out, err := cmd.Output()
	if err != nil || string(out) != "in\x00put" {
		t.Fatalf("stdin passthrough: %q, %v", out, err)
	}
}

func TestK7StatusPassesThrough(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	for _, code := range []int{0, 2, 7, 125, 126, 127, 255} {
		r := s.ctr("ctrhelper", "exit", strconv.Itoa(code))
		r.requireCode(t, code)
	}
	r := s.ctr("ctrhelper", "selfkill", "15")
	r.requireCode(t, 143)
	r = s.ctr("no-such-command-k7")
	r.requireCode(t, 127)
	requireContains(t, r.stderr, "no-such-command-k7: command not found")
	s.write("not-executable", "x")
	r = s.ctr("./not-executable")
	r.requireCode(t, 126)
}

func TestK7NoStatusRecordIs125(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	podmanPath, err := exec.LookPath("podman")
	must(t, err)
	// podman builds as usual but its `run` dies before the container starts.
	s.write(s.path("bin/podman"), "#!/bin/sh\n[ \"$1\" = run ] && { echo 'Error: injected' >&2; exit 125; }\nexec "+podmanPath+" \"$@\"\n")
	must(t, os.Chmod(s.path("bin/podman"), 0o755))
	s.setenv("PATH", s.path("bin")+":"+filepath.Join(s.base, "bin")+":"+os.Getenv("PATH"))
	r := s.ctr("ctrhelper", "exit", "0")
	r.requireCode(t, 125)
	requireContains(t, r.stderr, "Error: injected", "offload: the container ended without its command's status")
	git, err := exec.LookPath("git")
	must(t, err)
	must(t, os.MkdirAll(s.path("gitonly"), 0o700))
	must(t, os.Symlink(git, s.path("gitonly/git")))
	s.setenv("PATH", filepath.Join(s.base, "bin")+":"+s.path("gitonly"))
	r = s.ctr("ctrhelper", "exit", "0")
	r.requireCode(t, 125)
	requireContains(t, r.stderr, "podman")
}

// trapRun starts `ctrhelper trap` and waits until it logs ready; the log lives in the project.
func trapRun(t *testing.T, s *sandbox) (*running, string) {
	t.Helper()
	log := s.path("trap.log")
	p := s.start("ctr", "--", "ctrhelper", "trap", log, "3", "5")
	p.await("trap ready", func() bool { return strings.HasPrefix(readOr(log), "ready\n") })
	return p, log
}

func readOr(path string) string {
	b, _ := os.ReadFile(path)
	return string(b)
}

func TestK8SignalReachesCommandOnce(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		sig  syscall.Signal
		want string
	}{{"term", syscall.SIGTERM, "TERMINATED"}, {"int", syscall.SIGINT, "INTERRUPT"}, {"hup", syscall.SIGHUP, "HANGUP"}} {
		t.Run(tc.name, func(t *testing.T) {
			s := ctrSandbox(t, false, "")
			p, log := trapRun(t, s)
			p.signal(tc.sig)
			r := p.wait(20 * time.Second)
			r.requireCode(t, 5) // the command's own status after it handled the signal
			if got := readOr(log); got != "ready\n"+tc.want+"\n" {
				t.Fatalf("trap log %q, want one %s", got, tc.want)
			}
		})
	}
}

func TestK8GroupWideSignalStillOnce(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	p, log := trapRun(t, s)
	// offload + podman (its own group, offload's session), as X1's `pkill -s` reaches both.
	kids, err := exec.Command("pgrep", "-P", strconv.Itoa(p.cmd.Process.Pid)).Output()
	must(t, err)
	for _, pid := range append(strings.Fields(string(kids)), strconv.Itoa(p.cmd.Process.Pid)) {
		n, err := strconv.Atoi(pid)
		must(t, err)
		must(t, syscall.Kill(n, syscall.SIGTERM))
	}
	r := p.wait(20 * time.Second)
	r.requireCode(t, 5)
	if got := readOr(log); got != "ready\nTERMINATED\n" {
		t.Fatalf("trap log %q, want exactly one TERM", got)
	}
}

func markerGone(t *testing.T, marker string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		err := exec.Command("pgrep", "-f", "ctrhelper ignore "+marker).Run()
		var ee *exec.ExitError
		switch {
		case errors.As(err, &ee) && ee.ExitCode() == 1: // pgrep: no process matched
			return
		case err != nil: // pgrep itself failed: no evidence either way
			t.Fatalf("pgrep for %s: %v", marker, err)
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("processes tagged %s still run after %s", marker, timeout)
}

func TestK8SigkillEndsContainer(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	marker := fmt.Sprintf("k8kill-%d", time.Now().UnixNano())
	ready := s.path("ready")
	p := s.start("ctr", "--", "ctrhelper", "ignore", marker, ready)
	p.await("container start", func() bool { return exists(ready) })
	p.signal(syscall.SIGKILL)
	markerGone(t, marker, 5*time.Second)
	// --rm still applied: no container record left behind.
	deadline := time.Now().Add(5 * time.Second)
	for {
		r := podman(t, s, "ps", "-a", "--no-trunc", "--format", "{{.ID}} {{.Command}}")
		if r.code != 0 {
			t.Fatalf("podman ps: %d %s", r.code, r.stderr)
		}
		if !strings.Contains(r.stdout, marker) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("container records left: %q", r.stdout)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestK8LeftoverProcessesEndAfterCommand(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	marker := fmt.Sprintf("k8left-%d", time.Now().UnixNano())
	start := time.Now()
	r := s.ctr("ctrhelper", "spawn", marker, "6") // child ignores TERM ⇒ KILL at the 10 s grace
	r.requireCode(t, 6)
	if took := time.Since(start); took < 9*time.Second || took > 20*time.Second {
		t.Fatalf("returned after %s, want the 10 s grace", took)
	}
	markerGone(t, marker, 2*time.Second)
}

func TestK8UnderOffloadRunKillEndsContainer(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	marker := fmt.Sprintf("k8run-%d", time.Now().UnixNano())
	ready := s.path("ready")
	p := s.start("run", "--local", "--", binaryPath, "ctr", "--", "ctrhelper", "ignore", marker, ready)
	p.await("container start", func() bool { return exists(ready) })
	start := time.Now()
	p.signal(syscall.SIGKILL) // X1: watcher TERMs offload ctr (forwarded, ignored), KILLs at 10 s ⇒ liveness EOF
	markerGone(t, marker, 15*time.Second)
	if took := time.Since(start); took < 9*time.Second {
		t.Fatalf("container gone after %s although its command ignores TERM: X1's KILL comes at 10 s", took)
	}
}

func TestK9LayersMatchPodman(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "ENV K9=layers\n")
	r := s.run("ctr", "--layers")
	r.requireCode(t, 0)
	tag := regexp.MustCompile(`localhost/offload-ctr:[0-9a-f]{16}`).FindString(r.stderr)
	inspect := podman(t, s, "image", "inspect", "--format", "{{range .RootFS.Layers}}{{println .}}{{end}}", tag)
	if tag == "" || inspect.code != 0 || r.stdout != strings.TrimRight(inspect.stdout, "\n")+"\n" || !strings.HasPrefix(r.stdout, "sha256:") {
		t.Fatalf("--layers %q vs inspect %q (tag %q)", r.stdout, inspect.stdout, tag)
	}
}

func TestK10CheckComparesHosts(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, true, "")
	s.git("add", ".devcontainer")
	s.git("commit", "-qm", "image")
	// The fake host runs the installed `offload` from PATH: this very binary.
	bin, err := os.ReadFile(binaryPath)
	must(t, err)
	must(t, os.WriteFile(filepath.Join(s.base, "bin", "offload"), bin, 0o755))
	r := s.start("ctr", "--check").wait(90 * time.Second)
	r.requireCode(t, 0)
	if !regexp.MustCompile(`^alpha: same \([0-9]+ layers\)\n$`).MatchString(r.stdout) {
		t.Fatalf("stdout %q", r.stdout)
	}
	must(t, os.WriteFile(filepath.Join(s.base, "bin", "offload"), append(bin, 0), 0o755))
	r = s.start("ctr", "--check").wait(90 * time.Second)
	r.requireCode(t, 1)
	requireContains(t, r.stdout, "alpha: same (", "; offload binary differs")
	s.write(filepath.Join(s.base, "bin", "offload"), "#!/bin/sh\necho sha256:00\n")
	must(t, os.Chmod(filepath.Join(s.base, "bin", "offload"), 0o755))
	r = s.start("ctr", "--check").wait(90 * time.Second)
	r.requireCode(t, 1)
	requireContains(t, r.stdout, "alpha: different (here ", "there 1 layers)")
}

func TestK10CheckReportsDownHost(t *testing.T) {
	t.Parallel()
	requirePodman(t)
	s := newSandbox(t, true, host{name: "gone", down: true})
	s.write(".devcontainer/Containerfile", testImage)
	helper, err := os.ReadFile(helperPath)
	must(t, err)
	must(t, os.WriteFile(s.path(".devcontainer/ctrhelper"), helper, 0o755))
	s.setenv("CONTAINERS_STORAGE_CONF", storageConf)
	podmanBus(s)
	r := s.start("ctr", "--check").wait(60 * time.Second)
	r.requireCode(t, 1)
	requireContains(t, r.stdout, "gone: failed (down: ")
}

func TestK1HelpWinsOverBadSwitch(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	s.setenv("OFFLOAD_CTR", "bogus")
	r := s.run("ctr", "--help")
	r.requireCode(t, 0)
}

func TestK3DefinitionMustBeRegularFile(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.setenv("OFFLOAD_CTR", "1")
	s.write("real-Containerfile", testImage)
	must(t, os.MkdirAll(s.path(".devcontainer"), 0o700))
	must(t, os.Symlink(s.path("real-Containerfile"), s.path(".devcontainer/Containerfile")))
	r := s.run("ctr", "--", "true")
	r.requireCode(t, 2)
	requireContains(t, r.stderr, "no symlinks")
}

func TestK3DamagedRepoIs125(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	s.setenv("OFFLOAD_CTR", "1")
	must(t, os.Mkdir(s.path(".git"), 0o700)) // git: "not a git repository", yet a .git is there (Y1)
	s.write(".devcontainer/Containerfile", testImage)
	r := s.run("ctr", "--", "true")
	r.requireCode(t, 125)
	requireContains(t, r.stderr, "git does not recognize the repository")
}

func TestK3ColonInPathIs125(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	dir := filepath.Join(s.base, "a:b")
	must(t, os.MkdirAll(filepath.Join(dir, ".devcontainer"), 0o700))
	for _, f := range []string{"Containerfile", "ctrhelper"} {
		data, err := os.ReadFile(s.path(".devcontainer/" + f))
		must(t, err)
		must(t, os.WriteFile(filepath.Join(dir, ".devcontainer", f), data, 0o755))
	}
	s.cwd = dir
	r := s.ctr("ctrhelper", "exit", "0")
	r.requireCode(t, 125)
	requireContains(t, r.stderr, "holds ':'")
}

// fakePodman puts a podman wrapper first on PATH: script runs for the subcommands case names, the
// real podman for the rest.
func fakePodman(t *testing.T, s *sandbox, cases string) {
	t.Helper()
	podmanPath, err := exec.LookPath("podman")
	must(t, err)
	dir := filepath.Join(s.base, "fakepodman")
	must(t, os.MkdirAll(dir, 0o700))
	must(t, os.WriteFile(filepath.Join(dir, "podman"), []byte("#!/bin/bash\ncase \"$1 $2\" in\n"+cases+"\nesac\nexec "+podmanPath+" \"$@\"\n"), 0o755))
	s.setenv("PATH", dir+":"+filepath.Join(s.base, "bin")+":"+os.Getenv("PATH"))
}

func TestK5ImageExistsFailureIs125(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "ENV K5=exists-fails\n")
	fakePodman(t, s, `"image exists") echo 'Error: storage broken' >&2; exit 125 ;;
build*) echo build-called >&2 ;;`)
	r := s.ctr("ctrhelper", "exit", "0")
	r.requireCode(t, 125)
	requireContains(t, r.stderr, "podman image exists", "storage broken")
	if strings.Contains(r.stderr, "build-called") {
		t.Fatal("an exists failure counted as a cache miss")
	}
}

func TestK5BuildKeepsLockPastOffloadDeath(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "ENV K5=lock-inherit\n")
	started := filepath.Join(s.base, "build-started")
	fakePodman(t, s, `"build --timestamp") echo $$ > `+started+`.tmp; mv `+started+`.tmp `+started+`; exec sleep 30 ;;`)
	p := s.start("ctr", "--", "ctrhelper", "exit", "0")
	p.await("the build", func() bool { return exists(started) })
	p.signal(syscall.SIGKILL)
	p.wait(5 * time.Second)
	locks, err := filepath.Glob(filepath.Join(s.runtime, "offload", "ctr", "*.lock"))
	must(t, err)
	if len(locks) != 1 {
		t.Fatalf("lock files %q", locks)
	}
	f, err := os.OpenFile(locks[0], os.O_WRONLY, 0)
	must(t, err)
	defer func() { _ = f.Close() }()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err == nil {
		t.Fatal("build lock free while the orphaned build still runs")
	}
	must(t, syscall.Kill(readPID(t, started), syscall.SIGKILL)) // end the fake build
	deadline := time.Now().Add(5 * time.Second)
	for syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) != nil {
		if time.Now().After(deadline) {
			t.Fatal("lock still held after the build ended")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// TestK8SlotFreesAfterContainerLastWrite = X1/S5's barrier through the container: once the outer run
// dies, its heavy slot frees only after every container process stopped writing.
func TestK8SlotFreesAfterContainerLastWrite(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	s.setenv("HEAVY_SLOTS", "1")
	ticks := s.path("ticks")
	ready := s.path("ready")
	p := s.start("run", "--local", "--", binaryPath, "ctr", "--", "ctrhelper", "tick", ticks, ready)
	p.await("container start", func() bool { return exists(ready) })
	slot, err := os.OpenFile(filepath.Join(s.runtime, "heavy", "slot1"), os.O_WRONLY, 0)
	must(t, err)
	defer func() { _ = slot.Close() }()
	p.signal(syscall.SIGKILL)
	deadline := time.Now().Add(20 * time.Second)
	for syscall.Flock(int(slot.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) != nil {
		if time.Now().After(deadline) {
			t.Fatal("slot still held 20 s after the kill")
		}
		time.Sleep(time.Millisecond)
	}
	freed := time.Now().UnixNano()
	time.Sleep(300 * time.Millisecond) // a straggler would tick again by now
	lines := strings.Fields(readOr(ticks))
	last, err := strconv.ParseInt(lines[len(lines)-1], 10, 64)
	must(t, err)
	if last >= freed {
		t.Fatalf("container wrote at %d, after the slot freed at %d", last, freed)
	}
	t.Logf("slot freed %s after the container's last write", time.Duration(freed-last))
}

func TestK10CheckIgnoresLoginBannerAndReportsMissingOffload(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, true, "")
	s.git("add", ".devcontainer")
	s.git("commit", "-qm", "image")
	bin, err := os.ReadFile(binaryPath)
	must(t, err)
	must(t, os.WriteFile(filepath.Join(s.base, "bin", "offload"), bin, 0o755))
	// A login shell's own output, even imitating the frame, must not count (nonce-framed lines, K10).
	s.write(filepath.Join(s.hosts, "alpha", ".bash_profile"), "echo login-banner\necho layer sha256:fake\necho offload-check x 00\n")
	r := s.start("ctr", "--check").wait(90 * time.Second)
	r.requireCode(t, 0)
	if !regexp.MustCompile(`^alpha: same \([0-9]+ layers\)\n$`).MatchString(r.stdout) {
		t.Fatalf("banner leaked into the comparison: %q", r.stdout)
	}
	must(t, os.Remove(filepath.Join(s.hosts, "alpha", ".bash_profile")))
	must(t, os.Remove(filepath.Join(s.base, "bin", "offload")))
	s.setenv("PATH", filepath.Join(s.base, "bin")+":"+pathWithout("offload")) // an installed offload (nixbox) must not leak in
	r = s.start("ctr", "--check").wait(90 * time.Second)
	r.requireCode(t, 1)
	requireContains(t, r.stdout, "alpha: failed (rc ")
}

// pathWithout = $PATH minus every directory holding name.
func pathWithout(name string) string {
	dirs := []string{}
	for _, d := range filepath.SplitList(os.Getenv("PATH")) {
		if _, err := os.Stat(filepath.Join(d, name)); err != nil {
			dirs = append(dirs, d)
		}
	}
	return strings.Join(dirs, string(filepath.ListSeparator))
}

func TestK6RemoteWorktreeLayout(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, true, "")
	s.git("add", ".devcontainer")
	s.git("commit", "-qm", "image")
	bin, err := os.ReadFile(binaryPath)
	must(t, err)
	must(t, os.WriteFile(filepath.Join(s.base, "bin", "offload"), bin, 0o755))
	wt := filepath.Join(s.base, "wt")
	s.git("worktree", "add", "-q", wt)
	s.cwd = wt
	sum := sha1.Sum([]byte(wt))
	repo := filepath.Join(s.hosts, "alpha", "offload", ".meta", fmt.Sprintf("wt-%x", sum[:4]), "repo")
	r := s.start("run", "--host", "alpha", "--", "offload", "ctr", "--", "ctrhelper", "exists", filepath.Join(repo, "HEAD")).wait(60 * time.Second)
	r.requireCode(t, 0)
}

func TestK3DefinitionDirMustBeRealDirectory(t *testing.T) {
	t.Parallel()
	conf := requirePodman(t) // a regression would build: keep it in the isolated storage
	s := newSandbox(t, true)
	s.setenv("OFFLOAD_CTR", "1")
	s.setenv("CONTAINERS_STORAGE_CONF", conf)
	podmanBus(s)
	must(t, writeDefinition(filepath.Join(s.base, "elsewhere"), ""))
	// Tag's walk never enters a symlinked root ⇒ every change behind it would keep the tag.
	must(t, os.Symlink(filepath.Join(s.base, "elsewhere", ".devcontainer"), s.path(".devcontainer")))
	r := s.run("ctr", "--", "true")
	r.requireCode(t, 2)
	requireContains(t, r.stderr, ".devcontainer")
}

func TestK2PathNonExecutableIs126(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	must(t, os.MkdirAll(s.path("pathdir"), 0o700))
	s.write("pathdir/notexec", "#!/bin/sh\n")
	s.setenv("PATH", s.path("pathdir")+":"+filepath.Join(s.base, "bin")+":"+os.Getenv("PATH"))
	r := s.run("ctr", "--", "notexec") // bash: "Permission denied", 126 — found, not missing
	r.requireCode(t, 126)
	r = s.run("run", "--local", "--", "notexec") // X2 shares the lookup
	r.requireCode(t, 126)
}

func TestK7ImagePathNonExecutableIs126(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "COPY --chmod=0644 ctrhelper /usr/local/bin/notexec\n")
	r := s.ctr("notexec")
	r.requireCode(t, 126)
}

// TestK8UnderSystemdUnitConmonStaysOutOfX1Scope: offload runs inside a systemd unit (INVOCATION_ID
// set, as on a CI runner service) ⇒ podman would keep conmon in X1's scope, where X1's TERM and
// KILL reach it (the container dies early or its record leaks); conmon must get its own scope.
func TestK8UnderSystemdUnitConmonStaysOutOfX1Scope(t *testing.T) {
	t.Parallel()
	requireScope(t)
	s := ctrSandbox(t, false, "")
	s.setenv("INVOCATION_ID", "0123456789abcdef0123456789abcdef")
	marker := fmt.Sprintf("k8unit-%d", time.Now().UnixNano())
	ready := s.path("ready")
	p := s.start("run", "--local", "--", binaryPath, "ctr", "--", "ctrhelper", "ignore", marker, ready)
	p.await("container start", func() bool { return exists(ready) })
	start := time.Now()
	p.signal(syscall.SIGKILL)
	markerGone(t, marker, 15*time.Second)
	if took := time.Since(start); took < 9*time.Second {
		t.Fatalf("container gone after %s although its command ignores TERM", took)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		r := podman(t, s, "ps", "-a", "--no-trunc", "--format", "{{.ID}} {{.Command}}")
		if r.code != 0 {
			t.Fatalf("podman ps: %d %s", r.code, r.stderr)
		}
		if !strings.Contains(r.stdout, marker) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("container record left after X1's KILL (conmon killed before --rm): %q", r.stdout)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
