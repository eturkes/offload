package e2e_test

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Phase-close review (reviewer-4 F02/F03/F04): predicates the suite had left without an effective
// assertion.

func TestX6TransportFailureDuringRunIs125(t *testing.T) {
	s := newSandbox(t, true)
	s.write(filepath.Join(s.hosts, "alpha", ".run-fail"), "")
	marker := filepath.Join(s.base, "ran")
	r := s.run("run", "--pool", "--", "touch", marker)
	r.requireCode(t, 125)
	if exists(marker) {
		t.Fatal("the command ran although its ssh session failed")
	}
	requireContains(t, r.stderr, "simulated connection reset during the run")
}

func TestX6TransportFailureDuringNoSyncRunIs125(t *testing.T) { // no download to fail afterwards
	s := newSandbox(t, true)
	s.write(filepath.Join(s.hosts, "alpha", ".run-fail"), "")
	s.run("run", "--no-sync", "--", "true").requireCode(t, 125)
}

func TestR2StatusShowsProbeNumbers(t *testing.T) {
	s := newSandbox(t, false, host{name: "alpha", cpus: 12, load: 3, gpus: 2, slots: 3})
	r := s.run("status")
	r.requireCode(t, 0)
	var row []string
	for _, line := range strings.Split(r.stdout, "\n") {
		if f := strings.Fields(line); len(f) > 0 && f[0] == "alpha" {
			row = f
		}
	}
	// HOST STATE SLOTS CPUS LOAD BUSY FREE-MEM GPUS; the fake probe reports 1048576 kB = 1.0 GB.
	want := []string{"alpha", "up", "0/3", "12", "3.00", "25%", "1.0GB", "2"}
	if strings.Join(row, " ") != strings.Join(want, " ") {
		t.Fatalf("status row %q, want %q\n%s", row, want, r.stdout)
	}
}

func TestL3DownHostProbedAtMostEvery30s(t *testing.T) {
	s := newSandbox(t, true)
	alpha := filepath.Join(s.hosts, "alpha")
	s.write(filepath.Join(alpha, ".transport-fail"), "")
	must(t, os.Remove(filepath.Join(alpha, ".probe"))) // the fake answers probes from .probe before failing
	p := s.start("run", "--pool", "--", "true")
	p.await("waiting line", func() bool { return strings.Contains(p.output().stderr, "waiting for a free slot") })
	time.Sleep(5 * time.Second) // two more 2 s rounds
	must(t, p.cmd.Process.Signal(syscall.SIGINT))
	p.wait(10*time.Second).requireCode(t, 130)
	calls, err := filepath.Glob(filepath.Join(alpha, ".ssh-call-*"))
	must(t, err)
	if len(calls) != 1 {
		t.Fatalf("down host contacted %d times in ~6 s, want 1 (re-probe at most every 30 s)", len(calls))
	}
}

func TestS4SameTreeRunsSerializeAtController(t *testing.T) {
	s := newSandbox(t, true) // alpha: 2 slots
	first := filepath.Join(s.base, "first-done")
	a := s.start("run", "--pool", "--", "bash", "-c", "sleep 3; touch "+first)
	a.await("first run placed", func() bool { return strings.Contains(a.output().stderr, "alpha slot") })
	b := s.run("run", "--pool", "--", "bash", "-c", "test -e "+first)
	b.requireCode(t, 0) // the second run's command started only after the first finished
	requireContains(t, b.stderr, "waiting for a free slot")
	a.wait(10*time.Second).requireCode(t, 0)
	if _, err := os.Stat(first); err != nil {
		t.Fatal(err)
	}
}
