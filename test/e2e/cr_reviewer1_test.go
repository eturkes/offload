package e2e_test

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// J2/X3: a status returned by the remote watcher remains known when offload itself was signaled.
func TestReviewR1J2CancellationKeepsReturnedCommandStatus(t *testing.T) {
	for _, place := range []string{"--local", "--pool"} {
		t.Run(place, func(t *testing.T) {
			s := newSandbox(t, true)
			p := s.start("run", place, "--", "bash", "-c", "trap 'exit 17' TERM; printf ready; while :; do sleep 1; done")
			p.await("command ready", func() bool { return p.output().stdout == "ready" })
			p.signal(syscall.SIGTERM)
			p.wait(15*time.Second).requireCode(t, 143)
			if place == "--pool" {
				status := s.read(filepath.Join(s.hosts, "alpha", "offload", ".meta", filepath.Base(s.workdir("alpha")), "status"))
				if !strings.HasSuffix(status, " 17\n") {
					t.Fatalf("remote watcher did not return the fixture's 17: %q", status)
				}
			}
			_, end := s.runPair(time.Time{})
			if end["cmd_rc"] != 17.0 {
				t.Fatalf("%s returned command status was discarded: cmd_rc=%v, rc=%v, signal=%v; want cmd_rc=17", place, end["cmd_rc"], end["rc"], end["signal"])
			}
		})
	}
}

// L1/L2/X2: a hybrid run can use the host's login PATH even when CMD is absent locally.
func TestReviewR1L2HybridCommandExistsOnlyOnHost(t *testing.T) {
	s := newSandbox(t, true)
	const name = "offload-review-host-only-command"
	path := filepath.Join(s.hosts, "alpha", ".local", "bin", name)
	s.write(path, "#!/bin/sh\nprintf 'host-only\\n'\n")
	must(t, os.Chmod(path, 0o700))
	s.write(filepath.Join(s.hosts, "alpha", ".bash_profile"), "export PATH=\"$HOME/.local/bin:$PATH\"\n")
	// Positive control: the fake host's login shell does resolve the command.
	control := s.run("run", "--pool", "--", name)
	control.requireCode(t, 0)
	if control.stdout != "host-only\n" {
		t.Fatalf("host control stdout = %q", control.stdout)
	}
	r := s.run("run", "--", name)
	r.requireCode(t, 0)
	if r.stdout != "host-only\n" {
		t.Errorf("hybrid stdout = %q, want host-only", r.stdout)
	}
}

// R2: an unreadable-for-write but unlocked slot is not evidence of a live holder.
func TestReviewR1R2StatusDoesNotTurnIOErrorIntoHolder(t *testing.T) {
	s := newSandbox(t, false)
	slot := filepath.Join(s.runtime, "heavy", "slot1")
	s.write(slot, "")
	must(t, os.Chmod(slot, 0o400))
	defer func() { _ = os.Chmod(slot, 0o600) }()
	if f, err := os.OpenFile(slot, os.O_WRONLY, 0); err == nil {
		_ = f.Close()
		t.Fatal("fixture must refuse write-open")
	}
	// Positive control: Linux flock does not require write-open; the lock is free.
	f, err := os.Open(slot)
	must(t, err)
	must(t, syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB))
	must(t, f.Close())
	r := s.run("status")
	if r.code == 0 && strings.Contains(r.stdout, "local heavy slots: 1/2 held") {
		t.Fatalf("unlocked file's I/O error reported as a held slot: %q", r.stdout)
	}
}
