package remote

import (
	"os"
	"path/filepath"
	"testing"
)

// X5: a socket dir refusing new files → plain connections.
func TestReviewer9UnwritableSocketDirectoryRefused(t *testing.T) {
	rt, err := os.MkdirTemp("/tmp", "r9mux-")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(rt, "offload", "ssh")
	if err = os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o700)
		_ = os.RemoveAll(rt)
	})
	if err = os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	if f, createErr := os.Create(filepath.Join(dir, "write-control")); createErr == nil {
		_ = f.Close()
		t.Fatal("fixture must reject a new file")
	}
	opts, err := Multiplex(func(key string) string {
		if key == "XDG_RUNTIME_DIR" {
			return rt
		}
		return ""
	})
	if err == nil && len(opts) > 0 {
		if f, writeErr := os.Create(filepath.Join(dir, "write-after-setup")); writeErr == nil {
			_ = f.Close()
			return // Repairing the directory before enabling mux is also valid.
		}
	}
	if err == nil || len(opts) != 0 {
		t.Fatalf("X5 unusable socket dir accepted: opts=%q err=%v; want plain fallback or usable repair", opts, err)
	}
}
