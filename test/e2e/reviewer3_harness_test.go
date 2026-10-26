package e2e_test

import (
	"os"
	"path/filepath"
	"testing"
)

// A fully stubbed podman command must not depend on an installed real podman.
func TestR3FullFakePodmanNeedsNoInstalledPodman(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	s := newSandbox(t, false)
	reviewFakePodman(t, s, `"image exists") exit 0 ;;`)
	if _, err := os.Stat(filepath.Join(s.base, "fakepodman", "podman")); err != nil {
		t.Fatal(err)
	}
}
