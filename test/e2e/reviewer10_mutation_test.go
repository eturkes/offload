package e2e_test

import (
	"context"
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

// J2: options describe the parsed request, not merely its eventual placement.
func TestJ2HostAndGPUOptionsRecorded(t *testing.T) {
	s := newSandbox(t, false, host{name: "alpha", cpus: 8, gpus: 1})
	s.run("run", "--host", "alpha", "--gpu", "--no-sync", "--", "true").requireCode(t, 0)
	start, end := s.runPair(time.Time{})
	opts := noOptions()
	opts["pool"], opts["host"], opts["gpu"], opts["no_sync"] = true, "alpha", true, true
	requireField(t, start, "options", opts)
	requireField(t, end, "options", opts)
}

// C3: a vanished cwd is an offload setup failure, not a successful local command.
func TestC3RemovedCwdFailsBeforeCommand(t *testing.T) {
	s := newSandbox(t, false)
	ctx, cancel := context.WithTimeout(context.Background(), hang)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "-c", `rmdir "$PWD"; exec "$1" run --local -- /bin/true`, "_", binaryPath)
	cmd.Dir, cmd.Env = s.root, s.env
	out, err := cmd.CombinedOutput()
	if err == nil || cmd.ProcessState.ExitCode() != 125 {
		t.Fatalf("removed cwd: err=%v rc=%d output=%q, want 125", err, cmd.ProcessState.ExitCode(), out)
	}
	assertAbsent(t, filepath.Join(s.runtime, "heavy", "slot1"))
}

func unreadableTrackedDirectory(t *testing.T, s *sandbox) string {
	t.Helper()
	dir := s.path("locked")
	s.write("locked/file", "tracked\n")
	s.git("add", "locked/file")
	must(t, os.Chmod(dir, 0o600))
	t.Cleanup(func() { must(t, os.Chmod(dir, 0o700)) })
	return dir
}

// Y4: a non-ENOENT lstat failure in U is an error, even if placement could fall back locally.
func TestY4UnreadableUploadEntryIsFailure(t *testing.T) {
	s := newSandbox(t, true, host{name: "alpha", down: true})
	unreadableTrackedDirectory(t, s)
	s.run("run", "--", "/bin/true").requireCode(t, 125)
	assertAbsent(t, filepath.Join(s.runtime, "heavy", "slot1"))
}

// Y3: an unsupported layout takes the local lane; no upload list is needed there.
func TestY3LocalOnlyTreeSkipsUploadFileChecks(t *testing.T) {
	s := newSandbox(t, true)
	s.write(".git/objects/info/alternates", "")
	unreadableTrackedDirectory(t, s)
	r := s.run("run", "--", "/bin/true")
	r.requireCode(t, 0)
	requireContains(t, r.stderr, "offload: local slot ")
	assertAbsent(t, s.workdir("alpha"))
}

// S1/C3: invalid local slot configuration stays invalid for a pool-only request too.
func TestS1InvalidHeavySlotsRejectBeforePlacement(t *testing.T) {
	s := newSandbox(t, false)
	s.setenv("HEAVY_SLOTS", "0")
	s.run("run", "--no-sync", "--", "true").requireCode(t, 2)
	assertAbsent(t, s.workdir("alpha"))
}

// C3/J2: failing to start the watcher never fabricates a command status.
func TestC3LocalWatcherStartFailureIsNotCommandSuccess(t *testing.T) {
	s := newSandbox(t, false)
	s.setenv("PATH", filepath.Join(s.base, "bin")) // /bin/true resolves; bash does not.
	s.run("run", "--local", "--", "/bin/true").requireCode(t, 125)
	_, end := s.runPair(time.Time{})
	requireField(t, end, "cmd_rc", nil)
}

// J2/X6: a failed run-phase transport is not the command's 255.
func TestJ2RemoteSetupFailureHasNoCommandStatus(t *testing.T) {
	s := newSandbox(t, false)
	s.write(filepath.Join(s.hosts, "alpha", ".run-fail"), "fail\n")
	s.run("run", "--no-sync", "--", "true").requireCode(t, 125)
	_, end := s.runPair(time.Time{})
	requireField(t, end, "cmd_rc", nil)
}

// J2/X3/Y10: canceled remote 125/255 stay ambiguous; 254 is known; downloads never start.
func TestJ2SignalClassifiesRemoteStatus(t *testing.T) {
	for _, status := range []int{125, 254, 255} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			s := newSandbox(t, true)
			p := s.start("run", "--pool", "--", "bash", "-c", fmt.Sprintf("trap 'exit %d' TERM; printf ready; while :; do sleep 0.02; done", status))
			p.await("command ready", func() bool { return p.output().stdout == "ready" })
			p.signal(syscall.SIGTERM)
			p.wait(hang).requireCode(t, 143)
			_, end := s.runPair(time.Time{})
			var want any
			if status == 254 {
				want = float64(status)
			}
			requireField(t, end, "cmd_rc", want)
			requireField(t, end, "download", nil)
		})
	}
}

// Y8/Y10: a later successful --get cannot turn an earlier transfer failure into success.
func TestY10EarlierGetFailureSurvivesLaterSuccess(t *testing.T) {
	s := newSandbox(t, false)
	s.run("run", "--no-sync", "--get", "missing", "--get", "present", "--", "bash", "-c", "printf output > present").requireCode(t, 125)
}

// X2: a non-executable local command is rejected before taking a slot.
func TestX2NonExecutableLookupBeforeSlot(t *testing.T) {
	s := newSandbox(t, false)
	s.write("noexec", "#!/bin/sh\nexit 0\n")
	s.run("run", "--local", "--", "./noexec").requireCode(t, 126)
	assertAbsent(t, filepath.Join(s.runtime, "heavy", "slot1"))
}

// L1/C4: a non-executable local candidate explains why a hybrid request waits only for hosts.
func TestL1NonExecutableLocalCommandExplainsPoolOnly(t *testing.T) {
	s := newSandbox(t, false)
	s.write("noexec", "#!/bin/sh\nexit 0\n")
	s.holdSlot("alpha", 1)
	s.holdSlot("alpha", 2)
	p := s.start("run", "--", "./noexec")
	p.await("pool-only waiting", func() bool { return strings.Contains(p.output().stderr, "offload: waiting") })
	p.signal(syscall.SIGTERM)
	r := p.wait(hang)
	r.requireCode(t, 143)
	requireContains(t, r.stderr, "offload: ./noexec: permission denied; pool hosts only")
	assertAbsent(t, filepath.Join(s.runtime, "heavy", "slot1"))
}

// Probe barrier: mutate local permissions before upload starts, then keep the source quiescent.
func pauseProbe(s *sandbox) (ready, release string) {
	s.t.Helper()
	ssh := filepath.Join(s.base, "bin", "ssh")
	wrappedSSH := ssh + "-real"
	must(s.t, os.Rename(ssh, wrappedSSH))
	ready, release = filepath.Join(s.base, "probe-ready"), filepath.Join(s.base, "probe-release")
	s.write(ssh, fmt.Sprintf("#!/bin/bash\nif [[ ${!#} == nproc\\;* ]]; then\n  touch %q\n  until [[ -e %q ]]; do sleep 0.01; done\nfi\nexec %q \"$@\"\n", ready, release, wrappedSSH))
	must(s.t, os.Chmod(ssh, 0o700))
	return ready, release
}

// Y4/Y10: U is checked again after placement waits; a stale mirror cannot mask a failed upload.
func TestY4UploadRechecksEntriesAfterWaiting(t *testing.T) {
	s := newSandbox(t, true)
	s.write("locked/file", "tracked\n")
	s.git("add", "locked/file")
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	ready, release := pauseProbe(s)
	p := s.start("run", "--pool", "--", "touch", "must-not-run")
	p.await("probe barrier", func() bool { return exists(ready) })
	dir := s.path("locked")
	must(t, os.Chmod(dir, 0o600))
	t.Cleanup(func() { must(t, os.Chmod(dir, 0o700)) })
	s.write(release, "go\n")
	p.wait(hang).requireCode(t, 125)
	assertAbsent(t, filepath.Join(s.workdir("alpha"), "must-not-run"))
}

// Y5/Y10: unreadable global excludes are an upload failure, not an empty excludes record.
func TestY5UnreadableGlobalExcludesStopsRun(t *testing.T) {
	s := newSandbox(t, true)
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	excludes := filepath.Join(s.base, "excludes")
	s.write(excludes, "*.ignored\n")
	s.git("config", "core.excludesFile", excludes)
	must(t, os.Chmod(excludes, 0))
	t.Cleanup(func() { must(t, os.Chmod(excludes, 0o600)) })
	s.run("run", "--pool", "--", "touch", "must-not-run").requireCode(t, 125)
	assertAbsent(t, filepath.Join(s.workdir("alpha"), "must-not-run"))
}

// J2: the upload duration includes time inside the transfer, not only timer construction.
func TestJ2UploadDurationIncludesTransfer(t *testing.T) {
	s := newSandbox(t, false)
	ssh := filepath.Join(s.base, "bin", "ssh")
	wrappedSSH := ssh + "-real"
	must(t, os.Rename(ssh, wrappedSSH))
	s.write(ssh, fmt.Sprintf("#!/bin/bash\nif [[ $* == *rsync* && $* != *--sender* ]]; then sleep 0.1; fi\nexec %q \"$@\"\n", wrappedSSH))
	must(t, os.Chmod(ssh, 0o700))
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	_, end := s.runPair(time.Time{})
	duration, ok := end["upload"].(float64)
	if !ok || duration < 0.1 {
		t.Fatalf("upload=%v, want at least the transfer's 0.1 seconds", end["upload"])
	}
}
