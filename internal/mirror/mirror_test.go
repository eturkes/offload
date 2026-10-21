package mirror

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/eturkes/offload/internal/remote"
)

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t",
		"GIT_COMMITTER_EMAIL=t@t", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %v\n%s", args, err, out)
	}
	return string(out)
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func repo(t *testing.T) string {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	root := t.TempDir()
	run(t, root, "git", "init", "-q")
	return root
}

func TestUploadList(t *testing.T) { // Y4
	root := repo(t)
	write(t, filepath.Join(root, ".gitignore"), "*.log\n!keep.log\nbuild/\nsecret.txt\n")
	write(t, filepath.Join(root, "keep.log"), "k")
	write(t, filepath.Join(root, "forced.log"), "f")
	write(t, filepath.Join(root, "noise.log"), "n")
	write(t, filepath.Join(root, "build", "out.bin"), "b")
	write(t, filepath.Join(root, "sub", ".gitignore"), "local-only\n.gitignore\n") // self-ignored
	write(t, filepath.Join(root, "sub", "local-only"), "x")
	write(t, filepath.Join(root, "sub", "code.go"), "c")
	write(t, filepath.Join(root, "with space.txt"), "s")
	write(t, filepath.Join(root, "new\nline.txt"), "n")
	write(t, filepath.Join(root, "gone.txt"), "g")
	write(t, filepath.Join(root, "real", "file"), "r")
	if err := os.Symlink("real", filepath.Join(root, "alias")); err != nil {
		t.Fatal(err)
	}
	run(t, root, "git", "add", ".gitignore", "keep.log", "sub/code.go", "gone.txt", "real/file")
	run(t, root, "git", "add", "-f", "forced.log", "alias")
	run(t, root, "git", "commit", "-qm", "init")
	if err := os.Remove(filepath.Join(root, "gone.txt")); err != nil { // deleted, unstaged
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "untracked.txt"), "u")

	got, err := UploadList(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(got)
	want := []string{
		".gitignore", "alias", "forced.log", "keep.log", "new\nline.txt", "real/file",
		"sub/.gitignore", "sub/code.go", "untracked.txt", "with space.txt",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("UploadList =\n%q\nwant\n%q", got, want)
	}

	// A path reached through a symlinked directory is outside the tree.
	if err = os.RemoveAll(filepath.Join(root, "real")); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink("alias", filepath.Join(root, "real")); err != nil { // real → alias → real: a loop
		t.Fatal(err)
	}
	got, err = UploadList(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(got, "real/file") || !slices.Contains(got, "real") {
		t.Fatalf("symlinked ancestor: %q", got)
	}
}

func TestUploadListRefusesNestedRepo(t *testing.T) { // Y3
	root := repo(t)
	write(t, filepath.Join(root, "a"), "a")
	run(t, root, "git", "add", "a")
	run(t, root, "git", "commit", "-qm", "init")
	nested := filepath.Join(root, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	run(t, nested, "git", "init", "-q")
	write(t, filepath.Join(nested, "f"), "f")
	if _, err := UploadList(context.Background(), root); !errors.Is(err, ErrLocalOnly) {
		t.Fatalf("err = %v, want ErrLocalOnly", err)
	}
}

func TestInspect(t *testing.T) { // Y1 + Y3
	ctx := context.Background()
	plain := t.TempDir()
	tr, err := Inspect(ctx, plain)
	if err != nil || tr.Git || tr.Root != plain || tr.Sub != "." {
		t.Fatalf("plain: %+v %v", tr, err)
	}

	root := repo(t)
	write(t, filepath.Join(root, "x", "y", "f"), "f")
	run(t, root, "git", "add", ".")
	run(t, root, "git", "commit", "-qm", "init")
	tr, err = Inspect(ctx, filepath.Join(root, "x", "y"))
	if err != nil || !tr.Git || tr.Sub != "x/y" || tr.LocalOnly != "" {
		t.Fatalf("subdir: %+v %v", tr, err)
	}
	rootReal, _ := filepath.EvalSymlinks(root)
	if tr.Root != rootReal {
		t.Fatalf("root %q, want %q", tr.Root, rootReal)
	}

	wt := filepath.Join(t.TempDir(), "wt")
	run(t, root, "git", "worktree", "add", "-q", wt)
	common := filepath.Join(rootReal, ".git")
	if tr, err = Inspect(ctx, wt); err != nil || tr.LocalOnly != "" || tr.Worktree != "wt" || tr.CommonDir != common || !tr.Git {
		t.Fatalf("worktree: %+v %v; want eligible, id wt, common dir %s", tr, err, common)
	}
	odd := filepath.Join(t.TempDir(), "w[1]") // git's id = "w-1]": rsync filters + remote paths take [A-Za-z0-9._-] alone
	run(t, root, "git", "worktree", "add", "-q", "-b", "odd", odd)
	if tr, err = Inspect(ctx, odd); err != nil || !strings.Contains(tr.LocalOnly, "worktree id") || tr.Worktree != "" {
		t.Fatalf("odd worktree id: %+v %v", tr, err)
	}
	if tr, err = Inspect(ctx, root); err != nil || tr.Worktree != "" || tr.LocalOnly != "" {
		t.Fatalf("main worktree: %+v %v", tr, err)
	}

	sep := t.TempDir()
	run(t, sep, "git", "init", "-q", "--separate-git-dir", filepath.Join(t.TempDir(), "meta.git"))
	if tr, err = Inspect(ctx, sep); err != nil || tr.LocalOnly == "" {
		t.Fatalf("separate git dir: %+v %v", tr, err)
	}

	alt := filepath.Join(t.TempDir(), "alt")
	run(t, root, "git", "clone", "-q", "--shared", root, alt)
	if tr, err = Inspect(ctx, alt); err != nil || !strings.Contains(tr.LocalOnly, "alternates") {
		t.Fatalf("alternates: %+v %v", tr, err)
	}

	sub := repo(t)
	write(t, filepath.Join(sub, "f"), "f")
	run(t, sub, "git", "add", "f")
	run(t, sub, "git", "commit", "-qm", "init")
	run(t, sub, "git", "-c", "protocol.file.allow=always", "submodule", "add", "-q", root, "mod")
	if tr, err = Inspect(ctx, sub); err != nil || !strings.Contains(tr.LocalOnly, "submodules") {
		t.Fatalf("submodule: %+v %v", tr, err)
	}
}

func TestInspectRefusesBrokenRepo(t *testing.T) { // Y1: a git error is never plain mode
	root := repo(t)
	if err := os.WriteFile(filepath.Join(root, ".git", "HEAD"), []byte("garbage"), 0o644); err != nil {
		t.Fatal(err)
	}
	if tr, err := Inspect(context.Background(), root); err == nil {
		t.Fatalf("broken repo inspected as %+v", tr)
	}
}

func TestGetPath(t *testing.T) { // Y8
	tr := Tree{Root: "/r", Sub: "a/b"}
	ok := map[string]string{"out.bin": "a/b/out.bin", "../c": "a/c", "res/": "a/b/res/", "./x": "a/b/x"}
	for arg, want := range ok {
		if got, err := tr.GetPath(arg); err != nil || got != want {
			t.Errorf("GetPath(%q) = %q, %v; want %q", arg, got, err, want)
		}
	}
	for _, arg := range []string{"", "/etc/passwd", "../../..", "../../../x", "../../.git/config", "../../.git", "../.."} {
		if got, err := tr.GetPath(arg); err == nil {
			t.Errorf("GetPath(%q) = %q, want refusal", arg, got)
		}
	}
}

func TestWorkdirName(t *testing.T) { // Y2
	tr := Tree{Root: "/home/u/my proj:x"}
	if w := tr.Workdir(); !strings.HasPrefix(w, "offload/my_proj_x-") || len(w) != len("offload/my_proj_x-")+8 {
		t.Fatalf("Workdir = %q", w)
	}
	if tr.Meta() != "offload/.meta/"+strings.TrimPrefix(tr.Workdir(), "offload/") {
		t.Fatalf("Meta = %q", tr.Meta())
	}
	if (Tree{Root: "/a/x"}).Workdir() == (Tree{Root: "/b/x"}).Workdir() {
		t.Fatal("distinct roots share a workdir")
	}
}

func TestGlobalExcludes(t *testing.T) {
	root := repo(t)
	cfg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfg)
	if got, err := GlobalExcludes(context.Background(), root); err != nil || got != "" {
		t.Fatalf("none: %q %v", got, err)
	}
	write(t, filepath.Join(cfg, "git", "ignore"), "*.swp\n")
	if got, err := GlobalExcludes(context.Background(), root); err != nil || got != "*.swp\n" {
		t.Fatalf("xdg: %q %v", got, err)
	}
	custom := filepath.Join(t.TempDir(), "ex")
	write(t, custom, ".idea/\n")
	run(t, root, "git", "config", "core.excludesFile", custom)
	if got, err := GlobalExcludes(context.Background(), root); err != nil || got != ".idea/\n" {
		t.Fatalf("configured: %q %v", got, err)
	}
	write(t, custom, "*.tmp\x00junk\nkeep\x00\n\x00x\nlast") // git's pattern = the bytes before a NUL
	if got, err := GlobalExcludes(context.Background(), root); err != nil || got != "*.tmp\nkeep\n\nlast" {
		t.Fatalf("NUL: %q %v; want each line cut at its NUL", got, err)
	}
}

func TestRootsThatCopyEverythingStayLocal(t *testing.T) { // Y1: never mirror / or $HOME to a host
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, cwd := range []string{"/", home} {
		tr, err := Inspect(context.Background(), cwd)
		if err != nil || tr.LocalOnly == "" {
			t.Errorf("Inspect(%q) = %+v, %v; want LocalOnly", cwd, tr, err)
		}
	}
	for _, root := range []string{"", "relative", "/"} {
		err := UploadPlain(context.Background(), Tree{Root: root, Sub: "."}, remote.Target{Address: "nohost"})
		if err == nil || !strings.Contains(err.Error(), "refusing to sync") {
			t.Errorf("UploadPlain(root %q): %v; want a refusal before rsync", root, err)
		}
	}
}
