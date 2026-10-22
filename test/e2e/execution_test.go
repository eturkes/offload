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

const sessionCommand = `set -m
printf '%s\n' "$$" > "$1/leader.pid"
bash -c 'if [[ $2 == ignore ]]; then trap "" TERM; fi; printf "%s\n" "$$" > "$1/child.pid"; while :; do sleep 1; done' _ "$1" "$2" &
wait
`

func processIdentity(pid int) (group, session int, present bool) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, 0, false
	}
	end := strings.LastIndexByte(string(data), ')')
	if end < 0 {
		return 0, 0, false
	}
	fields := strings.Fields(string(data[end+1:]))
	if len(fields) < 4 {
		return 0, 0, false
	}
	group, groupErr := strconv.Atoi(fields[2])
	session, sessionErr := strconv.Atoi(fields[3])
	return group, session, groupErr == nil && sessionErr == nil
}

func sessionPIDs(t *testing.T, sid int) []int {
	t.Helper()
	entries, err := os.ReadDir("/proc")
	must(t, err)
	var pids []int
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		if _, session, present := processIdentity(pid); present && session == sid {
			pids = append(pids, pid)
		}
	}
	return pids
}

func TestX1ControllerDeathKillsSession(t *testing.T) {
	for _, tc := range []struct {
		name   string
		place  string
		ignore bool
	}{{"local", "--local", false}, {"remote", "--pool", false}, {"local-kill-escalation", "--local", true}} {
		t.Run(tc.name, func(t *testing.T) {
			s := newSandbox(t, true)
			markers := filepath.Join(s.base, "pids")
			must(t, os.MkdirAll(markers, 0o700))
			s.markPIDs(markers)
			mode := "normal"
			if tc.ignore {
				mode = "ignore"
			}
			p := s.start("run", tc.place, "--", "bash", "-c", sessionCommand, "_", markers, mode)
			p.await("session child", func() bool { return exists(filepath.Join(markers, "child.pid")) })
			leader := readPID(t, filepath.Join(markers, "leader.pid"))
			child := readPID(t, filepath.Join(markers, "child.pid"))
			leaderGroup, sid, present := processIdentity(leader)
			childGroup, childSID, childPresent := processIdentity(child)
			_, testSID, testPresent := processIdentity(os.Getpid())
			if !present || !childPresent || !testPresent || sid <= 1 || sid == testSID || sid != childSID || leaderGroup == childGroup {
				t.Fatalf("command must have a new session and two process groups: leader=%d/%d child=%d/%d test-session=%d", leaderGroup, sid, childGroup, childSID, testSID)
			}
			pids := sessionPIDs(t, sid)
			if len(pids) < 2 {
				t.Fatalf("fixture session has only %v", pids)
			}
			t.Cleanup(func() {
				for _, pid := range sessionPIDs(t, sid) {
					if err := syscall.Kill(pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
						t.Errorf("cleanup session pid %d: %v", pid, err)
					}
				}
			})
			p.signal(syscall.SIGKILL)
			p.wait(2 * time.Second)
			if tc.ignore {
				lock, err := os.OpenFile(filepath.Join(s.runtime, "heavy", "slot1"), os.O_RDWR, 0o600)
				must(t, err)
				err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
				must(t, lock.Close())
				if !errors.Is(err, syscall.EWOULDBLOCK) {
					t.Errorf("S5 local slot released before teardown: flock=%v", err)
				}
			}
			requirePIDsGone(t, pids, 13*time.Second)
			if left := sessionPIDs(t, sid); len(left) != 0 {
				t.Errorf("session survived controller SIGKILL: %v", left)
			}
		})
	}
}

func TestX1LeaderExitReapsBackground(t *testing.T) {
	for _, place := range []string{"--local", "--pool"} {
		t.Run(place, func(t *testing.T) {
			s := newSandbox(t, false)
			markers := filepath.Join(s.base, "pids")
			must(t, os.MkdirAll(markers, 0o700))
			s.markPIDs(markers)
			command := `bash -c 'echo $$ > "$1/child.pid"; exec sleep 120' _ "$1" &
while [[ ! -s $1/child.pid ]]; do sleep 0.01; done
exit 7`
			r := s.run("run", place, "--", "bash", "-c", command, "_", markers)
			r.requireCode(t, 7)
			requirePIDsGone(t, []int{readPID(t, filepath.Join(markers, "child.pid"))}, 2*time.Second)
		})
	}
}

// requireScope skips where no user systemd manager (found as the watcher finds it) starts a scope;
// OFFLOAD_REQUIRE_SCOPE=1 (CI with a lingering user manager) turns that absence into a failure.
func requireScope(t *testing.T) {
	t.Helper()
	for _, d := range []string{os.Getenv("XDG_RUNTIME_DIR"), "/run/user/" + strconv.Itoa(os.Getuid())} {
		if fi, err := os.Stat(filepath.Join(d, "systemd", "private")); d == "" || err != nil || fi.Mode()&os.ModeSocket == 0 {
			continue
		}
		c := exec.Command("systemd-run", "--user", "--scope", "--quiet", "--collect", "--expand-environment=no", "--", "true")
		c.Env = append(os.Environ(), "XDG_RUNTIME_DIR="+d)
		if c.Run() == nil {
			return
		}
		break
	}
	if os.Getenv("OFFLOAD_REQUIRE_SCOPE") != "" {
		t.Fatal("OFFLOAD_REQUIRE_SCOPE is set, but systemd-run --user --scope fails")
	}
	t.Skip("no user systemd manager: scope teardown is untestable here")
}

func TestX1ScopeEndsDoubleForkedDaemon(t *testing.T) {
	requireScope(t)
	daemon := `setsid -f bash -c 'echo $$ > "$0"; exec sleep 300' "$1"
while [ ! -s "$1" ]; do sleep 0.01; done
`
	for _, tc := range []struct {
		name, place string
		kill        bool
	}{
		{"local-leader-exit", "--local", false},
		{"remote-leader-exit", "--pool", false},
		{"local-controller-kill", "--local", true},
		{"remote-controller-kill", "--pool", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newSandbox(t, false)
			pidFile := filepath.Join(s.base, "daemon.pid")
			if tc.kill {
				p := s.start("run", tc.place, "--", "bash", "-c", daemon+"sleep 300", "_", pidFile)
				p.await("daemon pid", func() bool { return exists(pidFile) && readPIDIfAny(pidFile) > 1 })
				p.signal(syscall.SIGKILL)
				p.wait(2 * time.Second)
			} else {
				s.run("run", tc.place, "--", "bash", "-c", daemon+"exit 5", "_", pidFile).requireCode(t, 5)
			}
			pid := readPID(t, pidFile)
			t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
			requirePIDsGone(t, []int{pid}, 3*time.Second)
		})
	}
}

func TestX1RefusedScopeFallsBackQuietly(t *testing.T) {
	requireScope(t) // a refusal needs a manager to ask: without one the watcher never tries
	for _, place := range []string{"--local", "--pool"} {
		t.Run(place, func(t *testing.T) {
			s := newSandbox(t, false)
			s.write(filepath.Join(s.base, "bin", "systemd-run"), "#!/bin/sh\necho 'systemd-run: refused' >&2\nexit 1\n")
			must(t, os.Chmod(filepath.Join(s.base, "bin", "systemd-run"), 0o700))
			runs := filepath.Join(s.base, "runs")
			r := s.run("run", place, "--", "bash", "-c", `echo run >> "$1"; echo own-stderr >&2; exit 6`, "_", runs)
			r.requireCode(t, 6)
			if got := s.read(runs); got != "run\n" || strings.Contains(r.stderr, "refused") || !strings.Contains(r.stderr, "own-stderr") ||
				!strings.Contains(r.stderr, "no systemd scope") {
				t.Fatalf("runs %q stderr %q", got, r.stderr)
			}
		})
	}
}

func TestX1NestedRunEndsWithOuterRun(t *testing.T) { // X1 + X3: TERM to an inner watcher = EOF
	s := newSandbox(t, false)
	pidFile := filepath.Join(s.base, "inner.pid")
	p := s.start("run", "--local", "--", binaryPath, "run", "--local", "--",
		"bash", "-c", `echo $$ > "$1"; exec sleep 300`, "_", pidFile)
	p.await("inner command", func() bool { return readPIDIfAny(pidFile) > 1 })
	pid := readPID(t, pidFile)
	t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
	p.signal(syscall.SIGKILL)
	p.wait(2 * time.Second)
	requirePIDsGone(t, []int{pid}, 5*time.Second)
	// The inner offload empties its slot's owner file (O_CREAT) after its command is gone: wait for
	// both slots, so TempDir's cleanup never races that write.
	s.requireSlotsFree(10 * time.Second)
}

// requireSlotsFree waits until every heavy slot of the sandbox can be locked: each holder (offload
// + its watcher) has finished, owner-file writes included.
func (s *sandbox) requireSlotsFree(timeout time.Duration) {
	s.t.Helper()
	deadline := time.Now().Add(timeout)
	for i := 1; i <= 2; i++ {
		path := filepath.Join(s.runtime, "heavy", "slot"+strconv.Itoa(i))
		for {
			f, err := os.OpenFile(path, os.O_WRONLY, 0)
			if errors.Is(err, os.ErrNotExist) {
				break
			}
			must(s.t, err)
			err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
			must(s.t, f.Close())
			if err == nil {
				break
			}
			if time.Now().After(deadline) {
				s.t.Fatalf("slot %d still held after %s", i, timeout)
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
}

func readPIDIfAny(path string) int {
	data, _ := os.ReadFile(path)
	pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	return pid
}

func TestX8RemoteProcessShowsScriptPath(t *testing.T) {
	s := newSandbox(t, true)
	excludes := filepath.Join(s.base, "global-excludes")
	s.write(excludes, "# excludes-marker-7f3a\n*.tmp\n")
	s.git("config", "--global", "core.excludesFile", excludes)
	r := s.run("run", "--pool", "--", "bash", "-c", `tr '\0' ' ' < /proc/$PPID/cmdline`)
	r.requireCode(t, 0)
	requireContains(t, r.stdout, "bash offload/.meta/", "/run-", ".bash offload/project-", " git ")
	for _, leak := range []string{"offload_watch", "excludes-marker-7f3a", "prune_one"} {
		if strings.Contains(r.stdout, leak) {
			t.Errorf("remote argv shows %q: %q", leak, r.stdout)
		}
	}
	if len(r.stdout) > 400 {
		t.Errorf("remote argv is %d bytes: %q", len(r.stdout), r.stdout)
	}
}

func TestX2ArgvAndEnvironment(t *testing.T) {
	for _, place := range []string{"--local", "--pool"} {
		t.Run(place, func(t *testing.T) {
			s := newSandbox(t, false)
			s.setenv("OFFLOAD_TEST_INHERITED", "inherited-value")
			words := []string{"$HOME", "$(touch injection)", "; touch injection", "two words", "'single'", `"double"`, "line\nbreak", "", "*", "\\"}
			args := append([]string{"run", place, "--", "printf", `%s\0`}, words...)
			r := s.run(args...)
			r.requireCode(t, 0)
			if want := strings.Join(words, "\x00") + "\x00"; r.stdout != want {
				t.Errorf("argv changed: got %q, want %q", r.stdout, want)
			}
			assertAbsent(t, filepath.Join(s.root, "injection"), filepath.Join(s.workdir("alpha"), "injection"))
			r = s.run("run", place, "--", "bash", "-c", "printf '%s' \"$OFFLOAD_TEST_INHERITED\"; if read -r line; then exit 88; fi")
			r.requireCode(t, 0)
			if place == "--local" && r.stdout != "inherited-value" {
				t.Errorf("local environment not inherited: %q", r.stdout)
			}
		})
	}
	t.Run("remote-login", func(t *testing.T) {
		s := newSandbox(t, false)
		s.write(filepath.Join(s.hosts, "alpha", ".bash_profile"), "export OFFLOAD_LOGIN_MARKER=login-shell\n")
		r := s.run("run", "--pool", "--", "bash", "-c", "printf '%s' \"$OFFLOAD_LOGIN_MARKER\"")
		r.requireCode(t, 0)
		if r.stdout != "login-shell" {
			t.Errorf("remote command did not receive bash login environment: %q", r.stdout)
		}
	})
}

func TestX3SignalsSkipDownload(t *testing.T) {
	for _, signal := range []struct {
		name string
		sig  syscall.Signal
		code int
	}{{"TERM", syscall.SIGTERM, 143}, {"INT", syscall.SIGINT, 130}, {"HUP", syscall.SIGHUP, 129}} {
		t.Run(signal.name, func(t *testing.T) {
			s := newSandbox(t, true)
			s.write(".gitignore", "explicit\n")
			markers := filepath.Join(s.base, "pids")
			must(t, os.MkdirAll(markers, 0o700))
			s.markPIDs(markers)
			command := "printf automatic > automatic; printf explicit > explicit; echo $$ > \"$1/leader.pid\"; exec sleep 120"
			p := s.start("run", "--pool", "--get", "explicit", "--", "bash", "-c", command, "_", markers)
			p.await("remote output and command pid", func() bool { return exists(filepath.Join(markers, "leader.pid")) })
			pid := readPID(t, filepath.Join(markers, "leader.pid"))
			p.signal(signal.sig)
			r := p.wait(13 * time.Second)
			r.requireCode(t, signal.code)
			requirePIDsGone(t, []int{pid}, time.Second)
			assertAbsent(t, filepath.Join(s.root, "automatic"), filepath.Join(s.root, "explicit"))
			if s.read(filepath.Join(s.workdir("alpha"), "automatic")) != "automatic" {
				t.Error("fixture did not create remote output")
			}
		})
	}
}

func TestX4ByteExactStreaming(t *testing.T) {
	for _, place := range []string{"--local", "--pool"} {
		t.Run(place, func(t *testing.T) {
			s := newSandbox(t, false)
			release := filepath.Join(s.base, "release")
			command := `printf '\000\001\377\r\n'; printf '\002\376\r\n' >&2; while [[ ! -e $1 ]]; do sleep 0.02; done; printf 'tail\000'`
			p := s.start("run", place, "--", "bash", "-c", command, "_", release)
			first := []byte{0, 1, 255, '\r', '\n'}
			p.await("live stdout/stderr before command can exit", func() bool {
				r := p.output()
				return bytes.Equal([]byte(r.stdout), first) && strings.Contains(r.stderr, string([]byte{2, 254, '\r', '\n'}))
			})
			s.write(release, "continue\n")
			r := p.wait(5 * time.Second)
			r.requireCode(t, 0)
			want := append(first, []byte("tail\x00")...)
			if !bytes.Equal([]byte(r.stdout), want) {
				t.Errorf("stdout bytes=%x, want %x", []byte(r.stdout), want)
			}
		})
	}
}

func TestX6Remote255AndTransport(t *testing.T) {
	t.Run("command-255", func(t *testing.T) {
		s := newSandbox(t, true)
		r := s.run("run", "--pool", "--", "bash", "-c", "printf result > result; exit 255")
		r.requireCode(t, 255)
		if s.read("result") != "result" {
			t.Error("command's 255 suppressed download")
		}
	})
	t.Run("transport-255", func(t *testing.T) {
		s := newSandbox(t, true)
		s.write(filepath.Join(s.hosts, "alpha", ".transport-fail"), "fail after probe\n")
		r := s.run("run", "--pool", "--", "true")
		r.requireCode(t, 125)
		requireContains(t, r.stderr, "simulated connection reset")
	})
}

// x5 = the options offload passes on every connection; the fake ssh refuses a call without them.
var x5 = []string{
	"-o", "BatchMode=yes", "-o", "ConnectTimeout=5", "-o", "StrictHostKeyChecking=accept-new",
	"-o", "ServerAliveInterval=15", "-o", "ServerAliveCountMax=4",
}

func TestHarnessFakeSSH(t *testing.T) {
	s := newSandbox(t, false)
	ssh := filepath.Join(s.base, "bin", "ssh")
	t.Run("missing-x5-option-refused", func(t *testing.T) {
		cmd := exec.Command(ssh, "-o", "BatchMode=yes", "alpha", "true")
		cmd.Env = s.env
		out, err := cmd.CombinedOutput()
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 254 {
			t.Fatalf("call without X5 options: exit=%v %s", err, out)
		}
	})
	t.Run("null-stdin", func(t *testing.T) {
		cmd := exec.Command(ssh, append(append([]string{}, x5...), "-n", "alpha", "cat")...)
		cmd.Env = s.env
		cmd.Stdin = strings.NewReader("must not arrive")
		out, err := cmd.CombinedOutput()
		if err != nil || len(out) != 0 {
			t.Fatalf("-n forwarded stdin: %q %v", out, err)
		}
	})
	t.Run("options-and-stdin", func(t *testing.T) {
		cmd := exec.Command(ssh, append(append([]string{}, x5...), "-T", "-l", "agent", "--", "alpha", "printf '%s\\n' \"$HOME\"; cat")...)
		cmd.Env = s.env
		cmd.Stdin = strings.NewReader("stdin-through-EOF\x00")
		out, err := cmd.CombinedOutput()
		must(t, err)
		want := filepath.Join(s.hosts, "alpha") + "\nstdin-through-EOF\x00"
		if string(out) != want {
			t.Errorf("fake ssh bytes=%q, want %q", out, want)
		}
	})
	t.Run("down-host", func(t *testing.T) {
		cmd := exec.Command(ssh, append(append([]string{}, x5...), "missing", "true")...)
		cmd.Env = s.env
		out, err := cmd.CombinedOutput()
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 255 {
			t.Fatalf("down host exit=%v", err)
		}
		requireContains(t, string(out), "ssh: connect to host missing port 22: Connection refused")
	})
	t.Run("real-rsync", func(t *testing.T) {
		s.write("payload", "binary\x00payload\xff")
		cmd := exec.Command("rsync", "-a", "-e", ssh+" "+strings.Join(x5, " ")+" -T", filepath.Join(s.root, "payload"), "agent@alpha:received")
		cmd.Env = s.env
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("rsync over fake ssh: %v: %s", err, out)
		}
		if got := s.read(filepath.Join(s.hosts, "alpha", "received")); got != s.read("payload") {
			t.Errorf("rsync payload=%q", got)
		}
	})
}
