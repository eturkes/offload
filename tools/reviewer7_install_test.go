package tools_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// R28: a failed checksum command cannot turn two absent digests into a successful comparison.
func TestR7InstallRejectsFailedChecksum(t *testing.T) {
	sha, err := exec.LookPath("sha256sum")
	if err != nil {
		t.Fatal(err)
	}
	goenv, err := exec.Command("go", "env", "GOPATH", "GOMODCACHE", "GOCACHE").Output()
	if err != nil {
		t.Fatal(err)
	}
	dirs := strings.Split(strings.TrimSpace(string(goenv)), "\n")
	bin := t.TempDir()
	for name, body := range map[string]string{
		"sha256sum": "#!/bin/sh\nif [ \"${OFFLOAD_R7_SUM_FAIL:-}\" = 1 ]; then echo 'injected checksum failure' >&2; exit 23; fi\nexec \"" + sha + "\" \"$@\"\n",
		"ssh":       "#!/bin/sh\nsha256sum | cut -d' ' -f1\n",
	} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte(body), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for _, fail := range []string{"0", "1"} {
		t.Run("checksum-fails-"+fail, func(t *testing.T) {
			cmd := exec.Command("sh", "install", "fake-host")
			cmd.Env = append(os.Environ(), "HOME="+t.TempDir(), "PATH="+bin+":"+os.Getenv("PATH"),
				"GOPATH="+dirs[0], "GOMODCACHE="+dirs[1], "GOCACHE="+dirs[2], "OFFLOAD_R7_SUM_FAIL="+fail)
			out, err := cmd.CombinedOutput()
			if fail == "0" {
				if err != nil || !strings.Contains(string(out), "fake-host heavy: ") {
					t.Fatalf("positive control: err=%v output=%q", err, out)
				}
				return
			}
			if !strings.Contains(string(out), "injected checksum failure") {
				t.Fatalf("failure not exercised: err=%v output=%q", err, out)
			}
			if err == nil {
				t.Fatalf("install accepted missing checksums (rc 0): %s", out)
			}
		})
	}
}
