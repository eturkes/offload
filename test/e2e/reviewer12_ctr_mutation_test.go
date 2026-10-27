package e2e_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestC4CtrFlagErrorsUseOffloadPrefix(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	r := s.run("ctr", "--bogus", "--", "true")
	r.requireCode(t, 2)
	if !strings.HasPrefix(r.stderr, "offload: ") || strings.Contains(r.stderr, "Usage of ctr:") {
		t.Fatalf("unprefixed ctr diagnostic: %q", r.stderr)
	}
}

func TestK3LayersMissingDefinitionNamesPath(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	r := s.run("ctr", "--layers")
	r.requireCode(t, 2)
	requireContains(t, r.stderr, filepath.Join(s.root, ".devcontainer", "Containerfile"))
}

func TestS6LayersLockFailureNamesDirectory(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.write(".devcontainer/Containerfile", "FROM scratch\n")
	dir := filepath.Join(s.runtime, "offload", "ctr")
	must(t, os.MkdirAll(dir, 0o700))
	must(t, os.Chmod(dir, 0o777))
	reviewFakePodman(t, s, `"image exists") exit 1 ;;`)
	r := s.run("ctr", "--layers")
	r.requireCode(t, 125)
	requireContains(t, r.stderr, dir)
}

func TestK9LayersExistsFailureCannotBecomeSuccess(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.write(".devcontainer/Containerfile", "FROM scratch\n")
	reviewFakePodman(t, s, `"image exists") exit 125 ;;
"image inspect") printf 'sha256:a\n'; exit 0 ;;`)
	r := s.run("ctr", "--layers")
	r.requireCode(t, 125)
	if r.stdout != "" {
		t.Fatalf("failed image setup emitted layers: %q", r.stdout)
	}
}

func TestK3CtrUnlinkedCwdFails(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	r := s.run("ctr", "--", "bash", "-c", `rmdir -- "$PWD" || exit 99
export OFFLOAD_CTR=1
exec "$1" ctr -- true`, "_", binaryPath)
	r.requireCode(t, 125)
}

func TestK10UnreadableBinaryDoesNotInventDifference(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.write(".devcontainer/Containerfile", "FROM scratch\n")
	bin, err := os.ReadFile(binaryPath)
	must(t, err)
	self := filepath.Join(s.base, "unreadable-offload")
	must(t, os.WriteFile(self, bin, 0o700))
	must(t, os.Chmod(self, 0o111))
	must(t, os.WriteFile(filepath.Join(s.base, "bin", "offload"), bin, 0o755))
	if _, err := os.ReadFile(self); !os.IsPermission(err) {
		t.Fatalf("positive control: unreadable binary read = %v, want permission denied", err)
	}
	reviewFakePodman(t, s, `"image exists") exit 0 ;;
"image inspect") printf 'sha256:a\n'; exit 0 ;;`)
	r := s.run("ctr", "--", self, "ctr", "--check")
	// K10: identical bytes cannot differ; a failed comparison cannot claim all-host success.
	if r.code == 0 || strings.Contains(r.stdout, "offload binary differs") {
		t.Fatalf("unreadable identical binary: exit=%d stdout=%q stderr=%q", r.code, r.stdout, r.stderr)
	}
}

func TestK10CheckMissingFrameIsFailed(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.write(".devcontainer/Containerfile", "FROM scratch\n")
	s.write(filepath.Join(s.hosts, "alpha", ".bash_profile"), "echo login-banner\nexit 0\n")
	reviewFakePodman(t, s, `"image exists") exit 0 ;;
"image inspect") printf 'sha256:a\n'; exit 0 ;;`)
	r := s.run("ctr", "--check")
	r.requireCode(t, 1)
	if !strings.HasPrefix(r.stdout, "alpha: failed (") || strings.Count(r.stdout, "\n") != 1 {
		t.Fatalf("missing host frame is not a failed host verdict: %q", r.stdout)
	}
}

func TestK10CheckLaunchFailureNamesHost(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.write(".devcontainer/Containerfile", "FROM scratch\n")
	bin, err := os.ReadFile(binaryPath)
	must(t, err)
	self := filepath.Join(s.base, "vanishing-offload")
	must(t, os.WriteFile(self, bin, 0o755))
	s.setenv("REVIEW12_SELF", self)
	reviewFakePodman(t, s, `"image exists") exit 0 ;;
"image inspect") rm -- "$REVIEW12_SELF"; printf 'sha256:a\n'; exit 0 ;;`)
	r := s.run("ctr", "--", self, "ctr", "--check")
	r.requireCode(t, 1)
	assertAbsent(t, self)
	if !strings.HasPrefix(r.stdout, "alpha: failed (") || strings.Count(r.stdout, "\n") != 1 {
		t.Fatalf("launch failure lost its host verdict: %q", r.stdout)
	}
}
