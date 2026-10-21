package slots

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestS1HeavyRejectsInvalidCounts(t *testing.T) {
	for _, value := range []string{"0", "bad", "9223372036854775808"} {
		if d, err := Heavy(envOf(map[string]string{"HEAVY_SLOTS": value, "XDG_RUNTIME_DIR": t.TempDir()})); err == nil {
			t.Errorf("S1: HEAVY_SLOTS=%q → %+v without error", value, d)
		}
	}
}

func TestS1RuntimeDirFallback(t *testing.T) {
	if got := RuntimeDir(envOf(nil)); got != "/tmp" {
		t.Fatalf("S1: unset XDG_RUNTIME_DIR → %q, want /tmp", got)
	}
}

func TestS3ReleasePreservesSuccessorOwner(t *testing.T) {
	d := Dir{Path: t.TempDir(), Total: 1}
	first, ok, err := d.TryAcquire(" first")
	if err != nil || !ok {
		t.Fatalf("first acquire = %v, %v", ok, err)
	}
	first.Release()
	second, ok, err := d.TryAcquire(" successor")
	if err != nil || !ok {
		t.Fatalf("successor acquire = %v, %v", ok, err)
	}
	defer second.Release()
	first.Release()
	if holders := d.Holders(); len(holders) != 1 || !strings.Contains(holders[0], ": successor") {
		t.Fatalf("S3: released predecessor erased live owner's text: %q", holders)
	}
}

func TestR2HoldersScanOnlyConfiguredRange(t *testing.T) {
	d := Dir{Path: t.TempDir(), Total: 2}
	first, ok, err := d.TryAcquire(" first")
	if err != nil || !ok {
		t.Fatalf("first acquire = %v, %v", ok, err)
	}
	defer first.Release()
	second, ok, err := d.TryAcquire(" second")
	if err != nil || !ok {
		t.Fatalf("second acquire = %v, %v", ok, err)
	}
	defer second.Release()
	first.Release()
	outside, held, err := TryLock(filepath.Join(d.Path, "slot0"))
	if err != nil || !held {
		t.Fatalf("fixture slot0 = %v, %v", held, err)
	}
	defer func() { _ = outside.Close() }()
	if holders := d.Holders(); len(holders) != 1 || !strings.HasPrefix(holders[0], "slot 2:") || !strings.Contains(holders[0], ": second") {
		t.Fatalf("R2/S1: holder after free slot1, ignoring slot0 = %q", holders)
	}
}

func TestS4WorkdirLockCreatesParents(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "new-host")
	path := filepath.Join(dir, "wd-project.lock")
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
	first, held, err := TryLock(path)
	if err != nil || !held {
		t.Fatalf("S4: new workdir lock = %v, %v", held, err)
	}
	defer func() { _ = first.Close() }()
	second, held, err := TryLock(path)
	if second != nil {
		_ = second.Close()
	}
	if err != nil || held {
		t.Fatalf("S4: held workdir lock = %v, %v", held, err)
	}
}

func TestS3OwnerWriteFailureReleasesSlot(t *testing.T) {
	const helper = "OFFLOAD_TEST_OWNER_WRITE_LIMIT"
	if os.Getenv(helper) == "1" {
		dir := t.TempDir()
		// Process-local file-size limit makes the private regular owner's write fail, not its open/close.
		if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 0, Max: 0}); err != nil {
			t.Fatal(err)
		}
		d := Dir{Path: dir, Total: 1}
		s, held, err := d.TryAcquire(" owner-must-fit")
		if err == nil || held {
			s.Release()
			t.Fatalf("S3: failed owner write admitted command: held=%v err=%v", held, err)
		}
		f, locked, err := TryLock(filepath.Join(dir, "slot1"))
		if f != nil {
			_ = f.Close()
		}
		if err != nil || !locked {
			t.Fatalf("S2/S3: failed owner write leaked its slot: held=%v err=%v", locked, err)
		}
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestS3OwnerWriteFailureReleasesSlot$")
	cmd.Env = append(os.Environ(), helper+"=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("owner write failure helper: %v\n%s", err, out)
	}
}
