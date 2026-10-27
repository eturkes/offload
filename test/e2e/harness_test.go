package e2e_test

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io/fs"
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

var (
	binaryPath string
	helperPath string // ctrhelper, the e2e image's one program (ctr_test.go)
	fixtureDir string
)

func TestMain(m *testing.M) {
	// Fixture modes (WriteFile, git checkouts) pass through the umask, and an image's root-owned
	// copy must stay readable + executable by the keep-id user: an agent session's umask 077 made
	// the e2e image's ctrhelper 0700 → 126. offload's own files carry explicit modes.
	syscall.Umask(0o022)
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fixtureDir = filepath.Join(cwd, "testdata")
	tmp, err := os.MkdirTemp("", "offload-e2e-build-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	binaryPath, helperPath = filepath.Join(tmp, "offload"), filepath.Join(tmp, "ctrhelper")
	// Static, as installed: the binary is also the init inside `FROM scratch` test images (K11).
	for out, pkg := range map[string]string{binaryPath: "./cmd/offload", helperPath: "./test/e2e/testdata/ctrhelper"} {
		build := exec.Command("go", "build", "-o", out, pkg)
		build.Dir = filepath.Clean(filepath.Join(cwd, "../.."))
		build.Env = append(os.Environ(), "CGO_ENABLED=0")
		build.Stdout, build.Stderr = os.Stdout, os.Stderr
		if err := build.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			if removeErr := os.RemoveAll(tmp); removeErr != nil {
				fmt.Fprintln(os.Stderr, removeErr)
			}
			os.Exit(1)
		}
	}
	code := m.Run()
	if err := removeStorage(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	if err := os.RemoveAll(tmp); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	os.Exit(code)
}

type host struct {
	name  string
	cpus  int
	load  float64
	gpus  int
	slots int
	down  bool
	mux   bool
}

type sandbox struct {
	t       *testing.T
	base    string
	root    string
	cwd     string
	runtime string
	hosts   string
	config  string
	env     []string
}

func newSandbox(t *testing.T, git bool, hosts ...host) *sandbox {
	t.Helper()
	base := t.TempDir()
	s := &sandbox{
		t: t, base: base, root: filepath.Join(base, "project"),
		runtime: filepath.Join(base, "runtime"), hosts: filepath.Join(base, "hosts"),
		config: filepath.Join(base, "pool.toml"),
	}
	s.cwd = s.root
	for _, dir := range []string{s.root, s.runtime, s.hosts, filepath.Join(base, "home"), filepath.Join(base, "bin")} {
		must(t, os.MkdirAll(dir, 0o700))
	}
	ssh, err := os.ReadFile(filepath.Join(fixtureDir, "ssh"))
	must(t, err)
	must(t, os.WriteFile(filepath.Join(base, "bin", "ssh"), ssh, 0o700))
	s.env = replaceEnv(os.Environ(), map[string]string{
		"HOME": filepath.Join(base, "home"), "XDG_CONFIG_HOME": filepath.Join(base, "config"),
		"XDG_RUNTIME_DIR": s.runtime, "XDG_STATE_HOME": filepath.Join(base, "state"),
		"OFFLOAD_CONFIG": s.config, "HEAVY_SLOTS": "2",
		"OFFLOAD_FAKE_HOSTS": s.hosts, "PATH": filepath.Join(base, "bin") + ":" + os.Getenv("PATH"),
		"GIT_CONFIG_NOSYSTEM": "1", "GIT_CONFIG_GLOBAL": filepath.Join(base, "gitconfig"),
		"GIT_AUTHOR_NAME": "e2e", "GIT_AUTHOR_EMAIL": "e2e@example.invalid",
		"GIT_COMMITTER_NAME": "e2e", "GIT_COMMITTER_EMAIL": "e2e@example.invalid",
		"LC_ALL": "C", "TERM": "dumb",
		"OFFLOAD_CTR": "", // the machine's own switch (K2) must not leak in; ctrSandbox turns it on
	})
	if len(hosts) == 0 {
		hosts = []host{{name: "alpha", cpus: 8, load: 2, slots: 2}}
	}
	var pool strings.Builder
	for _, h := range hosts {
		if h.slots == 0 {
			h.slots = 2
		}
		fmt.Fprintf(&pool, "[[host]]\nname = %q\naddress = %q\nslots = %d\nmultiplex = %t\n", h.name, "agent@"+h.name, h.slots, h.mux)
		if !h.down {
			must(t, os.MkdirAll(filepath.Join(s.hosts, h.name), 0o700))
			s.write(filepath.Join(s.hosts, h.name, ".probe"), fmt.Sprintf("%d %.6f 1048576 %d\n", h.cpus, h.load, h.gpus))
		}
	}
	s.write(s.config, pool.String())
	if git {
		s.git("init", "-q", "-b", "main")
		s.write("tracked", "initial\n")
		s.git("add", ".")
		s.git("commit", "-qm", "fixture")
	}
	return s
}

func replaceEnv(env []string, updates map[string]string) []string {
	out := make([]string, 0, len(env)+len(updates))
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		if _, replaced := updates[key]; !replaced && !strings.HasPrefix(key, "GIT_") {
			out = append(out, entry)
		}
	}
	for key, value := range updates {
		out = append(out, key+"="+value)
	}
	return out
}

func (s *sandbox) setenv(key, value string) {
	prefix := key + "="
	for i, entry := range s.env {
		if strings.HasPrefix(entry, prefix) {
			s.env[i] = prefix + value
			return
		}
	}
	s.env = append(s.env, prefix+value)
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func (s *sandbox) path(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(s.root, path)
}

func (s *sandbox) write(path, content string) {
	s.t.Helper()
	path = s.path(path)
	must(s.t, os.MkdirAll(filepath.Dir(path), 0o700))
	must(s.t, os.WriteFile(path, []byte(content), 0o600))
}

func (s *sandbox) read(path string) string {
	s.t.Helper()
	data, err := os.ReadFile(s.path(path))
	must(s.t, err)
	return string(data)
}

func (s *sandbox) git(args ...string) []byte {
	s.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir, cmd.Env = s.root, s.env
	out, err := cmd.CombinedOutput()
	if err != nil {
		s.t.Fatalf("git %q: %v\n%s", args, err, out)
	}
	return out
}

// shell = stdout of bash -c script in the sandbox cwd + environment (a local oracle).
func (s *sandbox) shell(script string) string {
	s.t.Helper()
	cmd := exec.Command("bash", "-c", script)
	cmd.Dir, cmd.Env = s.cwd, s.env
	out, err := cmd.Output()
	must(s.t, err)
	return string(out)
}

func (s *sandbox) workdir(name string) string {
	digest := sha1.Sum([]byte(s.root))
	return filepath.Join(s.hosts, name, "offload", fmt.Sprintf("%s-%x", filepath.Base(s.root), digest[:4]))
}

type result struct {
	code   int
	stdout string
	stderr string
}

func (r result) requireCode(t *testing.T, want int) {
	t.Helper()
	if r.code != want {
		t.Fatalf("exit = %d, want %d; stdout=%q stderr=%q", r.code, want, r.stdout, r.stderr)
	}
}

type running struct {
	t      *testing.T
	cmd    *exec.Cmd
	out    *os.File
	errout *os.File
	done   chan struct{}
	err    error
}

func (s *sandbox) start(args ...string) *running {
	s.t.Helper()
	out, err := os.CreateTemp(s.base, "stdout-")
	must(s.t, err)
	errout, err := os.CreateTemp(s.base, "stderr-")
	must(s.t, err)
	cmd := exec.Command(binaryPath, args...)
	cmd.Env, cmd.Dir = s.env, s.cwd
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout, cmd.Stderr = out, errout
	must(s.t, cmd.Start())
	p := &running{t: s.t, cmd: cmd, out: out, errout: errout, done: make(chan struct{})}
	go func() {
		p.err = cmd.Wait()
		close(p.done)
	}()
	s.t.Cleanup(func() {
		select {
		case <-p.done:
		default:
			if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
				s.t.Errorf("cleanup offload: %v", err)
			}
			select {
			case <-p.done:
			case <-time.After(3 * time.Second):
				s.t.Error("offload did not exit after cleanup SIGKILL")
			}
		}
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
			s.t.Errorf("cleanup process group: %v", err)
		}
		must(s.t, out.Close())
		must(s.t, errout.Close())
	})
	return p
}

func (p *running) output() result {
	p.t.Helper()
	out, err := os.ReadFile(p.out.Name())
	must(p.t, err)
	errout, err := os.ReadFile(p.errout.Name())
	must(p.t, err)
	return result{stdout: string(out), stderr: string(errout)}
}

func (p *running) wait(timeout time.Duration) result {
	p.t.Helper()
	select {
	case <-p.done:
		r := p.output()
		if p.err != nil {
			var exit *exec.ExitError
			if !errors.As(p.err, &exit) {
				p.t.Fatal(p.err)
			}
			r.code = exit.ExitCode()
		}
		return r
	case <-time.After(timeout):
		r := p.output()
		p.t.Fatalf("offload timed out after %s; stdout=%q stderr=%q", timeout, r.stdout, r.stderr)
		return result{}
	}
}

// hang bounds a wait whose test claims no latency (readiness, completion): a hung run still fails,
// a loaded host only waits longer (podman steps slow down 10-20x under CPU contention). Waits
// that check a promised latency (signal response, X1's 10 s, J4's 2 s, K8's "at once") keep
// their own bounds.
const hang = 2 * time.Minute

func (s *sandbox) run(args ...string) result {
	s.t.Helper()
	return s.start(args...).wait(hang)
}

func (p *running) await(what string, condition func() bool) {
	p.t.Helper()
	deadline := time.Now().Add(hang)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		select {
		case <-p.done:
			r := p.wait(time.Second)
			p.t.Fatalf("offload exited before %s: exit=%d stdout=%q stderr=%q", what, r.code, r.stdout, r.stderr)
		case <-time.After(20 * time.Millisecond):
		}
	}
	r := p.output()
	p.t.Fatalf("no %s within %s; stdout=%q stderr=%q", what, hang, r.stdout, r.stderr)
}

func (p *running) signal(signal syscall.Signal) {
	p.t.Helper()
	must(p.t, p.cmd.Process.Signal(signal))
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func assertAbsent(t *testing.T, paths ...string) {
	t.Helper()
	for _, path := range paths {
		if _, err := os.Lstat(path); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("path must be absent: %q (lstat error=%v)", path, err)
		}
	}
}

func requireContains(t *testing.T, got string, want ...string) {
	t.Helper()
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("%q does not contain %q", got, part)
		}
	}
}

func (s *sandbox) holdSlot(hostname string, index int) func() {
	s.t.Helper()
	dir := filepath.Join(s.runtime, "heavy")
	if hostname != "" {
		dir = filepath.Join(s.runtime, "offload", "hosts", hostname)
	}
	must(s.t, os.MkdirAll(dir, 0o700))
	file, err := os.OpenFile(filepath.Join(dir, fmt.Sprintf("slot%d", index)), os.O_CREATE|os.O_RDWR, 0o600)
	must(s.t, err)
	must(s.t, syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB))
	released := false
	release := func() {
		if !released {
			released = true
			must(s.t, file.Close())
		}
	}
	s.t.Cleanup(release)
	return release
}

func (s *sandbox) markPIDs(dir string) {
	s.t.Helper()
	s.t.Cleanup(func() {
		paths, err := filepath.Glob(filepath.Join(dir, "*.pid"))
		must(s.t, err)
		for _, path := range paths {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err == nil && pid > 1 {
				if err := syscall.Kill(pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
					s.t.Errorf("cleanup pid %d: %v", pid, err)
				}
			}
		}
	})
}

func readPID(t *testing.T, path string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	must(t, err)
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	must(t, err)
	if pid <= 1 {
		t.Fatalf("invalid child pid %d", pid)
	}
	return pid
}

func requirePIDsGone(t *testing.T, pids []int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		remaining := []int{}
		for _, pid := range pids {
			if err := syscall.Kill(pid, 0); !errors.Is(err, syscall.ESRCH) {
				remaining = append(remaining, pid)
			}
		}
		if len(remaining) == 0 {
			return
		}
		time.Sleep(30 * time.Millisecond)
	}
	for _, pid := range pids {
		if err := syscall.Kill(pid, 0); !errors.Is(err, syscall.ESRCH) {
			t.Errorf("session process %d remains after %s (kill(0)=%v)", pid, timeout, err)
		}
	}
}

func leafPaths(t *testing.T, root string) []string {
	t.Helper()
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == ".git" && entry.IsDir() {
			return filepath.SkipDir
		}
		if !entry.IsDir() {
			paths = append(paths, rel)
		}
		return nil
	})
	must(t, err)
	slices.Sort(paths)
	return paths
}

func nulStrings(data []byte) []string {
	var paths []string
	for _, path := range bytes.Split(data, []byte{0}) {
		if len(path) != 0 {
			paths = append(paths, string(path))
		}
	}
	return paths
}
