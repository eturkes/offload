package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestG1CancellationBeforeSetsidKillsSession(t *testing.T) {
	s := newSandbox(t, false)
	realSetsid, err := exec.LookPath("setsid")
	must(t, err)
	realPgrep, err := exec.LookPath("pgrep")
	must(t, err)
	markers := filepath.Join(s.base, "pids")
	must(t, os.MkdirAll(markers, 0o700))
	s.markPIDs(markers)
	entered := filepath.Join(markers, "launcher.pid")
	checks := filepath.Join(s.base, "empty-session-checks")
	release := filepath.Join(s.base, "release-setsid")
	setsid := filepath.Join(s.base, "bin", "setsid")
	s.write(setsid, fmt.Sprintf("#!/bin/bash\necho $$ > %q\nwhile [[ ! -e %q ]]; do sleep 0.01; done\nexec %q \"$@\"\n", entered, release, realSetsid))
	must(t, os.Chmod(setsid, 0o700))
	pgrep := filepath.Join(s.base, "bin", "pgrep")
	s.write(pgrep, fmt.Sprintf("#!/bin/bash\n%q \"$@\"\nrc=$?\nif [[ $rc == 1 && $1 == -s && -f %q && $2 == $(cat %q) ]]; then printf x >> %q; fi\nexit \"$rc\"\n", realPgrep, entered, entered, checks))
	must(t, os.Chmod(pgrep, 0o700))
	leader := filepath.Join(markers, "leader.pid")
	p := s.start("run", "--local", "--", "bash", "-c", `echo $$ > "$1"; exec sleep 120`, "_", leader)
	p.await("launcher before setsid", func() bool { return exists(entered) })
	pid := readPID(t, entered)
	p.signal(syscall.SIGTERM)
	// On the target, both probes finish before setsid. A corrected watcher may cancel the
	// launcher or wait for its startup handshake instead; either path is valid.
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-p.done:
			p.wait(time.Second).requireCode(t, 143)
			requirePIDsGone(t, []int{pid}, time.Second)
			return
		default:
		}
		b, readErr := os.ReadFile(checks)
		if readErr == nil && len(b) >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	s.write(release, "enter session\n")
	select {
	case <-p.done:
		p.wait(time.Second).requireCode(t, 143)
		requirePIDsGone(t, []int{pid}, time.Second)
	case <-time.After(12 * time.Second):
		stillAlive := syscall.Kill(pid, 0) == nil
		must(t, syscall.Kill(pid, syscall.SIGKILL))
		p.wait(3*time.Second).requireCode(t, 143)
		t.Errorf("X1 startup race: command alive=%v after 12s; watcher checked before setsid and never retried", stillAlive)
	}
}

func TestG3DownloadListFailureIsNotSuccess(t *testing.T) {
	s := newSandbox(t, true)
	r := s.run("run", "--pool", "--", "bash", "-c", "printf changed-output > tracked; printf '[invalid' > .git/config")
	r.requireCode(t, 125)
}

func TestG5UploadTypeChangeKeepsIgnoredRemoteFile(t *testing.T) {
	s := newSandbox(t, true)
	s.write(".gitignore", "cache/\n")
	s.run("run", "--pool", "--", "bash", "-c", "mkdir cache; printf keep > cache/precious").requireCode(t, 0)
	remoteFile := filepath.Join(s.workdir("alpha"), "cache", "precious")
	if got := s.read(remoteFile); got != "keep" {
		t.Fatalf("ignored remote fixture = %q", got)
	}
	s.write("cache", "tracked file replacing remote cache directory\n")
	s.git("add", "cache")
	r := s.run("run", "--pool", "--", "true")
	if !exists(remoteFile) {
		t.Errorf("Y5/README ignored remote file deleted by upload type change; exit=%d stderr=%q", r.code, r.stderr)
	}
}

func TestG7SignalDuringStatusLookupCancels(t *testing.T) {
	s := newSandbox(t, true)
	ssh := filepath.Join(s.base, "bin", "ssh")
	original := ssh + "-original"
	must(t, os.Rename(ssh, original))
	marker := filepath.Join(s.base, "status-lookup.pid")
	s.write(ssh, fmt.Sprintf("#!/bin/bash\nfor arg do\n  if [[ $arg == cat\\ */status ]]; then\n    echo $$ > %q\n    exec sleep 120\n  fi\ndone\nexec %q \"$@\"\n", marker, original))
	must(t, os.Chmod(ssh, 0o700))
	p := s.start("run", "--pool", "--", "bash", "-c", "exit 125")
	p.await("finished-command status lookup", func() bool { return exists(marker) })
	lookupPID := readPID(t, marker)
	p.signal(syscall.SIGTERM)
	select {
	case <-p.done:
		p.wait(time.Second).requireCode(t, 143)
		requirePIDsGone(t, []int{lookupPID}, time.Second)
	case <-time.After(12 * time.Second):
		must(t, syscall.Kill(lookupPID, syscall.SIGKILL))
		p.wait(3*time.Second).requireCode(t, 143)
		t.Error("C3/X3 signal did not cancel status lookup after watcher finished; still blocked after 12s")
	}
}

func TestG7SignalDuringInspectionWins(t *testing.T) {
	s := newSandbox(t, true)
	realGit, err := exec.LookPath("git")
	must(t, err)
	marker := filepath.Join(s.base, "inspection-entered")
	git := filepath.Join(s.base, "bin", "git")
	s.write(git, fmt.Sprintf("#!/bin/bash\nfor arg do\n  if [[ $arg == rev-parse ]]; then\n    printf ready > %q\n    while :; do :; done\n  fi\ndone\nexec %q \"$@\"\n", marker, realGit))
	must(t, os.Chmod(git, 0o700))
	p := s.start("run", "--pool", "--", "true")
	p.await("git inspection", func() bool { return exists(marker) })
	p.signal(syscall.SIGTERM)
	p.wait(3*time.Second).requireCode(t, 143)
}
