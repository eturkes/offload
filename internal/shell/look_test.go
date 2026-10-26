package shell

import (
	"os"
	"path/filepath"
	"testing"
)

// X2 + K2: lookup = bash's (127 / 126).
func TestLookMatchesBash(t *testing.T) {
	dir := t.TempDir()
	write := func(name string, mode os.FileMode) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("#!/bin/sh\n"), mode); err != nil {
			t.Fatal(err)
		}
		return p
	}
	exe, plain := write("exe", 0o755), write("plain", 0o644)
	other := t.TempDir()
	if err := os.WriteFile(filepath.Join(other, "plain"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(filepath.ListSeparator)+other)
	for _, tc := range []struct {
		name, path string
		status     int
	}{
		{"exe", exe, 0},
		{"plain", filepath.Join(other, "plain"), 0}, // a later executable beats an earlier non-executable
		{"missing", "", NotFound},
		{"sub", "", NotFound}, // a directory on PATH is no match
		{exe, exe, 0},
		{plain, plain, NotExecutable},
		{filepath.Join(dir, "sub"), filepath.Join(dir, "sub"), NotExecutable},
		{filepath.Join(dir, "nope"), filepath.Join(dir, "nope"), NotFound},
	} {
		path, status, _ := Look(tc.name)
		if status != tc.status || (tc.path != "" && path != tc.path) {
			t.Errorf("Look(%q) = %q %d, want %q %d", tc.name, path, status, tc.path, tc.status)
		}
	}
	t.Setenv("PATH", dir)
	if path, status, err := Look("plain"); status != NotExecutable || path != plain || err == nil {
		t.Errorf("non-executable PATH match alone = %q %d %v, want %q 126 + error", path, status, err, plain)
	}
}
