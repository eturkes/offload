package e2e_test

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestS3PoolReleaseClearsOwner(t *testing.T) { // S3 + S4: normal pool completion empties the owner's text before exit.
	t.Parallel()
	s := newSandbox(t, false)
	s.run("run", "--no-sync", "--", "true").requireCode(t, 0)
	owner := filepath.Join(s.runtime, "offload", "hosts", "alpha", "slot1.owner")
	if got := s.read(owner); got != "" {
		t.Fatalf("completed pool run left owner text: %q", got)
	}
}

func TestS4AcquisitionErrorClearsEarlierCandidates(t *testing.T) { // S2 + S3 + S4 + C3: real lock-open errors stop work and clear every acquired owner.
	t.Parallel()
	for _, leaf := range []string{"slot", "workdir"} {
		t.Run(leaf, func(t *testing.T) {
			s := newSandbox(t, false, host{name: "alpha", cpus: 8}, host{name: "beta", cpus: 8})
			dir := filepath.Join(s.runtime, "offload", "hosts", "beta")
			name := "slot1"
			if leaf == "workdir" {
				name = "wd-" + filepath.Base(s.workdir("beta")) + ".lock"
			}
			bad := filepath.Join(dir, name)
			must(t, os.MkdirAll(bad, 0o700))
			marker := filepath.Join(s.base, "must-not-run")
			r := s.run("run", "--", "touch", marker)
			r.requireCode(t, 125)
			requireContains(t, r.stderr, bad)
			assertAbsent(t, marker)
			for _, h := range []string{"alpha", "beta"} {
				if leaf == "slot" && h == "beta" {
					continue
				}
				owner := filepath.Join(s.runtime, "offload", "hosts", h, "slot1.owner")
				if got := s.read(owner); got != "" {
					t.Errorf("acquisition failure left %s owner text: %q", h, got)
				}
			}
		})
	}
}

func TestL3WaitingNamesEveryPlace(t *testing.T) { // L3 + R1: the waiting line includes local occupancy, down hosts/reasons, and later busy hosts.
	t.Parallel()
	s := newSandbox(t, false, host{name: "alpha", down: true}, host{name: "beta", cpus: 8})
	for _, h := range []string{"", "beta"} {
		s.holdSlot(h, 1)
		s.holdSlot(h, 2)
	}
	p := s.start("run", "--", "true")
	p.await("waiting line", func() bool { return strings.Contains(p.output().stderr, "offload: waiting") })
	p.signal(syscall.SIGINT)
	r := p.wait(3 * time.Second)
	r.requireCode(t, 130)
	requireContains(t, r.stderr, "local 2/2 held", "alpha down:", "Connection refused", "beta 2/2 held")
}

func TestR2StatusContinuesAfterDownHost(t *testing.T) { // R2: every configured host is shown, and busy percent is 100 × load/CPUs.
	t.Parallel()
	s := newSandbox(t, false, host{name: "alpha", down: true}, host{name: "beta", cpus: 8, load: 12, gpus: 1, slots: 3})
	r := s.run("status")
	r.requireCode(t, 0)
	var rows []string
	for _, line := range strings.Split(r.stdout, "\n") {
		f := strings.Fields(line)
		if len(f) > 0 && (f[0] == "alpha" || f[0] == "beta") {
			rows = append(rows, strings.Join(f, " "))
		}
	}
	if len(rows) != 2 || !strings.HasPrefix(rows[0], "alpha down 0/2 ") || !strings.Contains(rows[0], "Connection refused") || rows[1] != "beta up 0/3 8 12.00 150% 1.0GB 1" {
		t.Fatalf("status rows=%q, want down alpha followed by beta at 150%% busy", rows)
	}
}

func TestS1StatusRejectsInvalidHeavySlots(t *testing.T) { // S1 + C3 + C4: invalid local slot counts remain usage errors in status too.
	t.Parallel()
	s := newSandbox(t, false)
	s.setenv("HEAVY_SLOTS", "not-a-count")
	r := s.run("status")
	r.requireCode(t, 2)
	requireContains(t, r.stderr, "offload: ", "HEAVY_SLOTS")
	if r.stdout != "" {
		t.Fatalf("invalid slot count produced status output: %q", r.stdout)
	}
}
