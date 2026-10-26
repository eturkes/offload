package shell

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestR4X2EmptyPATHSearchesCurrentDirectory(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	command := filepath.Join(dir, "r4-command")
	if err = os.WriteFile(command, []byte("#!/bin/sh\nprintf 'current-directory-command'\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	t.Setenv("PATH", "")
	t.Setenv("BASH_ENV", "")
	oracle, err := exec.Command(bash, "--noprofile", "--norc", "-c", "r4-command").CombinedOutput()
	if err != nil || string(oracle) != "current-directory-command" {
		t.Fatalf("bash oracle: output=%q err=%v", oracle, err)
	}
	path, status, err := Look("r4-command")
	if status != 0 || err != nil {
		t.Fatalf("X2/K2: bash ran the current-directory command with empty PATH, Look returned path=%q status=%d err=%v", path, status, err)
	}
	got, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.Stat(command)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(got, want) {
		t.Fatalf("Look selected %q, want %q", path, command)
	}
}
