package tools_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// tools/install compares each side's sha256 with the built file's: a host reporting another sum
// fails the install (rc 1). ssh = a stub that hashes what it receives, or lies for host "bad"; HOME
// = a temp dir, so nothing real is installed.
func TestInstallFailsWhenASideReportsAnotherSum(t *testing.T) {
	stubs := t.TempDir()
	ssh := "#!/bin/sh\nwhile [ \"$1\" != -- ]; do shift; done\nsum=$(sha256sum | cut -d' ' -f1)\n" +
		"[ \"$2\" != bad ] || sum=0000\necho \"$sum\"\n"
	if err := os.WriteFile(filepath.Join(stubs, "ssh"), []byte(ssh), 0o700); err != nil {
		t.Fatal(err)
	}
	// The temp HOME must not move Go's module + build caches into it (read-only files, refetches).
	goenv, err := exec.Command("go", "env", "GOPATH", "GOMODCACHE", "GOCACHE").Output()
	if err != nil {
		t.Fatal(err)
	}
	dirs := strings.Split(strings.TrimSpace(string(goenv)), "\n")
	for _, c := range []struct {
		hosts []string
		rc    int
	}{{[]string{"good"}, 0}, {[]string{"good", "bad"}, 1}} {
		cmd := exec.Command("sh", append([]string{"install"}, c.hosts...)...)
		cmd.Env = append(os.Environ(), "HOME="+t.TempDir(), "PATH="+stubs+":"+os.Getenv("PATH"),
			"GOPATH="+dirs[0], "GOMODCACHE="+dirs[1], "GOCACHE="+dirs[2])
		out, err := cmd.CombinedOutput()
		rc := 0
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			rc = ee.ExitCode()
		} else if err != nil {
			t.Fatal(err)
		}
		if rc != c.rc || (c.rc == 1) != strings.Contains(string(out), "bad offload: 0000 (differs") {
			t.Errorf("hosts %v: rc %d, want %d; output %q", c.hosts, rc, c.rc, out)
		}
	}
}
