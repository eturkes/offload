package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// S6: a lock directory component another user could have planted (here: world-writable) → 125,
// before any slot, lock or command.
func plantOpenDir(t *testing.T, s *sandbox, rel string) string {
	t.Helper()
	p := filepath.Join(s.runtime, rel)
	must(t, os.MkdirAll(p, 0o700))
	must(t, os.Chmod(p, 0o777)) // umask-proof
	return p
}

func TestS6PlantedHeavyDirIs125(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	dir := plantOpenDir(t, s, "heavy")
	r := s.run("run", "--local", "--", "touch", "ran")
	r.requireCode(t, 125)
	requireContains(t, r.stderr, dir)
	assertAbsent(t, s.path("ran"), filepath.Join(dir, "slot1"))
}

func TestS6PlantedHostDirIs125(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	dir := plantOpenDir(t, s, filepath.Join("offload", "hosts", "alpha"))
	r := s.run("run", "--pool", "--", "touch", "ran")
	r.requireCode(t, 125)
	requireContains(t, r.stderr, dir)
	assertAbsent(t, filepath.Join(dir, "slot1"), filepath.Join(s.workdir("alpha"), "ran"))
}

func TestS6PlantedCtrLockDirIs125(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.setenv("OFFLOAD_CTR", "1")
	s.write(".devcontainer/Containerfile", "FROM scratch\n")
	reviewFakePodman(t, s, `"image exists") exit 1 ;;`) // build lock reached; any build = exit 98
	dir := plantOpenDir(t, s, filepath.Join("offload", "ctr"))
	r := s.run("ctr", "--", "unused-command")
	r.requireCode(t, 125)
	requireContains(t, r.stderr, dir)
	if entries, err := os.ReadDir(dir); err != nil || len(entries) != 0 {
		t.Fatalf("lock created in the planted dir: %v (err %v)", entries, err)
	}
}

// The core S6 case: a lock directory of another uid. `podman unshare chown` hands it to a subordinate
// uid of ours (no root needed); its cleanup hands it back before TempDir's removal (LIFO).
func TestS6ForeignOwnedHeavyDirIs125(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("podman"); err != nil {
		if os.Getenv("OFFLOAD_REQUIRE_PODMAN") != "" {
			t.Fatal("OFFLOAD_REQUIRE_PODMAN set but podman is missing")
		}
		t.Skip("podman missing: no subordinate uid to own the planted dir")
	}
	s := newSandbox(t, true)
	dir := filepath.Join(s.runtime, "heavy")
	must(t, os.Mkdir(dir, 0o755))
	// The attack: a dir of theirs (0755) holding slot files we may write ⇒ shared locks, read owners.
	for _, f := range []string{"slot1", "slot1.owner", "slot2", "slot2.owner"} {
		s.write(filepath.Join(dir, f), "")
		must(t, os.Chmod(filepath.Join(dir, f), 0o666))
	}
	unshare := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("podman", append([]string{"unshare"}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("podman unshare %q: %v\n%s", args, err, out)
		}
	}
	unshare("chown", "-R", "1:1", dir)
	t.Cleanup(func() { unshare("chown", "-R", "0:0", dir) })
	r := s.run("run", "--local", "--", "touch", "ran")
	r.requireCode(t, 125)
	requireContains(t, r.stderr, dir)
	assertAbsent(t, s.path("ran"))
}
