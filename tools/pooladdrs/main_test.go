package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

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
