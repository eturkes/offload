package e2e_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// Y5: a NUL in the global excludes leaves the upload list whole.
func TestReviewer9GlobalExcludesNULPreservesUpload(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	excludes := filepath.Join(s.base, "global-excludes")
	s.write(excludes, "*.tmp\x00\n")
	s.git("config", "--global", "core.excludesFile", excludes)
	if got := string(s.git("check-ignore", "out.tmp")); got != "out.tmp\n" {
		t.Fatalf("positive control: git must ignore out.tmp, got %q", got)
	}
	r := s.run("run", "--pool", "--", "cat", "tracked")
	kept := exists(filepath.Join(s.workdir("alpha"), "tracked"))
	if r.code == 125 && r.stdout == "" && kept {
		return // Explicit refusal before prune is safe; silently shifted U is not.
	}
	if r.code != 0 || r.stdout != "initial\n" || !kept {
		t.Fatalf("Y5 valid upload lost: rc=%d stdout=%q stderr=%q remote tracked exists=%v", r.code, r.stdout, r.stderr, kept)
	}
}

// Y5: the excludes record never counts as an upload path.
func TestReviewer9ExcludesRecordIsNotAnUploadPath(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	excludes := filepath.Join(s.base, "global-excludes")
	s.write(excludes, "!stale")
	s.git("config", "--global", "core.excludesFile", excludes)
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	stale := filepath.Join(s.workdir("alpha"), "!stale")
	s.write(stale, "remote-only, nonignored\n")
	r := s.run("run", "--pool", "--", "bash", "-c", "test ! -e '!stale'")
	if r.code != 0 || exists(stale) {
		t.Fatalf("Y5 excludes record became upload path: rc=%d stale exists=%v stderr=%q", r.code, exists(stale), r.stderr)
	}
}

// X5: one multiplex fallback note per run, across hosts.
func TestReviewer9MultiplexFallbackOnceAcrossHosts(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true,
		host{name: "alpha", cpus: 8, load: 1, mux: true},
		host{name: "beta", cpus: 8, load: 2, mux: true},
		host{name: "gamma", cpus: 8, load: 3})
	r := s.run("status")
	r.requireCode(t, 0)
	if got := strings.Count(r.stderr, "connection multiplexing off:"); got != 1 {
		t.Fatalf("fallback notes=%d, want one: %q", got, r.stderr)
	}
	for _, name := range []string{"alpha", "beta", "gamma"} {
		calls := sshCalls(t, s, name)
		if len(calls) != 1 {
			t.Fatalf("host %s: calls=%d, want one", name, len(calls))
		}
		if strings.Contains(calls[0], "ControlMaster") || !strings.Contains(calls[0], "-T\x00") {
			t.Errorf("fallback call %s: %q", name, calls[0])
		}
	}
}
