package mirror

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func interceptGit(t *testing.T, body string) {
	t.Helper()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	path := filepath.Join(bin, "git")
	write(t, path, "#!/bin/bash\n"+body+fmt.Sprintf("\nexec %q \"$@\"\n", realGit))
	if err := os.Chmod(path, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+":"+os.Getenv("PATH"))
}

func TestMirrorY1GitFailurePlainTree(t *testing.T) {
	interceptGit(t, "printf 'fatal: injected inspection failure\\n' >&2; exit 128")
	if tree, err := Inspect(context.Background(), t.TempDir()); err == nil || !strings.Contains(err.Error(), "injected inspection failure") {
		t.Fatalf("git failure became plain mode: tree=%+v err=%v", tree, err)
	}
}

func TestMirrorY1RevParseFraming(t *testing.T) {
	interceptGit(t, "printf 'root\\ngitdir\\ncommondir\\n'; exit 0")
	if tree, err := Inspect(context.Background(), t.TempDir()); err == nil {
		t.Fatalf("incomplete git response accepted: %+v", tree)
	}
}

func TestMirrorY1DamagedAncestor(t *testing.T) {
	root := repo(t)
	write(t, filepath.Join(root, ".git", "HEAD"), "broken\n")
	cwd := filepath.Join(root, "sub", "dir")
	if err := os.MkdirAll(cwd, 0o700); err != nil {
		t.Fatal(err)
	}
	if tree, err := Inspect(context.Background(), cwd); err == nil {
		t.Fatalf("damaged ancestor became plain mode: %+v", tree)
	}
}

func TestMirrorY1IndexFailure(t *testing.T) {
	root := repo(t)
	write(t, filepath.Join(root, ".git", "index"), "corrupt index\n")
	if tree, err := Inspect(context.Background(), root); err == nil {
		t.Fatalf("git ls-files failure accepted: %+v", tree)
	}
}

func TestMirrorY3SplitGitDirs(t *testing.T) {
	for _, layout := range []string{"common-outside", "git-dir-outside", "both-outside", "no-dotgit"} {
		t.Run(layout, func(t *testing.T) {
			root, other := repo(t), repo(t)
			dotgit, external := filepath.Join(root, ".git"), filepath.Join(other, ".git")
			gitDir, common := dotgit, external
			switch layout {
			case "git-dir-outside":
				gitDir, common = external, dotgit
			case "both-outside", "no-dotgit":
				gitDir, common = external, external
			}
			if layout == "no-dotgit" {
				if err := os.RemoveAll(dotgit); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("GIT_DIR", gitDir)
			t.Setenv("GIT_COMMON_DIR", common)
			t.Setenv("GIT_WORK_TREE", root)
			tree, err := Inspect(context.Background(), root)
			if err != nil || !tree.Git || tree.Root != root || tree.LocalOnly == "" {
				t.Fatalf("unsupported %s must remain local-eligible: tree=%+v err=%v", layout, tree, err)
			}
		})
	}
}

func TestMirrorY3NestedReplacesTracked(t *testing.T) {
	root := repo(t)
	nested := filepath.Join(root, "nested")
	write(t, nested, "tracked file\n")
	run(t, root, "git", "add", "nested")
	if err := os.Remove(nested); err != nil {
		t.Fatal(err)
	}
	run(t, root, "git", "init", "-q", "nested")
	write(t, filepath.Join(nested, "file"), "nested repository\n")
	write(t, filepath.Join(root, ".gitignore"), "nested/\n")
	if list, err := UploadList(context.Background(), root); !errors.Is(err, ErrLocalOnly) {
		t.Fatalf("nested repository occupying tracked path: list=%q err=%v", list, err)
	}
}

func TestMirrorY4ListErrors(t *testing.T) {
	for _, option := range []string{"-co", "-i"} {
		t.Run(option, func(t *testing.T) {
			root := repo(t)
			interceptGit(t, fmt.Sprintf("case \" $* \" in *\" %s \"*) printf 'fatal: injected list failure\\n' >&2; exit 128;; esac", option))
			if list, err := UploadList(context.Background(), root); err == nil || !strings.Contains(err.Error(), "injected list failure") {
				t.Fatalf("incomplete upload set accepted: list=%q err=%v", list, err)
			}
		})
	}
}

func TestMirrorY4LstatError(t *testing.T) {
	root := repo(t)
	dir := filepath.Join(root, "locked")
	path := filepath.Join(dir, "tracked")
	write(t, path, "tracked\n")
	run(t, root, "git", "add", "locked/tracked")
	if err := os.Chmod(dir, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Error(err)
		}
	})
	if _, err := os.Lstat(path); !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("fixture must deny lstat: %v", err)
	}
	if list, err := UploadList(context.Background(), root); !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("non-ENOENT lstat error lost: list=%q err=%v", list, err)
	}
}

func TestMirrorY4IgnoredSymlink(t *testing.T) {
	root := repo(t)
	write(t, filepath.Join(root, ".gitignore"), ".gitignore\n")
	write(t, filepath.Join(root, "child", "visible"), "keep\n")
	target := filepath.Join(t.TempDir(), "ignore")
	write(t, target, "visible\n")
	if err := os.Symlink(target, filepath.Join(root, "child", ".gitignore")); err != nil {
		t.Fatal(err)
	}
	list, err := UploadList(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(list, "child/.gitignore") || !slices.Contains(list, ".gitignore") || !slices.Contains(list, "child/visible") {
		t.Fatalf("ignored symlink is not a self-ignored rules file: %q", list)
	}
}

func TestMirrorY4AfterSymlinkAncestor(t *testing.T) {
	root := repo(t)
	write(t, filepath.Join(root, "a", "tracked"), "tracked\n")
	write(t, filepath.Join(root, "z"), "later single-letter path\n")
	run(t, root, "git", "add", ".")
	if err := os.RemoveAll(filepath.Join(root, "a")); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	write(t, filepath.Join(outside, "tracked"), "outside\n")
	if err := os.Symlink(outside, filepath.Join(root, "a")); err != nil {
		t.Fatal(err)
	}
	list, err := UploadList(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(list, "a") || !slices.Contains(list, "z") || slices.Contains(list, "a/tracked") {
		t.Fatalf("symlink ancestor must skip only its own descendants: %q", list)
	}
}

func TestMirrorY5ExcludesErrors(t *testing.T) {
	for _, failure := range []string{"git-config", "unreadable-file"} {
		t.Run(failure, func(t *testing.T) {
			root := repo(t)
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			if failure == "git-config" {
				interceptGit(t, "printf 'fatal: injected config failure\\n' >&2; exit 128")
			} else {
				run(t, root, "git", "config", "core.excludesFile", t.TempDir())
			}
			if excludes, err := GlobalExcludes(context.Background(), root); err == nil {
				t.Fatalf("failed exclusions silently became %q", excludes)
			}
		})
	}
}

func TestMirrorY3UnreadableMetadata(t *testing.T) {
	root := repo(t)
	if err := os.Chmod(root, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(root, 0o700); err != nil {
			t.Error(err)
		}
	})
	dotgit := filepath.Join(root, ".git")
	if _, err := os.Lstat(dotgit); !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("fixture must deny metadata access: %v", err)
	}
	if reason, err := localOnly(context.Background(), root, dotgit, dotgit); err == nil && reason == "" {
		t.Fatal("inaccessible metadata must not admit remote work")
	}
}
