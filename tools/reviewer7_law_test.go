package tools_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// R4/R5/R6/R8: the law check must reject its committed seeds and accept a clean fixture.
func TestR7LawSeeds(t *testing.T) {
	for _, c := range []struct {
		name, path, text string
		want             int
	}{
		{name: "clean", want: 0},
		{"frontmatter", ".claude/rules/seed.md", "---\npaths:\n  - machines/**\n---\n", 1},
		{"filler", "README.md", "simply\n", 1},
		{"deferred", ".agent/deferred.md", "- seed\n", 1},
		{"test-ids", "internal/cli/seed_test.go", "package cli\nimport \"testing\"\nfunc TestSeed(t *testing.T) {}\n", 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			rc, out := r7Law(t, c.path, c.text)
			if rc != c.want {
				t.Fatalf("law rc=%d, want %d; output=%q", rc, c.want, out)
			}
		})
	}
}

// R4: a double-quoted glob with an undefined YAML escape is not parseable frontmatter.
func TestR7LawRejectsUnparseableFrontmatter(t *testing.T) {
	rc, out := r7Law(t, ".claude/rules/seed.md", "---\npaths:\n  - \"internal/\\q/**\"\n---\n")
	if rc != 1 {
		t.Fatalf("invalid YAML escape \\q: law rc=%d, want 1; output=%q", rc, out)
	}
}

func r7Law(t *testing.T, path, text string) (int, string) {
	t.Helper()
	law, err := os.ReadFile("../.agent/law")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	files := map[string]string{
		".agent/law":            string(law),
		".agent/contract.md":    "- C1 contract\n",
		".agent/deferred.md":    "- entry. Accept: a check.\n",
		".claude/rules/seed.md": "---\npaths:\n  - \"internal/**\"\n---\n",
		"README.md":             "Read this.\n",
		"internal/cli/cli.go":   "package cli\n",
		"internal/cli/base_test.go": "package cli\nimport \"testing\"\n" +
			"func TestC1Base(t *testing.T) {}\n",
		"tools/heavy": "#!/bin/sh\n",
	}
	if path != "" {
		files[path] = text
	}
	for name, data := range files {
		p := filepath.Join(dir, name)
		if err = os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(p, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	init := exec.Command("git", "init", "-q", dir)
	if out, initErr := init.CombinedOutput(); initErr != nil {
		t.Fatalf("git init: %v: %s", initErr, out)
	}
	cmd := exec.Command("bash", filepath.Join(dir, ".agent/law"))
	out, err := cmd.CombinedOutput()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode(), string(out)
	}
	if err != nil {
		t.Fatal(err)
	}
	return 0, string(out)
}
