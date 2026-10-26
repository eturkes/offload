package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// X2 + R21: SHELLOPTS belongs to CMD; noexec/onecmd cannot stop the watcher before CMD starts.
func TestR7X2ExportedShelloptsCannotSkipCommand(t *testing.T) {
	touch, err := exec.LookPath("touch")
	must(t, err)
	for _, option := range []string{"noexec", "onecmd"} {
		t.Run(option, func(t *testing.T) {
			s := newSandbox(t, false)
			s.run("run", "--local", "--", touch, "control").requireCode(t, 0)
			if _, err := os.Stat(filepath.Join(s.root, "control")); err != nil {
				t.Fatalf("positive control: touch did not run: %v", err)
			}
			s.setenv("SHELLOPTS", option)
			r := s.run("run", "--local", "--", touch, "ran")
			r.requireCode(t, 0)
			if _, err := os.Stat(filepath.Join(s.root, "ran")); err != nil {
				t.Fatalf("SHELLOPTS=%s: offload returned success without running touch: %v; stderr=%q", option, err, r.stderr)
			}
		})
	}
}
