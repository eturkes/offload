package ctr

import (
	"crypto/sha1"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"
)

// tree writes files (path → content; a "->" prefix = symlink target, a trailing "/" = empty dir)
// under a new dir with mode 0644 (0755 for dirs) and returns it.
func tree(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for path, content := range files {
		full := filepath.Join(dir, path)
		if strings.HasSuffix(path, "/") {
			must(t, os.MkdirAll(full, 0o755))
			continue
		}
		must(t, os.MkdirAll(filepath.Dir(full), 0o755))
		if target, ok := strings.CutPrefix(content, "->"); ok {
			must(t, os.Symlink(target, full))
			continue
		}
		must(t, os.WriteFile(full, []byte(content), 0o644))
		must(t, os.Chmod(full, 0o644)) // umask-independent
	}
	must(t, filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() || path == dir {
			return err
		}
		return os.Chmod(path, 0o755) // umask-independent: a 077 umask leaves 0700 = "dir mode"'s target
	}))
	return dir
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func tag(t *testing.T, dir string) string {
	t.Helper()
	got, err := Tag(dir)
	must(t, err)
	if !strings.HasPrefix(got, Repo+":") || len(got) != len(Repo)+1+16 {
		t.Fatalf("tag %q: want %s:<16 hex>", got, Repo)
	}
	return got
}

func TestK4TagEqualForEqualTrees(t *testing.T) {
	files := map[string]string{"Containerfile": "FROM scratch\n", "a/b": "x", "a.txt": "y", "link": "->a/b", "empty/": ""}
	one, two := tree(t, files), tree(t, files)
	// Owners can't change here; mtimes + the tree's location must not count.
	must(t, os.Chtimes(filepath.Join(two, "a/b"), time.Unix(1, 0), time.Unix(1, 0)))
	if tag(t, one) != tag(t, two) {
		t.Fatal("equal trees in different places/mtimes got different tags")
	}
}

func TestK4TagChangesWithEveryInput(t *testing.T) {
	base := map[string]string{"Containerfile": "FROM scratch\n", "a/b": "x", "link": "->a/b", "empty/": ""}
	ref := tag(t, tree(t, base))
	variants := map[string]func(dir string){
		"content":       func(d string) { must(t, os.WriteFile(filepath.Join(d, "a/b"), []byte("z"), 0o644)) },
		"rename":        func(d string) { must(t, os.Rename(filepath.Join(d, "a/b"), filepath.Join(d, "a/c"))) },
		"new file":      func(d string) { must(t, os.WriteFile(filepath.Join(d, "new"), nil, 0o644)) },
		"new empty dir": func(d string) { must(t, os.Mkdir(filepath.Join(d, "dir2"), 0o755)) },
		"x bit":         func(d string) { must(t, os.Chmod(filepath.Join(d, "a/b"), 0o744)) },
		"other x bit":   func(d string) { must(t, os.Chmod(filepath.Join(d, "a/b"), 0o654)) },
		"read bits":     func(d string) { must(t, os.Chmod(filepath.Join(d, "a/b"), 0o600)) },
		"setgid":        func(d string) { must(t, os.Chmod(filepath.Join(d, "a/b"), 0o644|os.ModeSetgid)) },
		"dir mode":      func(d string) { must(t, os.Chmod(filepath.Join(d, "a"), 0o700)) },
		"link target": func(d string) {
			must(t, os.Remove(filepath.Join(d, "link")))
			must(t, os.Symlink("Containerfile", filepath.Join(d, "link")))
		},
		"link to file": func(d string) {
			must(t, os.Remove(filepath.Join(d, "link")))
			must(t, os.WriteFile(filepath.Join(d, "link"), []byte("a/b"), 0o644))
		},
	}
	for name, change := range variants {
		dir := tree(t, base)
		change(dir)
		if tag(t, dir) == ref {
			t.Errorf("%s: tag unchanged", name)
		}
	}
	// Framing: moving bytes between path and content must not collide.
	if tag(t, tree(t, map[string]string{"ab": "c"})) == tag(t, tree(t, map[string]string{"a": "bc"})) {
		t.Error("ab=c and a=bc share a tag")
	}
}

func TestK4TagRefusesOtherFileTypes(t *testing.T) {
	dir := tree(t, map[string]string{"Containerfile": "FROM scratch\n"})
	must(t, syscall.Mkfifo(filepath.Join(dir, "fifo"), 0o644))
	if got, err := Tag(dir); err == nil {
		t.Fatalf("Tag with a FIFO = %q, want an error", got)
	}
}

func TestK4TagRefusesSymlinkedRoot(t *testing.T) {
	target := tree(t, map[string]string{"Containerfile": "FROM scratch\n"})
	link := filepath.Join(t.TempDir(), "ctx")
	must(t, os.Symlink(target, link))
	if got, err := Tag(link); err == nil {
		t.Fatalf("Tag(symlinked root) = %q, want an error (its content would go unhashed)", got)
	}
}

func TestK6RunArgs(t *testing.T) {
	p := Project{Root: "/r", Cwd: "/r/sub", Binary: "/bin/offload", Context: "/r/.devcontainer"}
	got := RunArgs(p, "localhost/offload-ctr:0123456789abcdef", "/h", []string{"cmd", "a b", "--x"})
	want := []string{
		"run", "--rm", "--pull=never", "--pid=private", "--userns=keep-id", "--log-driver=none", "--http-proxy=false",
		"-i", "--preserve-fds=1",
		"--entrypoint", "/.offload/init", "-v", "/bin/offload:/.offload/init:ro", "-v", "/r:/r",
		"-v", "/h:/.offload/home", "-e", "HOME=/.offload/home",
		"--passwd-entry", "$USERNAME:*:$UID:$GID:$NAME:/.offload/home:/bin/sh", "-w", "/r/sub",
		"localhost/offload-ctr:0123456789abcdef", "__ctr-init", "cmd", "a b", "--x",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("RunArgs =\n%q\nwant\n%q", got, want)
	}
	p.CommonDir = "/main/.git"
	got = RunArgs(p, "t", "/h", []string{"cmd"})
	if i := slices.Index(got, "/main/.git:/main/.git"); i < 1 || got[i-1] != "-v" {
		t.Fatalf("worktree common dir not mounted at its own path: %q", got)
	}
}

func TestK6Home(t *testing.T) {
	env := map[string]string{"XDG_CACHE_HOME": "/c"}
	got, err := Home(func(k string) string { return env[k] }, "/p/my proj")
	must(t, err)
	sum := sha1.Sum([]byte("/p/my proj"))
	if want := fmt.Sprintf("/c/offload/ctr/my_proj-%x/home", sum[:4]); got != want {
		t.Fatalf("Home = %q, want %q (Y2 naming)", got, want)
	}
	other, err := Home(func(k string) string { return env[k] }, "/q/my proj")
	must(t, err)
	if other == got {
		t.Fatal("two roots with one basename share a HOME")
	}
}

func TestK3MountableRefusesColons(t *testing.T) {
	if err := mountable("/a b", "/c,d"); err != nil {
		t.Fatalf("spaces + commas are fine for -v: %v", err)
	}
	if err := mountable("/ok", "/x:y"); err == nil || !strings.Contains(err.Error(), "/x:y") {
		t.Fatalf("mountable with ':' = %v", err)
	}
}

func TestK6RunningBinaryIdentity(t *testing.T) {
	self, err := Executable()
	must(t, err)
	if err = running(self); err != nil {
		t.Fatalf("own binary: %v", err)
	}
	other := filepath.Join(t.TempDir(), "copy")
	data, err := os.ReadFile(self)
	must(t, err)
	must(t, os.WriteFile(other, data, 0o755))
	if err := running(other); err == nil {
		t.Fatal("a copy at another inode passed as the running binary")
	}
}

func TestK7ReadRecord(t *testing.T) {
	for in, want := range map[string]int{"0\n": 0, "7\n": 7, "143\n": 143, "255\n": 255} {
		if got, ok := sendRecord(t, in); !ok || got != want {
			t.Errorf("%q → %d %v, want %d", in, got, ok, want)
		}
	}
	for _, in := range []string{"", "7", "x\n", "-1\n", "1 2\n"} {
		if got, ok := sendRecord(t, in); ok {
			t.Errorf("%q → %d, want no record", in, got)
		}
	}
}

func sendRecord(t *testing.T, in string) (int, bool) {
	t.Helper()
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM|syscall.SOCK_CLOEXEC, 0)
	must(t, err)
	mine, theirs := os.NewFile(uintptr(fds[0]), "mine"), os.NewFile(uintptr(fds[1]), "theirs")
	defer func() { _ = mine.Close() }()
	_, err = theirs.WriteString(in)
	must(t, err)
	must(t, theirs.Close())
	return readRecord(mine)
}

func TestK7MissingRecordWithOpenPeerReturns(t *testing.T) {
	// podman gone while the container still holds its end: no record, no EOF — the read must end.
	mine, theirs, err := controlPair()
	must(t, err)
	defer func() { _ = mine.Close(); _ = theirs.Close() }()
	done := make(chan bool, 1)
	go func() { _, ok := readRecord(mine); done <- ok }()
	select {
	case ok := <-done:
		if ok {
			t.Fatal("record read from a silent peer")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("readRecord blocked on an open, silent peer")
	}
}
