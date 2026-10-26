package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// P1 + P2: every pool-file spelling offload accepts yields its address.
func TestEveryValidPoolSpellingYieldsItsAddress(t *testing.T) {
	// tools/install once grepped `address = "…"` lines: literal strings + quoted keys were skipped.
	path := filepath.Join(t.TempDir(), "pool.toml")
	pool := "[[host]]\nname = 'a'\naddress = 'u@h1'\n\n[[host]]\n\"name\" = \"b\"\n\"address\" = \"h2\"\n\n" +
		"[[host]]\nname = \"c\"\naddress = \"h3\" # trailing comment\n"
	if err := os.WriteFile(path, []byte(pool), 0o600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := run(func(k string) string { return map[string]string{"OFFLOAD_CONFIG": path}[k] }, &out); err != nil {
		t.Fatal(err)
	}
	if out.String() != "u@h1\nh2\nh3\n" {
		t.Fatalf("addresses %q", out.String())
	}
}

// P1 + P3 for tools/install (`addresses=$(go run ./tools/pooladdrs)` under set -e): a pool file
// problem returns its error, so install stops instead of installing on no host.
func TestRunReturnsThePoolFileError(t *testing.T) {
	var out bytes.Buffer
	err := run(func(k string) string { return map[string]string{"OFFLOAD_CONFIG": "/nonexistent/pool.toml"}[k] }, &out)
	if err == nil || out.Len() != 0 {
		t.Fatalf("err %v, output %q; want the missing-file error and no addresses", err, out.String())
	}
}

// main as tools/install runs it: addresses + rc 0 for a valid file; rc ≠ 0 + the reason on stderr
// for a missing one. The test binary re-runs itself as main (OFFLOAD_POOLADDRS_MAIN=1).
func TestMainExitStatusAndMessage(t *testing.T) {
	if os.Getenv("OFFLOAD_POOLADDRS_MAIN") == "1" {
		main()
		return
	}
	valid := filepath.Join(t.TempDir(), "pool.toml")
	if err := os.WriteFile(valid, []byte("[[host]]\nname = \"a\"\naddress = \"h1\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		pool, stdout, stderr string
		ok                   bool
	}{{valid, "h1\n", "", true}, {"/nonexistent/pool.toml", "", "pooladdrs: no pool file", false}} {
		cmd := exec.Command(os.Args[0], "-test.run=^TestMainExitStatusAndMessage$")
		cmd.Env = append(os.Environ(), "OFFLOAD_POOLADDRS_MAIN=1", "OFFLOAD_CONFIG="+c.pool)
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		err := cmd.Run()
		if (err == nil) != c.ok || !strings.HasPrefix(stdout.String(), c.stdout) || !strings.Contains(stderr.String(), c.stderr) {
			t.Errorf("pool %s: err %v stdout %q stderr %q", c.pool, err, stdout.String(), stderr.String())
		}
	}
}
