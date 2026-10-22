package e2e_test

import (
	"errors"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestY1RootAndSubdirectory(t *testing.T) {
	s := newSandbox(t, true)
	s.write("sub/input", "subdir\n")
	s.cwd = filepath.Join(s.root, "sub")
	r := s.run("run", "--pool", "--", "bash", "-c", "printf '%s\\n' \"$PWD\"; cat ../tracked; cat input; printf result > result")
	r.requireCode(t, 0)
	if want := filepath.Join(s.workdir("alpha"), "sub") + "\ninitial\nsubdir\n"; r.stdout != want {
		t.Errorf("subdir/root mapping: got %q, want %q", r.stdout, want)
	}
	if got := s.read("sub/result"); got != "result" {
		t.Errorf("subdir output=%q", got)
	}
}

func TestY2WorkdirIdentity(t *testing.T) {
	s := newSandbox(t, true)
	for range 2 {
		r := s.run("run", "--pool", "--", "pwd")
		r.requireCode(t, 0)
		if want := s.workdir("alpha") + "\n"; r.stdout != want {
			t.Fatalf("workdir=%q, want %q", r.stdout, want)
		}
	}
	meta := filepath.Join(s.hosts, "alpha", "offload", ".meta", filepath.Base(s.workdir("alpha")))
	info, err := os.Stat(meta)
	must(t, err)
	if !info.IsDir() {
		t.Errorf("meta is not a directory: %s", meta)
	}
}

func TestY3UnsupportedGitLayouts(t *testing.T) {
	for _, layout := range []string{"worktree-odd-id", "submodule", "nested", "external-git-dir", "alternates"} {
		t.Run(layout, func(t *testing.T) {
			s := newSandbox(t, true)
			switch layout {
			case "worktree-odd-id": // id = the worktree dir's basename; `[` would act as an rsync filter glob
				linked := filepath.Join(s.base, "linked[1]")
				s.git("worktree", "add", "-qb", "linked", linked)
				s.root, s.cwd = linked, linked
			case "submodule":
				other := newSandbox(t, true)
				s.git("-c", "protocol.file.allow=always", "submodule", "add", "-q", other.root, "module")
			case "nested":
				s.git("init", "-q", "nested")
			case "external-git-dir":
				external := filepath.Join(s.base, "external-git")
				must(t, os.Rename(filepath.Join(s.root, ".git"), external))
				s.write(".git", "gitdir: "+external+"\n")
			case "alternates":
				other := newSandbox(t, true)
				s.write(".git/objects/info/alternates", filepath.Join(other.root, ".git", "objects")+"\n")
			}
			r := s.run("run", "--", "pwd")
			r.requireCode(t, 0)
			if r.stdout != s.root+"\n" {
				t.Errorf("unsupported layout did not run locally: %q", r.stdout)
			}
			r = s.run("run", "--pool", "--", "touch", "must-not-run")
			r.requireCode(t, 125)
			if strings.TrimSpace(r.stderr) == "" {
				t.Error("unsupported layout has no reason")
			}
			assertAbsent(t, s.workdir("alpha"), filepath.Join(s.root, "must-not-run"))
			calls, err := filepath.Glob(filepath.Join(s.hosts, "alpha", ".ssh-call-*"))
			must(t, err)
			if len(calls) != 0 {
				t.Errorf("remote work began before layout rejection: %v", calls)
			}
		})
	}
}

// Git, not an independently reimplemented ignore parser, defines Y4's oracle.
func (s *sandbox) uploadOracle() []string {
	s.t.Helper()
	set := map[string]bool{}
	for _, path := range nulStrings(s.git("ls-files", "-z", "-co", "--exclude-standard")) {
		if _, err := os.Lstat(filepath.Join(s.root, path)); err == nil {
			set[path] = true
		} else if !os.IsNotExist(err) {
			s.t.Fatal(err)
		}
	}
	err := filepath.WalkDir(s.root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(s.root, path)
		if err != nil {
			return err
		}
		if rel == ".git" {
			return filepath.SkipDir
		}
		if entry.IsDir() && rel != "." {
			check := exec.Command("git", "check-ignore", "-q", "--", rel+"/")
			check.Dir, check.Env = s.root, s.env
			checkErr := check.Run()
			if checkErr == nil {
				return filepath.SkipDir
			}
			var exit *exec.ExitError
			if !errors.As(checkErr, &exit) || exit.ExitCode() != 1 {
				return fmt.Errorf("git check-ignore %q: %w", rel, checkErr)
			}
		}
		if entry.Name() == ".gitignore" && !entry.IsDir() {
			set[rel] = true
		}
		return nil
	})
	must(s.t, err)
	paths := make([]string, 0, len(set))
	for path := range set {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	return paths
}

func TestY4GitUploadDifferential(t *testing.T) {
	for _, seed := range []uint64{1, 7, 91} {
		t.Run(fmt.Sprintf("seed-%d", seed), func(t *testing.T) {
			s := newSandbox(t, true)
			s.write(".gitignore", "*.skip\n!keep.skip\nignored-dir/\n")
			s.write("forced.skip", "tracked wins\n")
			s.git("add", "-f", "forced.skip", ".gitignore")
			s.write("keep.skip", "negative pattern\n")
			s.write("ignored.skip", "ignored\n")
			s.write("ignored-dir/secret", "ignored directory\n")
			s.write("ignored-dir/.gitignore", "!secret\n")
			s.write("nest/.gitignore", "*.cache\n!keep.cache\n")
			s.write("nest/no.cache", "ignored nested\n")
			s.write("nest/keep.cache", "nested reinclude\n")
			s.write("self/.gitignore", ".gitignore\nsecret\n")
			s.write("self/secret", "self ignored rules\n")
			s.write("self/keep", "visible\n")
			s.write("name with space", "space\n")
			s.write("name\nwith newline", "newline\n")
			must(t, os.Symlink("tracked", filepath.Join(s.root, "link")))
			must(t, os.Symlink("missing-target", filepath.Join(s.root, "dangling-link")))
			s.write(".git/info/exclude", "info-only\n")
			s.write("info-only", "git info excludes\n")
			excludes := filepath.Join(s.base, "global-excludes")
			s.write(excludes, "global-only\n")
			s.git("config", "--global", "core.excludesFile", excludes)
			s.write("global-only", "global excludes\n")
			rng := rand.New(rand.NewPCG(seed, seed+1))
			for i := range 16 {
				ext := []string{".txt", ".skip", ".cache"}[rng.IntN(3)]
				name := fmt.Sprintf("generated/%08x-%d%s", rng.Uint32(), i, ext)
				s.write(name, fmt.Sprintf("%d\n", rng.Uint64()))
				if i%3 == 0 {
					s.git("add", "-f", "--", name)
				}
			}
			s.git("add", "tracked")
			must(t, os.Remove(filepath.Join(s.root, "tracked")))
			s.write(".git/transient.lock", "must not cross\n")
			s.write(".git/refs/transient.lock", "nested lock\n")
			want := s.uploadOracle()
			r := s.run("run", "--pool", "--", "true")
			r.requireCode(t, 0)
			got := leafPaths(t, s.workdir("alpha"))
			if !slices.Equal(got, want) {
				t.Fatalf("upload differs from git oracle:\ngot  %q\nwant %q", got, want)
			}
			remote := s.workdir("alpha")
			assertAbsent(t, filepath.Join(remote, ".git", "transient.lock"), filepath.Join(remote, ".git", "refs", "transient.lock"))
			if got, err := os.Readlink(filepath.Join(remote, "link")); err != nil || got != "tracked" {
				t.Errorf("symlink not preserved: target=%q err=%v", got, err)
			}
			if s.read(filepath.Join(remote, ".git", "HEAD")) != s.read(".git/HEAD") {
				t.Error(".git was not mirrored")
			}
		})
	}
}

func TestY5GlobalExcludesReachTheHost(t *testing.T) { // Y5 + Y6 + X8: the excludes record governs R and D
	s := newSandbox(t, true)
	excludes := filepath.Join(s.base, "global-excludes")
	s.write(excludes, "*.tmp\n")
	s.git("config", "--global", "core.excludesFile", excludes)
	s.run("run", "--pool", "--", "touch", "out.tmp").requireCode(t, 0)
	remote := filepath.Join(s.workdir("alpha"), "out.tmp")
	if !exists(remote) {
		t.Fatal("the command's ignored output is missing on the host")
	}
	assertAbsent(t, s.path("out.tmp")) // ignored ⇒ never downloaded (Y6)
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	if !exists(remote) {
		t.Error("an output the global excludes ignore was pruned on the host (Y5)")
	}
}

func TestY4SameSecondRewriteCrosses(t *testing.T) { // Y4: rsync's quick check at nanosecond precision
	s := newSandbox(t, true)
	second := time.Now().Truncate(time.Second).Add(-time.Minute)
	stamp := func(at time.Duration) { must(t, os.Chtimes(s.path("tracked"), second.Add(at), second.Add(at))) }
	s.write("tracked", "aaaa\n")
	stamp(100 * time.Millisecond)
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	s.write("tracked", "bbbb\n") // same size, same second, other nanoseconds
	stamp(600 * time.Millisecond)
	r := s.run("run", "--pool", "--", "cat", "tracked")
	r.requireCode(t, 0)
	if r.stdout != "bbbb\n" {
		t.Fatalf("host ran on %q, want the rewrite", r.stdout)
	}
}

// sameStatRewrite = tools that stamp fixed mtimes (cargo vendor's archive times, tar --mtime) rewrite
// a file with its size + mtime unchanged ⇒ only content tells the versions apart (Y4, Y6, Y7).
func sameStatRewrite(t *testing.T, path, content string) {
	t.Helper()
	fi, err := os.Stat(path)
	must(t, err)
	must(t, os.WriteFile(path, []byte(content), 0o600))
	must(t, os.Chtimes(path, fi.ModTime(), fi.ModTime()))
}

func TestY4SameStatRewriteCrosses(t *testing.T) { // Y4: content decides, not size + mtime
	s := newSandbox(t, true)
	s.write("tracked", "aaaa\n")
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	sameStatRewrite(t, s.path("tracked"), "bbbb\n")
	r := s.run("run", "--pool", "--", "cat", "tracked")
	r.requireCode(t, 0)
	if r.stdout != "bbbb\n" {
		t.Fatalf("host ran on %q, want the rewrite", r.stdout)
	}
}

func TestY6SameStatRewriteReturns(t *testing.T) { // Y6: content decides, not size + mtime
	s := newSandbox(t, true)
	s.write("tracked", "aaaa\n")
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	s.run("run", "--pool", "--", "bash", "-c",
		`m=$(stat -c %y tracked); printf 'bbbb\n' >tracked; touch -d "$m" tracked`).requireCode(t, 0)
	if got := s.read("tracked"); got != "bbbb\n" {
		t.Fatalf("local holds %q after the download, want the host's rewrite", got)
	}
}

func TestY7SameStatRewriteCrosses(t *testing.T) { // Y7: content decides, not size + mtime
	s := newSandbox(t, false)
	s.write("file", "aaaa\n")
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	sameStatRewrite(t, s.path("file"), "bbbb\n")
	r := s.run("run", "--pool", "--", "cat", "file")
	r.requireCode(t, 0)
	if r.stdout != "bbbb\n" {
		t.Fatalf("host ran on %q, want the rewrite", r.stdout)
	}
}

// gitState = what a history-reading command sees in a checkout: branch, HEAD, history depth,
// staged + unstaged changes.
const gitState = `git rev-parse --abbrev-ref HEAD; git rev-parse HEAD; git rev-list --count HEAD
git diff --cached --name-only; git status --porcelain`

func TestY4LinkedWorktreeRunsRemotely(t *testing.T) {
	s := newSandbox(t, true)
	s.write("main-only", "x\n")
	s.git("add", "main-only")
	s.git("commit", "-qm", "main moves on")
	linked := filepath.Join(s.base, "linked")
	s.git("worktree", "add", "-qb", "feat", linked, "HEAD~1")
	s.git("worktree", "add", "-qb", "other", filepath.Join(s.base, "other")) // its admin dir stays home
	main := s.root
	s.root, s.cwd = linked, linked
	s.write("feature", "f\n")
	s.git("add", "feature")
	s.git("commit", "-qm", "feature commit")
	s.write("staged", "s\n")
	s.git("add", "staged")
	s.write("tracked", "edited, unstaged\n")
	want := s.shell(gitState)
	r := s.run("run", "--pool", "--", "bash", "-c", gitState)
	r.requireCode(t, 0)
	if r.stdout != want {
		t.Fatalf("host sees another checkout:\ngot  %q\nwant %q", r.stdout, want)
	}
	remote := s.workdir("alpha")
	if b, err := os.ReadFile(filepath.Join(remote, ".git")); err != nil || !strings.HasPrefix(string(b), "gitdir: ") {
		t.Fatalf("remote .git = %q, %v; want a gitfile", b, err)
	}
	admins, err := os.ReadDir(filepath.Join(filepath.Dir(remote), ".meta", filepath.Base(remote), "repo", "worktrees"))
	if err != nil || len(admins) != 1 || admins[0].Name() != "linked" {
		t.Fatalf("host worktree admin dirs = %v, %v; want [linked]", admins, err)
	}
	s.git("commit", "-qm", "second feature commit") // the next run follows the worktree's HEAD
	want = s.shell(gitState)
	r = s.run("run", "--pool", "--", "bash", "-c", gitState)
	r.requireCode(t, 0)
	if r.stdout != want {
		t.Fatalf("second run:\ngot  %q\nwant %q", r.stdout, want)
	}
	s.root, s.cwd = main, main // the main worktree of the same repo keeps its own remote checkout
	want = s.shell(gitState)
	r = s.run("run", "--pool", "--", "bash", "-c", gitState)
	r.requireCode(t, 0)
	if r.stdout != want {
		t.Fatalf("main worktree:\ngot  %q\nwant %q", r.stdout, want)
	}
}

func TestY4WorktreeRefusesAHostGitDirectory(t *testing.T) {
	s := newSandbox(t, true)
	linked := filepath.Join(s.base, "linked")
	s.git("worktree", "add", "-qb", "feat", linked)
	s.root, s.cwd = linked, linked
	must(t, os.MkdirAll(filepath.Join(s.workdir("alpha"), ".git"), 0o700)) // a main-repo copy on the host
	r := s.run("run", "--pool", "--", "touch", "must-not-run")
	r.requireCode(t, 125)
	requireContains(t, r.stderr, "holds a .git directory")
	assertAbsent(t, filepath.Join(s.workdir("alpha"), "must-not-run"))
}

func TestY4GitMetadataMirrorDeletes(t *testing.T) {
	s := newSandbox(t, true)
	s.write(".git/old-metadata", "obsolete\n")
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	must(t, os.Remove(filepath.Join(s.root, ".git", "old-metadata")))
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	assertAbsent(t, filepath.Join(s.workdir("alpha"), ".git", "old-metadata"))
}

func TestY5Prune(t *testing.T) {
	s := newSandbox(t, true)
	s.write(".gitignore", "cache/\n")
	s.write("remove-me", "tracked stale\n")
	s.git("add", ".")
	s.git("commit", "-qm", "prune fixture")
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	remote := s.workdir("alpha")
	s.write(filepath.Join(remote, "stale untracked"), "stale\n")
	s.write(filepath.Join(remote, "cache", "keep"), "remote cache\n")
	must(t, os.MkdirAll(filepath.Join(remote, "empty-directory"), 0o700))
	must(t, os.Remove(filepath.Join(s.root, "remove-me")))
	r := s.run("run", "--pool", "--", "bash", "-c", "test ! -e remove-me && test ! -e 'stale untracked' && test -f cache/keep && test -d empty-directory")
	r.requireCode(t, 0)
	assertAbsent(t, filepath.Join(remote, "remove-me"), filepath.Join(remote, "stale untracked"))
	if got := s.read(filepath.Join(remote, "cache", "keep")); got != "remote cache\n" {
		t.Errorf("ignored cache changed: %q", got)
	}
}

func TestY5PruneSymlinkAncestor(t *testing.T) {
	s := newSandbox(t, true)
	s.write("parent/victim", "local tracked\n")
	s.git("add", ".")
	s.git("commit", "-qm", "ancestor")
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	must(t, os.RemoveAll(filepath.Join(s.root, "parent")))
	remote := s.workdir("alpha")
	must(t, os.RemoveAll(filepath.Join(remote, "parent")))
	outside := filepath.Join(s.hosts, "alpha", "outside")
	s.write(filepath.Join(outside, "victim"), "outside must survive\n")
	must(t, os.Symlink(outside, filepath.Join(remote, "parent")))
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	if got := s.read(filepath.Join(outside, "victim")); got != "outside must survive\n" {
		t.Errorf("prune followed symlink ancestor: %q", got)
	}
}

func TestY6Download(t *testing.T) {
	for _, code := range []int{0, 2, 125, 255} {
		t.Run(fmt.Sprintf("command-%d", code), func(t *testing.T) {
			s := newSandbox(t, true)
			s.write(".gitignore", "ignored/\n")
			s.write("newer-local", "controller\n")
			s.write("gone", "delete remotely\n")
			s.git("add", ".")
			future := time.Now().Add(time.Hour)
			must(t, os.Chtimes(filepath.Join(s.root, "newer-local"), future, future))
			originalHead := s.read(".git/HEAD")
			command := "printf remote > tracked; printf new > created; printf older > newer-local; touch -t 200001010000 newer-local; rm gone; mkdir ignored; printf hidden > ignored/output; printf metadata > .git/remote-only; exit " + fmt.Sprint(code)
			r := s.run("run", "--pool", "--", "bash", "-c", command)
			r.requireCode(t, code)
			if s.read("tracked") != "remote" || s.read("created") != "new" {
				t.Error("modified/new outputs did not return")
			}
			if s.read("newer-local") != "controller\n" {
				t.Error("--update replaced newer regular local file")
			}
			if s.read("gone") != "delete remotely\n" {
				t.Error("download deleted local path")
			}
			assertAbsent(t, filepath.Join(s.root, "ignored", "output"), filepath.Join(s.root, ".git", "remote-only"))
			if s.read(".git/HEAD") != originalHead {
				t.Error("remote .git replaced local .git")
			}
		})
	}
	t.Run("symlink-and-type-change", func(t *testing.T) {
		s := newSandbox(t, true)
		s.write("regular-to-link", "regular\n")
		must(t, os.Symlink("tracked", filepath.Join(s.root, "link-to-regular")))
		must(t, os.Symlink("tracked", filepath.Join(s.root, "link-to-link")))
		future := time.Now().Add(time.Hour)
		must(t, os.Chtimes(filepath.Join(s.root, "regular-to-link"), future, future))
		r := s.run("run", "--pool", "--", "bash", "-c", "rm regular-to-link link-to-regular link-to-link; ln -s remote-target regular-to-link; printf remote > link-to-regular; ln -s different-target link-to-link")
		r.requireCode(t, 0)
		for path, want := range map[string]string{"regular-to-link": "remote-target", "link-to-link": "different-target"} {
			got, err := os.Readlink(filepath.Join(s.root, path))
			must(t, err)
			if got != want {
				t.Errorf("%s target=%q, want %q", path, got, want)
			}
		}
		if s.read("link-to-regular") != "remote" {
			t.Error("symlink → regular file was not copied")
		}
	})
}

func TestY7PlainMirror(t *testing.T) {
	s := newSandbox(t, false)
	s.write(".gitignore", "cache/\n*.skip\n")
	s.write("input", "controller\n")
	s.write("cache/local", "local cache\n")
	s.write("untracked.skip", "ignored\n")
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	remote := s.workdir("alpha")
	assertAbsent(t, filepath.Join(remote, "cache", "local"), filepath.Join(remote, "untracked.skip"))
	s.write(filepath.Join(remote, "stale"), "remove\n")
	s.write(filepath.Join(remote, "cache", "remote"), "remote cache\n")
	r := s.run("run", "--pool", "--", "bash", "-c", "test ! -e stale && test -f cache/remote && printf result > output && printf hidden > result.skip")
	r.requireCode(t, 0)
	if s.read("output") != "result" {
		t.Error("plain output did not return")
	}
	assertAbsent(t, filepath.Join(s.root, "cache", "remote"), filepath.Join(s.root, "result.skip"))
	if s.read("cache/local") != "local cache\n" {
		t.Error("local cache was changed")
	}
}

func TestY8Get(t *testing.T) {
	t.Run("ignored-and-directory", func(t *testing.T) {
		s := newSandbox(t, true)
		s.write(".gitignore", "ignored/\n")
		s.write("sub/input", "subdir\n")
		s.cwd = filepath.Join(s.root, "sub")
		r := s.run("run", "--get", "../ignored/one", "--get", "../ignored/tree/", "--", "bash", "-c", "mkdir -p ../ignored/tree/nested; printf one > ../ignored/one; printf two > ../ignored/tree/nested/two")
		r.requireCode(t, 0)
		if s.read("ignored/one") != "one" || s.read("ignored/tree/nested/two") != "two" {
			t.Error("--get lost ignored file/directory contents")
		}
		assertAbsent(t, filepath.Join(s.root, "ignored", "tree", "tree"))
	})
	for _, path := range []string{"../escape", "/absolute", ".git", ".git/config", "sub/../../escape", "sub/../.git/HEAD"} {
		t.Run("reject-"+strings.ReplaceAll(path, "/", "_"), func(t *testing.T) {
			s := newSandbox(t, true)
			r := s.run("run", "--get", path, "--", "touch", "must-not-run")
			r.requireCode(t, 2)
			assertAbsent(t, s.workdir("alpha"), filepath.Join(s.root, "must-not-run"))
		})
	}
	t.Run("missing", func(t *testing.T) {
		s := newSandbox(t, true)
		r := s.run("run", "--get", "missing", "--", "true")
		r.requireCode(t, 125)
		requireContains(t, r.stderr, "missing")
	})
}

func TestY9NoSync(t *testing.T) {
	s := newSandbox(t, true)
	remote := s.workdir("alpha")
	s.write(filepath.Join(remote, "remote-only"), "keep\n")
	r := s.run("run", "--no-sync", "--get", "explicit", "--", "bash", "-c", "test ! -e tracked && test -f remote-only && printf fetch > explicit && printf stay > automatic")
	r.requireCode(t, 0)
	if s.read("explicit") != "fetch" || s.read(filepath.Join(remote, "remote-only")) != "keep\n" {
		t.Error("--no-sync pruned or --get did not run")
	}
	assertAbsent(t, filepath.Join(remote, "tracked"), filepath.Join(remote, ".git"), filepath.Join(s.root, "automatic"))
}

func TestY10UploadFailure(t *testing.T) {
	s := newSandbox(t, true)
	s.write(filepath.Join(s.hosts, "alpha", ".upload-fail"), "fail\n")
	r := s.run("run", "--pool", "--", "touch", "must-not-run")
	r.requireCode(t, 125)
	requireContains(t, r.stderr, "injected upload failure")
	assertAbsent(t, filepath.Join(s.workdir("alpha"), "must-not-run"), filepath.Join(s.root, "must-not-run"))
}

func TestY11RemoteLock(t *testing.T) {
	s := newSandbox(t, true)
	remote := s.workdir("alpha")
	meta := filepath.Join(s.hosts, "alpha", "offload", ".meta", filepath.Base(remote))
	must(t, os.MkdirAll(meta, 0o700))
	lock, err := os.OpenFile(filepath.Join(meta, "lock"), os.O_CREATE|os.O_RDWR, 0o600)
	must(t, err)
	must(t, syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB))
	t.Cleanup(func() { must(t, lock.Close()) })
	p := s.start("run", "--pool", "--", "touch", "command-ran")
	p.await("remote connection", func() bool {
		calls, err := filepath.Glob(filepath.Join(s.hosts, "alpha", ".ssh-call-*"))
		must(t, err)
		return len(calls) >= 2
	})
	// Holding the specified lock models a predecessor after its controller died.
	time.Sleep(300 * time.Millisecond)
	assertAbsent(t, filepath.Join(remote, "tracked"), filepath.Join(remote, "command-ran"))
	select {
	case <-p.done:
		t.Fatalf("successor failed instead of waiting: %+v", p.wait(time.Second))
	default:
	}
	must(t, syscall.Flock(int(lock.Fd()), syscall.LOCK_UN))
	p.wait(8*time.Second).requireCode(t, 0)
	if !exists(filepath.Join(remote, "command-ran")) || !exists(filepath.Join(remote, "tracked")) {
		t.Error("successor did not proceed after predecessor released remote lock")
	}
}
