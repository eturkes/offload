package slots

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// tools/heavy = the heavy script tools/install deploys on every host: the protocol offload must
// interoperate with.

func envOf(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func heavyCmd(t *testing.T, runtime string, slots string, args ...string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command("bash", append([]string{"../../tools/heavy"}, args...)...)
	cmd.Env = append(os.Environ(), "XDG_RUNTIME_DIR="+runtime, "HEAVY_SLOTS="+slots)
	return cmd
}

func TestHeavySlotsAgreesWithHeavy(t *testing.T) { // S1, differential vs heavy itself
	runtime := t.TempDir()
	// Accepted values stay small: heavy --status walks every slot.
	for _, v := range []string{
		"", "1", "2", "007", "0", "00", "abc", "-1", "1.5", " 2", "2 ", "+3",
		"99999999999999999999", "9223372036854775808",
	} {
		err := heavyCmd(t, runtime, v, "--status").Run()
		heavyOK := err == nil
		var exitErr *exec.ExitError
		if err != nil && (!errors.As(err, &exitErr) || exitErr.ExitCode() != 2) {
			t.Fatalf("heavy HEAVY_SLOTS=%q: %v", v, err)
		}
		n, ourErr := HeavySlots(envOf(map[string]string{"HEAVY_SLOTS": v}))
		if (ourErr == nil) != heavyOK {
			t.Errorf("HEAVY_SLOTS=%q: heavy accepts=%v, offload n=%d err=%v", v, heavyOK, n, ourErr)
		}
	}
	if n, err := HeavySlots(envOf(map[string]string{"HEAVY_SLOTS": "007"})); n != 7 || err != nil {
		t.Errorf("007 → %d, %v", n, err)
	}
	if n, err := HeavySlots(envOf(nil)); n != 2 || err != nil {
		t.Errorf("unset → %d, %v", n, err)
	}
	if n, err := HeavySlots(envOf(map[string]string{"HEAVY_SLOTS": "9223372036854775807"})); err != nil || n <= 0 {
		t.Errorf("max int64 → %d, %v", n, err)
	}
}

func waitFor(t *testing.T, what string, ok func() bool) {
	t.Helper()
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if ok() {
			return
		}
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestSlotHeldByHeavyIsBusy(t *testing.T) { // S2
	runtime := t.TempDir()
	h := heavyCmd(t, runtime, "1", "sleep", "30")
	if err := h.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = h.Process.Kill(); _ = h.Wait() })
	owner := filepath.Join(runtime, "heavy", "slot1.owner")
	waitFor(t, "heavy's owner line", func() bool { b, _ := os.ReadFile(owner); return len(b) > 0 })

	d, err := Heavy(envOf(map[string]string{"XDG_RUNTIME_DIR": runtime, "HEAVY_SLOTS": "1"}))
	if err != nil {
		t.Fatal(err)
	}
	if s, ok, acqErr := d.TryAcquire(" x"); ok || acqErr != nil {
		s.Release()
		t.Fatalf("acquired a slot heavy holds (err %v)", acqErr)
	}
	if got := d.Holders(); len(got) != 1 || !strings.Contains(got[0], "sleep 30") {
		t.Fatalf("Holders = %q", got)
	}
	_ = h.Process.Signal(syscall.SIGTERM)
	_ = h.Wait()
	s, ok, err := d.TryAcquire(" x")
	if !ok || err != nil {
		t.Fatalf("slot still busy after heavy exited (err %v)", err)
	}
	s.Release()
}

func TestSlotHeldByOffloadIsBusyToHeavy(t *testing.T) { // S2 + S3
	runtime := t.TempDir()
	d := Dir{Path: filepath.Join(runtime, "heavy"), Total: 1}
	s, ok, err := d.TryAcquire(" make -j 'a b'")
	if !ok || err != nil {
		t.Fatalf("TryAcquire = %v, %v", ok, err)
	}
	defer s.Release()
	if s.Index != 1 || s.Total != 1 {
		t.Fatalf("slot %d/%d", s.Index, s.Total)
	}

	var out bytes.Buffer
	st := heavyCmd(t, runtime, "1", "--status")
	st.Stdout = &out
	if err := st.Run(); err != nil {
		t.Fatal(err)
	}
	want := regexp.MustCompile(`^slot 1: ` + strconv.Itoa(os.Getpid()) + ` since \d\d:\d\d:\d\d: make -j 'a b'\n$`)
	if !want.MatchString(out.String()) {
		t.Fatalf("heavy --status = %q", out.String())
	}
	slot := filepath.Join(runtime, "heavy", "slot1")
	if exec.Command("flock", "-n", slot, "true").Run() == nil {
		t.Fatal("flock -n got the slot offload holds")
	}

	s.Release()
	if b, _ := os.ReadFile(slot + ".owner"); len(b) != 0 {
		t.Fatalf("owner after release = %q", b)
	}
	if err := exec.Command("flock", "-n", slot, "true").Run(); err != nil {
		t.Fatalf("slot still locked after release: %v", err)
	}
}

func TestLowestFreeSlotFirst(t *testing.T) { // S2
	d := Dir{Path: t.TempDir(), Total: 3}
	var held []*Slot
	for want := 1; want <= 3; want++ {
		s, ok, err := d.TryAcquire(" x")
		if !ok || err != nil || s.Index != want {
			t.Fatalf("acquire %d: %+v %v %v", want, s, ok, err)
		}
		held = append(held, s)
	}
	if _, ok, _ := d.TryAcquire(" x"); ok {
		t.Fatal("4th slot of 3")
	}
	held[1].Release()
	s, ok, _ := d.TryAcquire(" x")
	if !ok || s.Index != 2 {
		t.Fatalf("reacquire = %+v %v", s, ok)
	}
	for _, h := range append(held, s) {
		h.Release()
	}
}

func TestCommandNeverInheritsLock(t *testing.T) { // S3
	d := Dir{Path: t.TempDir(), Total: 1}
	s, ok, err := d.TryAcquire(" x")
	if !ok || err != nil {
		t.Fatal(err)
	}
	defer s.Release()
	child := exec.Command("sleep", "5")
	if err = child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = child.Process.Kill(); _ = child.Wait() }()
	fds, err := os.ReadDir("/proc/" + strconv.Itoa(child.Process.Pid) + "/fd")
	if err != nil {
		t.Fatal(err)
	}
	for _, fd := range fds {
		target, _ := os.Readlink("/proc/" + strconv.Itoa(child.Process.Pid) + "/fd/" + fd.Name())
		if strings.HasSuffix(target, "/slot1") {
			t.Fatalf("child holds fd %s → %s", fd.Name(), target)
		}
	}
}

// Helper process for TestKilledHolderFreesSlot: holds a slot until killed.
func TestMain(m *testing.M) {
	if dir := os.Getenv("SLOTS_TEST_HOLD"); dir != "" {
		if _, ok, err := (Dir{Path: dir, Total: 1}).TryAcquire(" held"); !ok || err != nil {
			os.Exit(3)
		}
		time.Sleep(time.Minute)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestKilledHolderFreesSlot(t *testing.T) { // S2 + R2: the kernel frees the lock; stale owner text ignored
	dir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = append(os.Environ(), "SLOTS_TEST_HOLD="+dir)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	d := Dir{Path: dir, Total: 1}
	waitFor(t, "helper's owner line", func() bool { return len(d.Holders()) == 1 })
	_ = cmd.Process.Signal(syscall.SIGKILL)
	_ = cmd.Wait()
	if got := d.Holders(); len(got) != 0 {
		t.Fatalf("Holders after SIGKILL = %q (stale text must not count)", got)
	}
	s, ok, err := d.TryAcquire(" x")
	if !ok || err != nil {
		t.Fatalf("slot busy after holder was killed: %v", err)
	}
	s.Release()
}

func TestSlotFilesNeverFollowSymlinks(t *testing.T) { // security: /tmp fallback symlink plant
	for _, planted := range []string{"slot1", "slot1.owner"} {
		dir := t.TempDir()
		victim := filepath.Join(t.TempDir(), "victim")
		if err := os.WriteFile(victim, []byte("precious"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(victim, filepath.Join(dir, planted)); err != nil {
			t.Fatal(err)
		}
		s, ok, err := Dir{Path: dir, Total: 1}.TryAcquire(" x")
		s.Release()
		if b, _ := os.ReadFile(victim); string(b) != "precious" {
			t.Fatalf("%s planted: victim now %q (ok %v err %v)", planted, b, ok, err)
		}
		if ok || err == nil {
			t.Fatalf("%s planted: acquired through a symlink (err %v)", planted, err)
		}
	}
}

func TestSlotFilesNeverBlockOnPlantedFIFO(t *testing.T) { // security: /tmp fallback FIFO plant
	for _, planted := range []string{"slot1", "slot1.owner", "wait.lock"} {
		dir := t.TempDir()
		if err := syscall.Mkfifo(filepath.Join(dir, planted), 0o600); err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		go func() {
			if planted == "wait.lock" {
				f, err := WaitLock(context.Background(), filepath.Dir(dir), filepath.Join(dir, planted))
				if f != nil {
					_ = f.Close()
				}
				done <- err
				return
			}
			s, ok, err := Dir{Path: dir, Total: 1}.TryAcquire(" x")
			s.Release()
			if ok && err == nil {
				err = errors.New("acquired a FIFO as a slot")
			} else if err == nil {
				err = errors.New("no error")
			}
			done <- err
		}()
		select {
		case err := <-done:
			if err == nil || strings.Contains(err.Error(), "acquired") || err.Error() == "no error" {
				t.Fatalf("%s planted as a FIFO: %v, want an open error", planted, err)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("%s planted as a FIFO: open blocked (no reader)", planted)
		}
	}
}

// S6: under the /tmp fallback another user can pre-create a lock directory; every component below
// the runtime dir must be a real directory of ours without group/other write, else no slot.
func TestS6LockDirsRefusePlantedComponents(t *testing.T) {
	for _, tc := range []struct {
		name  string
		plant func(t *testing.T, base string) // plants under base
		dir   func(env func(string) string) (Dir, error)
	}{
		{"heavy-0777", plantDir("heavy", 0o777), heavyDir},
		{"heavy-0770", plantDir("heavy", 0o770), heavyDir},
		{"heavy-symlink", plantLink("heavy"), heavyDir},
		{"offload-0777", plantDir("offload", 0o777), hostDir},
		{"hosts-symlink", func(t *testing.T, base string) {
			t.Helper()
			if err := os.Mkdir(filepath.Join(base, "offload"), 0o700); err != nil {
				t.Fatal(err)
			}
			plantLink(filepath.Join("offload", "hosts"))(t, base)
		}, hostDir},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := t.TempDir()
			tc.plant(t, base)
			d, err := tc.dir(envOf(map[string]string{"XDG_RUNTIME_DIR": base}))
			if err != nil {
				t.Fatal(err)
			}
			s, ok, err := d.TryAcquire(" x")
			if err == nil || ok {
				s.Release()
				t.Fatalf("TryAcquire under a planted %s: ok=%v err=%v, want refusal", tc.name, ok, err)
			}
			if _, statErr := os.Lstat(filepath.Join(d.Path, "slot1")); !errors.Is(statErr, os.ErrNotExist) {
				t.Errorf("slot1 created under the planted dir (lstat err=%v)", statErr)
			}
		})
	}
}

func TestS6LockDirsAcceptOwnPrivateTree(t *testing.T) { // S6 positive control: 0755 (heavy's mkdir -p) passes
	base := t.TempDir()
	plantDir("heavy", 0o755)(t, base)
	d, err := heavyDir(envOf(map[string]string{"XDG_RUNTIME_DIR": base}))
	if err != nil {
		t.Fatal(err)
	}
	s, ok, err := d.TryAcquire(" x")
	if err != nil || !ok {
		t.Fatalf("own 0755 heavy dir refused: ok=%v err=%v", ok, err)
	}
	s.Release()
	h, err := hostDir(envOf(map[string]string{"XDG_RUNTIME_DIR": base}))
	if err != nil {
		t.Fatal(err)
	}
	s, ok, err = h.TryAcquire(" x")
	if err != nil || !ok {
		t.Fatalf("fresh host dir refused: ok=%v err=%v", ok, err)
	}
	s.Release()
	for _, p := range []string{"offload", "offload/hosts", "offload/hosts/h"} {
		info, err := os.Lstat(filepath.Join(base, p))
		if err != nil || info.Mode().Perm() != 0o700 {
			t.Errorf("%s: created with %v (err %v), want 0700", p, info.Mode(), err)
		}
	}
}

func heavyDir(env func(string) string) (Dir, error) { return Heavy(env) }

func hostDir(env func(string) string) (Dir, error) { return Host(env, "h", 2), nil }

func plantDir(rel string, mode os.FileMode) func(*testing.T, string) {
	return func(t *testing.T, base string) {
		t.Helper()
		p := filepath.Join(base, rel)
		if err := os.Mkdir(p, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(p, mode); err != nil { // umask-proof
			t.Fatal(err)
		}
	}
}

func plantLink(rel string) func(*testing.T, string) {
	return func(t *testing.T, base string) {
		t.Helper()
		target := t.TempDir() // a real 0700 dir of ours: only the symlink itself is wrong
		if err := os.Symlink(target, filepath.Join(base, rel)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestS6MakeDirsBoundsAndBase(t *testing.T) { // S6: dir strictly below base; a missing base = 0700
	root := t.TempDir()
	base := filepath.Join(root, "runtime")
	for _, dir := range []string{base, root, filepath.Join(root, "elsewhere", "x")} {
		if err := MakeDirs(base, dir); err == nil {
			t.Errorf("MakeDirs(%q, %q) accepted a dir not below base", base, dir)
		}
	}
	assertAbsentPath(t, base)
	assertAbsentPath(t, filepath.Join(root, "elsewhere"))
	if err := MakeDirs(base, filepath.Join(base, "heavy")); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Lstat(base); err != nil || info.Mode().Perm() != 0o700 {
		t.Fatalf("created base %v (err %v), want 0700", info.Mode(), err)
	}
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := MakeDirs(filepath.Join(file, "runtime"), filepath.Join(file, "runtime", "heavy")); err == nil {
		t.Fatal("MakeDirs under a regular file succeeded")
	}
}

func assertAbsentPath(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("%s exists (lstat err=%v)", path, err)
	}
}
