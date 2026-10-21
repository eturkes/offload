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

func TestHarnessFakeSSH(t *testing.T) {
	s := newSandbox(t, false)
	ssh := filepath.Join(s.base, "bin", "ssh")
	t.Run("options-and-stdin", func(t *testing.T) {
		cmd := exec.Command(ssh, "-o", "BatchMode=yes", "-oStrictHostKeyChecking=accept-new", "-T", "-l", "agent", "--", "alpha", "printf '%s\\n' \"$HOME\"; cat")
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
		cmd := exec.Command(ssh, "missing", "true")
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
		cmd := exec.Command("rsync", "-a", "-e", ssh+" -o BatchMode=yes -T", filepath.Join(s.root, "payload"), "agent@alpha:received")
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
