package run

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCRR2Y10PruneSortFailurePreventsCommand(t *testing.T) {
	h := newHost(t)
	h.gitWorkdir()
	h.write("offload/w/keep", "keep\n")
	h.write("offload/w/stale", "stale\n")
	bin := t.TempDir()
	sort := filepath.Join(bin, "sort")
	if err := os.WriteFile(sort, []byte("#!/bin/sh\nprintf 'injected sort failure\\n' >&2\nexit 23\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+":"+os.Getenv("PATH"))
	rc, stderr := h.run("git", "R2", 1, frame("R2", "keep"), "touch ran", false)
	if !strings.Contains(stderr, "injected sort failure") {
		t.Fatalf("positive control: sort failure was not exercised: %s", stderr)
	}
	if rc != 125 || h.exists("offload/w/ran") {
		t.Fatalf("Y10: failed prune preparation must stop before CMD: rc=%d ran=%v stderr=%q", rc, h.exists("offload/w/ran"), stderr)
	}
}
