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

func TestY1GitErrorIsNotPlainMode(t *testing.T) {
	s := newSandbox(t, true)
	realGit, err := exec.LookPath("git")
	must(t, err)
	git := filepath.Join(s.base, "bin", "git")
	s.write(git, fmt.Sprintf("#!/bin/bash\nfor arg do if [[ $arg == rev-parse ]]; then printf 'fatal: injected git error\\n' >&2; exit 128; fi; done\nexec %q \"$@\"\n", realGit))
	must(t, os.Chmod(git, 0o700))
	r := s.run("run", "--pool", "--", "touch", "must-not-run")
	r.requireCode(t, 125)
	requireContains(t, r.stderr, "injected git error")
	assertAbsent(t, s.workdir("alpha"), filepath.Join(s.root, "must-not-run"))
}

func TestY8GetKeepsNewerLocal(t *testing.T) {
	s := newSandbox(t, true)
	s.write(".gitignore", "output\n")
	s.write("output", "controller\n")
	future := time.Now().Add(time.Hour)
	must(t, os.Chtimes(filepath.Join(s.root, "output"), future, future))
	r := s.run("run", "--get", "output", "--", "bash", "-c", "printf remote > output")
	r.requireCode(t, 0)
	if got := s.read("output"); got != "controller\n" {
		t.Errorf("--get --update replaced newer local file: %q", got)
	}
}

func TestY9NoSyncCreatesWorkdir(t *testing.T) {
	s := newSandbox(t, true)
	assertAbsent(t, s.workdir("alpha"))
	r := s.run("run", "--no-sync", "--", "bash", "-c", "test ! -e tracked && test ! -e .git && printf '%s' \"$PWD\"")
	r.requireCode(t, 0)
	if r.stdout != s.workdir("alpha") {
		t.Errorf("--no-sync workdir=%q", r.stdout)
	}
}

func TestY10DownloadFailureOverridesCommand(t *testing.T) {
	for _, code := range []int{0, 2, 255} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			s := newSandbox(t, true)
			s.write(filepath.Join(s.hosts, "alpha", ".download-fail"), "fail\n")
			r := s.run("run", "--pool", "--", "bash", "-c", fmt.Sprintf("printf output > command-ran; exit %d", code))
			r.requireCode(t, 125)
			requireContains(t, r.stderr, "injected download failure")
			if got := s.read(filepath.Join(s.workdir("alpha"), "command-ran")); got != "output" {
				t.Errorf("command did not execute before download failure: %q", got)
			}
			assertAbsent(t, filepath.Join(s.root, "command-ran"))
		})
	}
}

func TestY11LivePredecessorAfterControllerDeath(t *testing.T) {
	s := newSandbox(t, true)
	markers := filepath.Join(s.base, "pids")
	must(t, os.MkdirAll(markers, 0o700))
	s.markPIDs(markers)
	command := `trap 'touch "$1/term"; while [[ ! -e $1/release ]]; do sleep 0.02; done; exit 0' TERM
printf '%s\n' "$$" > "$1/leader.pid"
while :; do sleep 0.02; done`
	first := s.start("run", "--pool", "--", "bash", "-c", command, "_", markers)
	first.await("predecessor pid", func() bool { return exists(filepath.Join(markers, "leader.pid")) })
	pid := readPID(t, filepath.Join(markers, "leader.pid"))
	first.signal(syscall.SIGKILL)
	first.wait(2 * time.Second)
	deadline := time.Now().Add(3 * time.Second)
	for !exists(filepath.Join(markers, "term")) && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if !exists(filepath.Join(markers, "term")) {
		t.Fatal("watcher did not deliver TERM to predecessor")
	}
	s.write("tracked", "successor upload must wait\n")
	before, err := filepath.Glob(filepath.Join(s.hosts, "alpha", ".ssh-call-*"))
	must(t, err)
	second := s.start("run", "--pool", "--", "bash", "-c", "printf successor > successor")
	second.await("successor connection", func() bool {
		calls, err := filepath.Glob(filepath.Join(s.hosts, "alpha", ".ssh-call-*"))
		must(t, err)
		return len(calls) >= len(before)+2
	})
	time.Sleep(200 * time.Millisecond)
	if got := s.read(filepath.Join(s.workdir("alpha"), "tracked")); got != "initial\n" {
		t.Errorf("successor uploaded into live predecessor's workdir: %q", got)
	}
	assertAbsent(t, filepath.Join(s.workdir("alpha"), "successor"))
	s.write(filepath.Join(markers, "release"), "finish teardown\n")
	second.wait(8*time.Second).requireCode(t, 0)
	requirePIDsGone(t, []int{pid}, time.Second)
	if got := s.read("successor"); got != "successor" {
		t.Errorf("successor did not complete: %q", got)
	}
}

func TestL4DownOrBusyGPUHostsWait(t *testing.T) {
	for _, mode := range []string{"down", "busy"} {
		t.Run(mode, func(t *testing.T) {
			s := newSandbox(t, false, host{name: "alpha", cpus: 8}, host{name: "beta", cpus: 4, down: mode == "down"})
			if mode == "busy" {
				s.holdSlot("beta", 1)
				s.holdSlot("beta", 2)
			}
			p := s.start("run", "--gpu", "--", "touch", "must-not-run")
			p.await("waiting with down/busy GPU candidate", func() bool { return strings.Contains(p.output().stderr, "offload: waiting") })
			p.signal(syscall.SIGINT)
			p.wait(3*time.Second).requireCode(t, 130)
			assertAbsent(t, filepath.Join(s.root, "must-not-run"), filepath.Join(s.workdir("alpha"), "must-not-run"))
		})
	}
}
