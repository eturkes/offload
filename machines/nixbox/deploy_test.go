package nixbox_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Decision `nixbox security posture`: check runs at every deploy, whatever the nixos-rebuild action.
// deploy runs against stub rsync/ssh/check on PATH; no host is contacted.
func TestDeployRunsCheckAfterEveryAction(t *testing.T) {
	deploy, err := os.ReadFile("deploy")
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"switch", "test", "boot", "build"} {
		t.Run(action, func(t *testing.T) {
			dir := t.TempDir()
			write := func(name, text string, mode os.FileMode) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(dir, name), []byte(text), mode); err != nil {
					t.Fatal(err)
				}
			}
			write("deploy", string(deploy), 0o700)
			write("host.nix", "{\n  address = \"192.0.2.1\";\n}\n", 0o600)
			write("wifi.env", "fixture\n", 0o600)
			write("rsync", "#!/bin/sh\nprintf 'rsync\\n' >>\"$CALLS\"\n", 0o700)
			write("ssh", "#!/bin/sh\nprintf 'ssh:%s\\n' \"$*\" >>\"$CALLS\"\n", 0o700)
			write("check", "#!/bin/sh\nprintf 'check\\n' >>\"$CALLS\"\n", 0o700)
			calls := filepath.Join(dir, "calls")
			cmd := exec.Command(filepath.Join(dir, "deploy"), action)
			cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"), "CALLS="+calls)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("stub deploy %s: %v: %s", action, err, out)
			}
			log, err := os.ReadFile(calls)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(log), "rsync\n") || !strings.Contains(string(log), "nixos-rebuild "+action) {
				t.Fatalf("stub deploy did not reach rsync + nixos-rebuild: %q", log)
			}
			if strings.Count(string(log), "check\n") != 1 {
				t.Fatalf("deploy %s finished without check: %q", action, log)
			}
		})
	}
}
