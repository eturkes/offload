// Package slots acquires advisory command slots: heavy's local ones and offload's per-host ones.
// A slot = flock(2) on a file; heavy's `flock -n` and offload exclude each other on the same file.
package slots

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Slot is one held lock. Release frees it; the kernel frees it if offload dies first.
type Slot struct {
	Index, Total int
	file         *os.File
	owner        string
}

// Dir = a set of slot files: <path>/slot1 … slot<Total>, each with a slot<i>.owner text file.
type Dir struct {
	Path  string
	Total int
}

// RuntimeDir = $XDG_RUNTIME_DIR, else /tmp (heavy's fallback).
func RuntimeDir(getenv func(string) string) string {
	if d := getenv("XDG_RUNTIME_DIR"); d != "" {
		return d
	}
	return "/tmp"
}

var digits = regexp.MustCompile(`^[0-9]+$`)

// HeavySlots parses HEAVY_SLOTS exactly as heavy does: default 2, decimal digits, leading zeros
// dropped, zero or out-of-range refused (heavy's messages).
func HeavySlots(getenv func(string) string) (int, error) {
	v := getenv("HEAVY_SLOTS")
	if v == "" {
		return 2, nil
	}
	if !digits.MatchString(v) {
		return 0, errors.New("HEAVY_SLOTS must be a positive integer")
	}
	// heavy: leading zeros stripped, then `$((slots)) == "$slots"` ⇒ 64-bit range.
	n, err := strconv.ParseInt(strings.TrimLeft(v, "0"), 10, 64)
	if err != nil || n == 0 {
		return 0, errors.New("HEAVY_SLOTS must be a positive shell integer")
	}
	return int(n), nil
}

// Heavy = heavy's slot directory.
func Heavy(getenv func(string) string) (Dir, error) {
	n, err := HeavySlots(getenv)
	if err != nil {
		return Dir{}, err
	}
	return Dir{Path: filepath.Join(RuntimeDir(getenv), "heavy"), Total: n}, nil
}

// Host = offload's slot directory for one pool host.
func Host(getenv func(string) string, name string, total int) Dir {
	return Dir{Path: filepath.Join(RuntimeDir(getenv), "offload", "hosts", name), Total: total}
}

// TryAcquire takes the lowest free slot without blocking; ok = false when every slot is held.
// owner = the text after "<pid> since <HH:MM:SS>:" in the owner file.
func (d Dir) TryAcquire(owner string) (s *Slot, ok bool, err error) {
	if err := os.MkdirAll(d.Path, 0o700); err != nil {
		return nil, false, err
	}
	for i := 1; i <= d.Total; i++ {
		f, held, err := tryLock(filepath.Join(d.Path, "slot"+strconv.Itoa(i)))
		if err != nil {
			return nil, false, err
		}
		if !held {
			continue
		}
		s := &Slot{Index: i, Total: d.Total, file: f, owner: f.Name() + ".owner"}
		line := fmt.Sprintf("%d since %s:%s\n", os.Getpid(), time.Now().Format("15:04:05"), owner)
		if err := writeOwner(s.owner, line); err != nil {
			s.Release()
			return nil, false, err
		}
		return s, true, nil
	}
	return nil, false, nil
}

// Release empties the owner file (heavy's `: > owner`) and drops the lock.
func (s *Slot) Release() {
	if s == nil || s.file == nil {
		return
	}
	_ = writeOwner(s.owner, "")
	_ = s.file.Close()
	s.file = nil
}

// Holders lists "slot <i>: <owner text>" for every held slot (heavy --status format).
func (d Dir) Holders() []string {
	var out []string
	for i := 1; i <= d.Total; i++ {
		p := filepath.Join(d.Path, "slot"+strconv.Itoa(i))
		if _, err := os.Lstat(p); errors.Is(err, os.ErrNotExist) {
			continue // never created ⇒ never held
		}
		f, held, err := tryLock(p)
		if err == nil && held {
			_ = f.Close()
			continue
		}
		text, _ := os.ReadFile(p + ".owner") //nolint:gosec // G304: path built from the runtime dir, not input
		out = append(out, fmt.Sprintf("slot %d: %s", i, strings.TrimRight(string(text), "\n")))
	}
	return out
}

// writeOwner never follows a symlink: under the /tmp fallback another user could plant one at a
// slot path to make offload truncate a file it points at. O_NONBLOCK (no effect on regular files)
// turns a planted FIFO into an open error instead of a hang; every slot open carries both.
func writeOwner(path, text string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0o600) //nolint:gosec // G304: runtime dir path
	if err != nil {
		return err
	}
	_, err = f.WriteString(text)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

// tryLock opens path (created like heavy's `exec {fd}>`) and flocks it non-blocking. held = the
// lock is now ours; a busy lock closes the file and returns held = false.
func tryLock(path string) (f *os.File, held bool, err error) {
	// O_CLOEXEC (Go's default): the command never inherits the lock, so a stray daemon it spawns
	// cannot keep the slot held.
	f, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0o600) //nolint:gosec // G304: runtime dir path
	if err != nil {
		return nil, false, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("flock %s: %w", path, err)
	}
	return f, true, nil
}

// File = the slot's lock fd, for a child that must keep the slot held past offload's own exit.
func (s *Slot) File() *os.File { return s.file }

// WaitLock flocks path (created with its directory, symlinks never followed), polling every
// 200 ms while another holder has it, until ctx ends. The file stays open across polls.
func WaitLock(ctx context.Context, path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0o600) //nolint:gosec // G304: runtime dir path
	if err != nil {
		return nil, err
	}
	for {
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return f, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) {
			_ = f.Close()
			return nil, fmt.Errorf("flock %s: %w", path, err)
		}
		select {
		case <-ctx.Done():
			_ = f.Close()
			return nil, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
}

// TryLock flocks path non-blocking, creating it and its directory; held = false when busy.
func TryLock(path string) (f *os.File, held bool, err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, false, err
	}
	return tryLock(path)
}
