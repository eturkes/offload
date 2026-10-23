package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestReviewP2ExplicitZeroSlots(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	s.write(s.config, "[[host]]\nname = \"alpha\"\naddress = \"agent@alpha\"\nslots = 0\n")
	r := s.run("run", "--", "touch", "must-not-run")
	r.requireCode(t, 2)
	assertAbsent(t, filepath.Join(s.root, "must-not-run"), filepath.Join(s.runtime, "heavy", "slot1"))
}

func TestReviewR2FreshStatusHasNoHolders(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	r := s.run("status")
	r.requireCode(t, 0)
	if !strings.Contains(r.stdout, "local heavy slots: 0/2 held") {
		t.Errorf("R2: untouched local slots must be free: %q", r.stdout)
	}
	for _, line := range strings.Split(r.stdout, "\n") {
		if strings.HasPrefix(line, "alpha ") && !strings.Contains(line, "0/2") {
			t.Errorf("R2: untouched pool slots must be free: %q", line)
		}
	}
}

func TestReviewC3SignalDuringInspection(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	realGit, err := exec.LookPath("git")
	must(t, err)
	marker := filepath.Join(s.base, "inspecting")
	git := filepath.Join(s.base, "bin", "git")
	s.write(git, fmt.Sprintf("#!/bin/bash\nfor arg do if [[ $arg == rev-parse ]]; then touch %q; exec sleep 120; fi; done\nexec %q \"$@\"\n", marker, realGit))
	must(t, os.Chmod(git, 0o700))
	p := s.start("run", "--pool", "--", "touch", "must-not-run")
	p.await("git inspection", func() bool { return exists(marker) })
	p.signal(syscall.SIGINT)
	p.wait(3*time.Second).requireCode(t, 130)
	assertAbsent(t, s.workdir("alpha"), filepath.Join(s.root, "must-not-run"))
}

func TestReviewY6DownloadListFailure(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	r := s.run("run", "--pool", "--", "bash", "-c", "printf result > result; printf corrupt > .git/index")
	if got := s.read(filepath.Join(s.workdir("alpha"), "result")); got != "result" {
		t.Fatalf("command fixture did not produce output: %q", got)
	}
	r.requireCode(t, 125)
}

func TestReviewY8GetExistingDirectory(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.write(".gitignore", "output/\n")
	s.write("output/local-only", "keep\n")
	r := s.run("run", "--get", "output", "--", "bash", "-c", "mkdir -p output; printf result > output/result")
	r.requireCode(t, 0)
	if !exists(filepath.Join(s.root, "output", "result")) {
		t.Errorf("Y8: --get output must return output/result at its original path; nested=%v", exists(filepath.Join(s.root, "output", "output", "result")))
	}
	assertAbsent(t, filepath.Join(s.root, "output", "output"))
}

func TestReviewL4PreviouslyUpHostGoesDown(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false, host{name: "alpha", cpus: 8}, host{name: "beta", cpus: 4})
	release1 := s.holdSlot("beta", 1)
	release2 := s.holdSlot("beta", 2)
	p := s.start("run", "--gpu", "--", "touch", "must-not-run")
	p.await("first round waiting", func() bool { return strings.Contains(p.output().stderr, "offload: waiting") })
	s.write(filepath.Join(s.hosts, "alpha", ".probe"), "invalid probe: host now down\n")
	release1()
	release2()
	select {
	case <-p.done:
		r := p.wait(time.Second)
		t.Fatalf("L4: down alpha must keep waiting; stale no-GPU inventory exited %d: %q", r.code, r.stderr)
	case <-time.After(3 * time.Second):
	}
	p.signal(syscall.SIGINT)
	p.wait(3*time.Second).requireCode(t, 130)
}

func TestReviewS4PoolOwnerNamesHost(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	p := s.start("run", "--pool", "--", "sleep", "120")
	owner := filepath.Join(s.runtime, "offload", "hosts", "alpha", "slot1.owner")
	p.await("remote slot owner", func() bool { return exists(owner) })
	text := s.read(owner)
	p.signal(syscall.SIGINT)
	p.wait(3*time.Second).requireCode(t, 130)
	if !strings.Contains(text, ": alpha: "+s.root+" sleep 120") {
		t.Fatalf("S4: owner must name host then root; got %q", text)
	}
}

func TestReviewY10FailedPrunePreventsCommand(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	s.write(filepath.Join(s.workdir("alpha"), "stale"), "must be removed before command\n")
	realRM, err := exec.LookPath("rm")
	must(t, err)
	rm := filepath.Join(s.base, "bin", "rm")
	s.write(rm, fmt.Sprintf("#!/bin/bash\nif [[ ${!#} == stale ]]; then printf 'rm: injected prune failure\\n' >&2; exit 1; fi\nexec %q \"$@\"\n", realRM))
	must(t, os.Chmod(rm, 0o700))
	r := s.run("run", "--pool", "--", "touch", "must-not-run")
	if !strings.Contains(r.stderr, "injected prune failure") {
		t.Fatalf("prune failure fixture did not fire: %q", r.stderr)
	}
	r.requireCode(t, 125)
	assertAbsent(t, filepath.Join(s.workdir("alpha"), "must-not-run"))
}

func TestReviewX1DeathBeforeSessionCreation(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	markers := filepath.Join(s.base, "pids")
	must(t, os.MkdirAll(markers, 0o700))
	s.markPIDs(markers)
	realSetsid, err := exec.LookPath("setsid")
	must(t, err)
	realPgrep, err := exec.LookPath("pgrep")
	must(t, err)
	release := filepath.Join(s.base, "release-session")
	checks := filepath.Join(s.base, "session-checks")
	setsid := filepath.Join(s.base, "bin", "setsid")
	s.write(setsid, fmt.Sprintf("#!/bin/bash\nprintf '%%s\\n' $$ > %q\nprintf '%%s\\n' $PPID > %q\nwhile [[ ! -e %q ]]; do sleep 0.01; done\nexec %q \"$@\"\n", filepath.Join(markers, "wrapper.pid"), filepath.Join(markers, "watcher.pid"), release, realSetsid))
	must(t, os.Chmod(setsid, 0o700))
	pgrep := filepath.Join(s.base, "bin", "pgrep")
	s.write(pgrep, fmt.Sprintf("#!/bin/bash\n%q \"$@\"\nrc=$?\nprintf 'checked\\n' >> %q\nexit \"$rc\"\n", realPgrep, checks))
	must(t, os.Chmod(pgrep, 0o700))
	leaderPath := filepath.Join(markers, "leader.pid")
	p := s.start("run", "--local", "--", "bash", "-c", "echo $$ > \"$1\"; exec sleep 120", "_", leaderPath)
	p.await("pre-setsid child", func() bool { return exists(filepath.Join(markers, "wrapper.pid")) })
	p.signal(syscall.SIGKILL)
	p.wait(2 * time.Second)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		data, _ := os.ReadFile(checks)
		if strings.Count(string(data), "checked\n") >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	data, err := os.ReadFile(checks)
	must(t, err)
	if strings.Count(string(data), "checked\n") < 2 {
		t.Fatal("watcher did not reach its post-EOF session checks")
	}
	s.write(release, "create session\n")
	deadline = time.Now().Add(2 * time.Second)
	for !exists(leaderPath) && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if exists(leaderPath) {
		pid := readPID(t, leaderPath)
		requirePIDsGone(t, []int{pid}, 13*time.Second)
	}
}

func TestReviewX3SignalDuringStatusRecovery(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	ssh := filepath.Join(s.base, "bin", "ssh")
	marker := filepath.Join(s.base, "reading-status")
	s.write(ssh, fmt.Sprintf("#!/bin/bash\ncase \"${!#}\" in 'cat '*'/status') touch %q; exec sleep 120 ;; esac\nexec %q \"$@\"\n", marker, filepath.Join(fixtureDir, "ssh")))
	must(t, os.Chmod(ssh, 0o700))
	p := s.start("run", "--pool", "--", "bash", "-c", "exit 255")
	p.await("status recovery", func() bool { return exists(marker) })
	p.signal(syscall.SIGINT)
	p.wait(3*time.Second).requireCode(t, 130)
}
