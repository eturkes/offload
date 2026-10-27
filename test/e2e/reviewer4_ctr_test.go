package e2e_test

import (
	"bytes"
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

func reviewFakePodman(t *testing.T, s *sandbox, cases string) {
	t.Helper()
	fakePodman(t, s, cases+"\n*) echo unexpected-podman-command >&2; exit 98 ;;\n")
}

func TestK3ReviewFindErrorsBeforePodman(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.setenv("OFFLOAD_CTR", "1")
	s.write(".devcontainer/Containerfile", "FROM scratch\n")
	s.write(".git/HEAD", "broken\n")
	s.setenv("GIT_CEILING_DIRECTORIES", s.base)
	reviewFakePodman(t, s, "")
	r := s.run("ctr", "--", "unused-command")
	r.requireCode(t, 125)
	requireContains(t, r.stderr, "git does not recognize the repository")
}

func TestK1ReviewNativeWordsAndENOEXEC(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	s.setenv("OFFLOAD_CTR", "0")
	s.write(s.config, "not valid pool syntax")
	s.write(".git", "not a git repository")
	r := s.run("ctr", "printf", "%s\\n", "a b", "--check", "--help", "$HOME", "")
	r.requireCode(t, 0)
	if r.stdout != "a b\n--check\n--help\n$HOME\n\n" {
		t.Fatalf("native CMD words changed: %q", r.stdout)
	}
	s.write("script-without-shebang", "exit 73\n")
	must(t, os.Chmod(s.path("script-without-shebang"), 0o755))
	s.run("ctr", "--", "./script-without-shebang").requireCode(t, 126)
}

func TestK7ReviewCommandStatusWithoutPodman(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.setenv("OFFLOAD_CTR", "1")
	s.write(".devcontainer/Containerfile", "FROM scratch\n")
	reviewFakePodman(t, s, `"image exists") exit 0 ;;
run*) printf '7\n' >&3; exit 42 ;;`)
	r := s.run("ctr", "--", "unused-command")
	r.requireCode(t, 7)
}

func TestK6ReviewStreamsWithoutPodman(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.setenv("OFFLOAD_CTR", "1")
	s.write(".devcontainer/Containerfile", "FROM scratch\n")
	reviewFakePodman(t, s, `"image exists") exit 0 ;;
run*) printf 'out\000\377\r'; printf 'err\001\376' >&2; cat; printf '7\n' >&3; exit 42 ;;`)
	cmd := exec.Command(binaryPath, "ctr", "--", "unused-command")
	cmd.Dir, cmd.Env = s.cwd, s.env
	cmd.Stdin = strings.NewReader("in\x00put")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	var ee *exec.ExitError
	if !errors.As(err, &ee) || ee.ExitCode() != 7 || stdout.String() != "out\x00\xff\rin\x00put" || stderr.String() != "err\x01\xfe" {
		t.Fatalf("streams/status: err=%v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
}

func TestK8ReviewControlSignalWithoutPodman(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.setenv("OFFLOAD_CTR", "1")
	s.write(".devcontainer/Containerfile", "FROM scratch\n")
	ready := filepath.Join(s.base, "ready")
	reviewFakePodman(t, s, `"image exists") exit 0 ;;
run*)
  touch `+ready+`
  IFS= read -r -N1 byte <&3 || exit 41
  [ "$(printf '%d' "'$byte")" = 15 ] || exit 42
  printf '5\n' >&3; exit 43 ;;`)
	p := s.start("ctr", "--", "unused-command")
	p.await("control receiver", func() bool { return exists(ready) })
	p.signal(syscall.SIGTERM)
	p.wait(4*time.Second).requireCode(t, 5)
}

func TestK5ReviewCancelActiveBuild(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.setenv("OFFLOAD_CTR", "1")
	s.write(".devcontainer/Containerfile", "FROM scratch\n")
	ready, ended := filepath.Join(s.base, "build-ready"), filepath.Join(s.base, "build-term")
	reviewFakePodman(t, s, `"image exists") exit 1 ;;
"build --timestamp")
  trap 'echo TERM > `+ended+`; exit 44' TERM
  echo $$ > `+ready+`
  while :; do sleep 0.05; done ;;`)
	p := s.start("ctr", "--", "unused-command")
	p.await("build signal handler", func() bool { return exists(ready) })
	pid := readPID(t, ready)
	t.Cleanup(func() { _ = syscall.Kill(-pid, syscall.SIGKILL) })
	p.signal(syscall.SIGTERM)
	p.wait(4*time.Second).requireCode(t, 143)
	if got := readOr(ended); got != "TERM\n" {
		t.Fatalf("active build did not receive TERM: %q", got)
	}
}

func TestK5ReviewBuildKeepsLockWithoutPodman(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.setenv("OFFLOAD_CTR", "1")
	s.write(".devcontainer/Containerfile", "FROM scratch\n")
	ready := filepath.Join(s.base, "build-ready")
	reviewFakePodman(t, s, `"image exists") exit 1 ;;
"build --timestamp") echo $$ > `+ready+`; exec sleep 30 ;;`)
	p := s.start("ctr", "--", "unused-command")
	p.await("build lock holder", func() bool { return exists(ready) })
	build, err := os.FindProcess(readPID(t, ready))
	must(t, err)
	t.Cleanup(func() { _ = build.Kill(); _ = build.Release() })
	locks, err := filepath.Glob(filepath.Join(s.runtime, "offload", "ctr", "*.lock"))
	must(t, err)
	if len(locks) != 1 {
		t.Fatalf("lock count=%d want=1", len(locks))
	}
	lock, err := os.OpenFile(locks[0], os.O_WRONLY, 0)
	must(t, err)
	defer func() { _ = lock.Close() }()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); !errors.Is(err, syscall.EWOULDBLOCK) {
		t.Fatalf("positive control: active build lock = %v, want contention", err)
	}
	p.signal(syscall.SIGKILL)
	p.wait(4 * time.Second)
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); !errors.Is(err, syscall.EWOULDBLOCK) {
		t.Fatalf("orphaned build no longer owns lock: %v", err)
	}
	must(t, build.Kill())
	deadline := time.Now().Add(4 * time.Second)
	for syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) != nil {
		if time.Now().After(deadline) {
			t.Fatal("dead build still owns lock")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestK8ReviewInitRefusesOutsidePID1(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	// The nonexistent CMD makes the PID-check mutant stop in start(), before teardown can run.
	r := s.run("__ctr-init", s.path("guaranteed-missing-command"))
	r.requireCode(t, 125)
	requireContains(t, r.stderr, "inside offload ctr's container only")
}

func TestK9ReviewLayersFailureAndSwitch(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.setenv("OFFLOAD_CTR", "invalid-but-ignored")
	s.write(".devcontainer/Containerfile", "FROM scratch\n")
	reviewFakePodman(t, s, `"image exists") exit 0 ;;
"image inspect") printf 'sha256:partial\n'; exit 42 ;;`)
	r := s.run("ctr", "--layers")
	r.requireCode(t, 125)
	if r.stdout != "" {
		t.Fatalf("failed inspect leaked partial stdout: %q", r.stdout)
	}
	reviewFakePodman(t, s, `"image exists") exit 0 ;;
"image inspect") exit 0 ;;`)
	r = s.run("ctr", "--layers")
	r.requireCode(t, 0)
	if r.stdout != "" {
		t.Fatalf("zero layers emitted bytes: %q", r.stdout)
	}
}

func TestK10ReviewOrderedLayersAndHosts(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true, host{name: "zeta", cpus: 8, slots: 2}, host{name: "alpha", cpus: 8, slots: 2})
	s.setenv("OFFLOAD_CTR", "invalid-but-ignored")
	s.write(".devcontainer/Containerfile", "FROM scratch\n")
	bin, err := os.ReadFile(binaryPath)
	must(t, err)
	must(t, os.WriteFile(filepath.Join(s.base, "bin", "offload"), bin, 0o755))
	reviewFakePodman(t, s, `"image exists") exit 0 ;;
"image inspect")
  case "$(cat "$HOME/check-case" 2>/dev/null)" in
  reversed) printf 'sha256:b\nsha256:a\n' ;;
  first-only) printf 'sha256:a\n' ;;
  last-differs) printf 'sha256:a\nsha256:c\n' ;;
  *) printf 'sha256:a\nsha256:b\n' ;;
  esac
  exit 0 ;;`)
	for _, tc := range []struct {
		mode, first string
		code        int
	}{
		{"same", "zeta: same (2 layers)", 0},
		{"reversed", "zeta: different (here 2, there 2 layers)", 1},
		{"first-only", "zeta: different (here 2, there 1 layers)", 1},
		{"last-differs", "zeta: different (here 2, there 2 layers)", 1},
	} {
		s.write(filepath.Join(s.hosts, "zeta", "check-case"), tc.mode)
		r := s.start("ctr", "--check").wait(hang)
		r.requireCode(t, tc.code)
		if want := tc.first + "\nalpha: same (2 layers)\n"; r.stdout != want {
			t.Fatalf("%s: output=%q want=%q", tc.mode, r.stdout, want)
		}
	}
}

func TestK10ReviewHostsStartInParallel(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true, host{name: "zeta", cpus: 8, slots: 2}, host{name: "alpha", cpus: 8, slots: 2})
	s.write(".devcontainer/Containerfile", "FROM scratch\n")
	bin, err := os.ReadFile(binaryPath)
	must(t, err)
	must(t, os.WriteFile(filepath.Join(s.base, "bin", "offload"), bin, 0o755))
	markers := filepath.Join(s.base, "markers")
	must(t, os.Mkdir(markers, 0o700))
	s.setenv("OFFLOAD_REVIEW_PARALLEL", markers)
	reviewFakePodman(t, s, `"image exists") exit 0 ;;
"image inspect")
  case "$(basename "$HOME")" in
  alpha) own=alpha; other=zeta ;;
  zeta) own=zeta; other=alpha ;;
  esac
  if [ -n "${own:-}" ]; then
    touch "$OFFLOAD_REVIEW_PARALLEL/$own"
    deadline=$((SECONDS+5))
    until [ -f "$OFFLOAD_REVIEW_PARALLEL/$other" ]; do
      [ "$SECONDS" -lt "$deadline" ] || exit 44
      sleep .01
    done
  fi
  printf 'sha256:a\nsha256:b\n'; exit 0 ;;`)
	r := s.start("ctr", "--check").wait(hang)
	r.requireCode(t, 0)
	if r.stdout != "zeta: same (2 layers)\nalpha: same (2 layers)\n" {
		t.Fatalf("parallel host verdicts = %q", r.stdout)
	}
}

func TestK10ReviewLocalFailuresBeforeRemote(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.setenv("OFFLOAD_CTR", "1")
	s.write(".devcontainer/Containerfile", "FROM scratch\n")
	called := filepath.Join(s.base, "podman-called")
	reviewFakePodman(t, s, `*) echo called > `+called+`; exit 42 ;;`)
	must(t, os.Remove(s.config))
	r := s.run("ctr", "--check")
	r.requireCode(t, 2)
	assertAbsent(t, called)
	s.write(s.config, "[[host]]\nname='alpha'\naddress='agent@alpha'\n")
	r = s.run("ctr", "--check")
	r.requireCode(t, 125)
	if r.stdout != "" {
		t.Fatalf("local failure produced host verdicts: %q", r.stdout)
	}
	calls, err := filepath.Glob(filepath.Join(s.hosts, "alpha", ".ssh-call-*"))
	must(t, err)
	if len(calls) != 0 {
		t.Fatalf("local failure contacted a host: %v", calls)
	}
}

func TestK7ReviewImageENOEXECAndFD3(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	s.write("script-without-shebang", "exit 73\n")
	must(t, os.Chmod(s.path("script-without-shebang"), 0o755))
	s.ctr("./script-without-shebang").requireCode(t, 126)
	s.ctr("ctrhelper", "forge-status").requireCode(t, 7)
}

func TestK8ReviewAllSignals(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	log := s.path("signals.log")
	p := s.start("ctr", "--", "ctrhelper", "trap", log, "3", "5")
	p.await("signal handlers", func() bool { return readOr(log) == "ready\n" })
	want := "ready\n"
	for _, sig := range []syscall.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2} {
		p.signal(sig)
		want += strings.ToUpper(sig.String()) + "\n"
		p.await(fmt.Sprintf("signal %d", sig), func() bool { return readOr(log) == want })
	}
	p.wait(10*time.Second).requireCode(t, 5)
	if got := readOr(log); got != want {
		t.Fatalf("signal log=%q want=%q", got, want)
	}
}

func TestK8ReviewLeftoverTERM(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	log := s.path("leftover.log")
	s.ctr("ctrhelper", "spawn-trap", log, "6").requireCode(t, 6)
	if got := readOr(log); got != "ready\nTERMINATED\n" {
		t.Fatalf("leftover child did not receive teardown TERM: %q", got)
	}
}

func TestK8ReviewWholeCommandGroup(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	log := s.path("group.log")
	p := s.start("ctr", "--", "ctrhelper", "trap-pair", log, "6")
	p.await("both signal handlers", func() bool { return readOr(log) == "ready\n" && readOr(log+".child") == "ready\n" })
	p.signal(syscall.SIGUSR1)
	p.wait(10*time.Second).requireCode(t, 6)
	if parent, child := readOr(log), readOr(log+".child"); parent != "ready\nUSR1\n" || child != "ready\nUSER DEFINED SIGNAL 1\n" {
		t.Fatalf("group signal: parent=%q child=%q", parent, child)
	}
}

func reviewMarkerPIDs(t *testing.T, marker string) []int {
	t.Helper()
	out, err := exec.Command("pgrep", "-f", "ctrhelper ignore "+marker).Output()
	must(t, err)
	var pids []int
	for _, field := range strings.Fields(string(out)) {
		pid, err := strconv.Atoi(field)
		must(t, err)
		if pid <= 1 {
			t.Fatalf("invalid marker pid %d", pid)
		}
		pids = append(pids, pid)
	}
	if len(pids) == 0 {
		t.Fatal("positive control: the ready command is not visible")
	}
	return pids
}

func TestK8ReviewLivenessEndsContainer(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	marker := fmt.Sprintf("review-live-%d", time.Now().UnixNano())
	ready := s.path("ready")
	p := s.start("ctr", "--", "ctrhelper", "ignore", marker, ready)
	p.await("container start", func() bool { return exists(ready) })
	live := reviewMarkerPIDs(t, marker)
	listing := podman(t, s, "ps", "--no-trunc", "--format", "{{.ID}} {{.Command}}")
	listing.requireCode(t, 0)
	var ids []string
	for _, line := range strings.Split(listing.stdout, "\n") {
		if strings.Contains(line, marker) {
			ids = append(ids, strings.Fields(line)[0])
		}
	}
	if len(ids) != 1 {
		t.Fatalf("expected one owned container for %s, got %q", marker, listing.stdout)
	}
	t.Cleanup(func() { podman(t, s, "rm", "-f", "--ignore", ids[0]).requireCode(t, 0) })
	p.signal(syscall.SIGKILL)
	p.wait(2 * time.Second)
	requirePIDsGone(t, live, 5*time.Second)
}

func TestK8ReviewPodmanGoneFirst(t *testing.T) {
	t.Parallel()
	s := ctrSandbox(t, false, "")
	marker := fmt.Sprintf("k8podman-%d", time.Now().UnixNano())
	ready := s.path("ready")
	p := s.start("ctr", "--", "ctrhelper", "ignore", marker, ready)
	p.await("container start", func() bool { return exists(ready) })
	live := reviewMarkerPIDs(t, marker)
	kids, err := exec.Command("pgrep", "-P", fmt.Sprint(p.cmd.Process.Pid)).Output()
	must(t, err)
	pids := strings.Fields(string(kids))
	if len(pids) != 1 {
		t.Fatalf("expected one Podman child, got %q", kids)
	}
	cmd := exec.Command("kill", "-KILL", pids[0])
	must(t, cmd.Run())
	p.wait(10*time.Second).requireCode(t, 125)
	requirePIDsGone(t, live, 5*time.Second)
}
