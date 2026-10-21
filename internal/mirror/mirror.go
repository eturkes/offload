// Package mirror decides what crosses between the controller's tree and a host's workdir, and moves
// it with rsync: git's own file lists in a git work tree, rsync's .gitignore filter elsewhere.
package mirror

import (
	"bytes"
	"context"
	"crypto/sha1" //nolint:gosec // G505: workdir naming, not security
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/eturkes/offload/internal/remote"
)

// Tree = the sync root around the invocation cwd.
type Tree struct {
	Root string // absolute sync root
	Sub  string // cwd relative to Root ("." = Root)
	Git  bool
	// Worktree = a linked worktree's id (its git dir = CommonDir/worktrees/<id>); "" = main worktree.
	Worktree  string
	CommonDir string
	// LocalOnly != "" = why this tree never runs remotely (Y3).
	LocalOnly string
}

// Inspect finds the sync root of cwd (Y1) and its remote eligibility (Y3).
func Inspect(ctx context.Context, cwd string) (Tree, error) {
	out, err := git(ctx, cwd, "rev-parse", "--path-format=absolute", "--show-toplevel",
		"--git-dir", "--git-common-dir", "--show-prefix")
	if err != nil {
		var ge *gitError
		if !errors.As(err, &ge) || !strings.Contains(ge.stderr, "not a git repository") {
			return Tree{}, err
		}
		// git also says "not a git repository" for a damaged one: plain mode's --delete mirror
		// must not run over a tree that has a .git anywhere above it.
		for d := cwd; ; d = filepath.Dir(d) {
			if _, statErr := os.Lstat(filepath.Join(d, ".git")); statErr == nil {
				return Tree{}, fmt.Errorf("git does not recognize the repository at %s: %w", d, err)
			}
			if d == filepath.Dir(d) {
				return Tree{Root: cwd, Sub: ".", LocalOnly: copiesEverything(cwd)}, nil
			}
		}
	}
	f := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	if len(f) != 4 {
		return Tree{}, fmt.Errorf("git rev-parse printed %d lines, want 4", len(f))
	}
	t := Tree{Root: f[0], Sub: filepath.Clean(strings.TrimSuffix(f[3], "/")), Git: true}
	if t.Sub == "" {
		t.Sub = "."
	}
	if t.LocalOnly = copiesEverything(t.Root); t.LocalOnly != "" {
		return t, nil
	}
	t.LocalOnly, err = localOnly(ctx, t.Root, f[1], f[2])
	if t.LocalOnly == "" && err == nil && f[1] != f[2] {
		t.Worktree, t.CommonDir = filepath.Base(f[1]), f[2]
	}
	return t, err
}

// copiesEverything = why syncing root would ship far more than a project (Y1): `/` or the home
// directory (keys, credentials) never go to a host.
func copiesEverything(root string) string {
	if home, err := os.UserHomeDir(); err == nil && filepath.Clean(home) == filepath.Clean(root) {
		return "it is your home directory; syncing it would copy everything in it"
	}
	if root == "/" {
		return "it is the filesystem root"
	}
	return ""
}

// check = the root every transfer starts from: absolute and not `/`, whatever produced it.
func (t Tree) check() error {
	if !filepath.IsAbs(t.Root) || filepath.Clean(t.Root) == "/" {
		return fmt.Errorf("refusing to sync root %q", t.Root)
	}
	return nil
}

// worktreeID = git's linked-worktree ids that rsync filter rules + remote paths take verbatim.
var worktreeID = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func localOnly(ctx context.Context, root, gitDir, commonDir string) (string, error) {
	dotgit := filepath.Join(root, ".git")
	fi, err := os.Lstat(dotgit)
	switch {
	case gitDir != commonDir: // linked worktree: its .git is a gitfile into commonDir/worktrees/<id>
		if filepath.Dir(gitDir) != filepath.Join(commonDir, "worktrees") {
			return "its git dir lies outside the common dir's worktrees/", nil
		}
		if id := filepath.Base(gitDir); !worktreeID.MatchString(id) || id == "." || id == ".." {
			return fmt.Sprintf("its worktree id %q holds characters outside [A-Za-z0-9._-]", id), nil
		}
		if cfi, cerr := os.Stat(commonDir); cerr != nil || !cfi.IsDir() {
			return "its common git dir is not a directory", cerr
		}
	case errors.Is(err, fs.ErrNotExist) || (err == nil && !fi.IsDir()):
		return "its .git is not a directory (external git dir)", nil
	case err != nil:
		return "", err
	default:
		if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil || !same {
			return "its git dir lies outside the tree", sameErr
		}
	}
	if _, statErr := os.Stat(filepath.Join(commonDir, "objects", "info", "alternates")); statErr == nil {
		return "it borrows objects from another repository (alternates)", nil
	}
	staged, err := git(ctx, root, "ls-files", "-z", "--stage")
	if err != nil {
		return "", err
	}
	for _, e := range bytes.Split(staged, []byte{0}) {
		if bytes.HasPrefix(e, []byte("160000 ")) {
			return "it has submodules", nil
		}
	}
	return "", nil
}

func sameFile(a, b string) (bool, error) {
	fa, err := os.Stat(a)
	if err != nil {
		return false, err
	}
	fb, err := os.Stat(b)
	if err != nil {
		return false, err
	}
	return os.SameFile(fa, fb), nil
}

// ErrLocalOnly = the tree's file list holds a nested repository: remote runs are refused (Y3).
var ErrLocalOnly = errors.New("untracked nested git repository")

// UploadList = U (Y4): git's non-ignored files that exist + self-ignored .gitignore files in
// non-ignored directories, as root-relative paths.
func UploadList(ctx context.Context, root string) ([]string, error) {
	all, err := git(ctx, root, "ls-files", "-z", "-co", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	ignored, err := git(ctx, root, "ls-files", "-z", "-o", "-i", "--exclude-standard", "--directory")
	if err != nil {
		return nil, err
	}
	var list []string
	realDir := map[string]bool{".": true}
	for _, p := range splitZ(all) {
		if strings.HasSuffix(p, "/") {
			return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p)
		}
		if !inTree(root, p, realDir) {
			continue // reached through a symlink: not part of this tree
		}
		fi, err := os.Lstat(filepath.Join(root, p))
		switch {
		case errors.Is(err, fs.ErrNotExist):
			continue // deleted, not yet staged
		case err != nil:
			return nil, err
		case fi.IsDir():
			return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p)
		}
		list = append(list, p)
	}
	for _, p := range splitZ(ignored) {
		// Ignored directories arrive collapsed ("dir/") ⇒ a bare .gitignore entry sits in a
		// non-ignored directory, where git still reads it.
		if filepath.Base(p) == ".gitignore" && !strings.HasSuffix(p, "/") {
			if fi, err := os.Lstat(filepath.Join(root, p)); err == nil && fi.Mode().IsRegular() {
				list = append(list, p)
			}
		}
	}
	return list, nil
}

// inTree = every ancestor of p below root is a real directory (memoized in realDir).
func inTree(root, p string, realDir map[string]bool) bool {
	d := filepath.Dir(p)
	if ok, seen := realDir[d]; seen {
		return ok
	}
	ok := inTree(root, d, realDir)
	if ok {
		fi, err := os.Lstat(filepath.Join(root, d))
		ok = err == nil && fi.IsDir()
	}
	realDir[d] = ok
	return ok
}

// GlobalExcludes = the content of the controller's global git excludes file ("" when none).
func GlobalExcludes(ctx context.Context, root string) (string, error) {
	out, err := git(ctx, root, "config", "--path", "--get", "core.excludesFile")
	path := strings.TrimSpace(string(out))
	if err != nil {
		var ge *gitError
		if !errors.As(err, &ge) || ge.code != 1 { // 1 = unset
			return "", err
		}
		path = defaultExcludes()
	}
	data, err := os.ReadFile(path) //nolint:gosec // G304: git's own excludes path
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	return string(data), err
}

func defaultExcludes() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "git", "ignore")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "git", "ignore")
}

var unsafeName = regexp.MustCompile(`[^A-Za-z0-9._-]`)

// Workdir = the remote workdir (Y2), relative to the remote $HOME.
func (t Tree) Workdir() string { return "offload/" + t.name() }

// Meta = the workdir's control dir (lock, lists, status), relative to the remote $HOME.
func (t Tree) Meta() string { return "offload/.meta/" + t.name() }

func (t Tree) name() string {
	sum := sha1.Sum([]byte(t.Root)) //nolint:gosec // G401: naming, not security
	return unsafeName.ReplaceAllString(filepath.Base(t.Root), "_") + "-" + hex.EncodeToString(sum[:])[:8]
}

func splitZ(b []byte) []string {
	var out []string
	for _, e := range bytes.Split(b, []byte{0}) {
		if len(e) > 0 {
			out = append(out, string(e))
		}
	}
	return out
}

type gitError struct {
	args   []string
	code   int
	stderr string
}

func (e *gitError) Error() string {
	return fmt.Sprintf("git %s: %s", strings.Join(e.args, " "), strings.TrimSpace(e.stderr))
}

func git(ctx context.Context, dir string, args ...string) ([]byte, error) {
	//nolint:gosec // G204: fixed git subcommands; dir = the invocation tree
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return out, &gitError{args: args, code: ee.ExitCode(), stderr: stderr.String()}
	}
	return out, err
}

// rsyncPath = the remote rsync, run under the workdir's lock (Y11).
func rsyncPath(meta string) string {
	return "mkdir -p " + remote.Quote(meta) + " && flock -w 300 " + remote.Quote(meta+"/lock") + " rsync"
}

func rsync(ctx context.Context, stdin []byte, args ...string) error {
	cmd := remote.Command(ctx, "rsync", args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync: %w", err)
	}
	return nil
}

// base = every transfer's rsync flags. --modify-window=-1: rsync's quick check (size + mtime)
// otherwise compares whole seconds, so a same-size rewrite within the second of its last sync — a
// 41-byte ref after a quick commit, a quick edit — would never cross.
func base(target remote.Target, meta string) []string {
	return []string{"-a", "--modify-window=-1", "-e", target.RsyncShell(), "--rsync-path=" + rsyncPath(meta)}
}

// UploadGit sends list, then mirrors .git — a linked worktree's common dir to <meta>/repo (Y4). The list goes without --delete: a local file where
// the host keeps a directory of ignored files makes rsync fail (125) instead of deleting them.
// `.git` lock files never cross; stale remote ones go.
func UploadGit(ctx context.Context, t Tree, target remote.Target, list []string) error {
	if err := t.check(); err != nil {
		return err
	}
	var in bytes.Buffer
	for _, p := range list {
		in.WriteString(p)
		in.WriteByte(0)
	}
	files := append(base(target, t.Meta()), "--mkpath", "--from0", "--files-from=-",
		t.Root+"/", target.Address+":"+t.Workdir()+"/")
	if err := rsync(ctx, in.Bytes(), files...); err != nil {
		return err
	}
	dotgit := append(base(target, t.Meta()), "--mkpath", "--delete", "--delete-excluded",
		"--filter=- /*.lock", "--filter=- /**/*.lock")
	if t.Worktree == "" {
		dotgit = append(dotgit, t.Root+"/.git/", target.Address+":"+t.Workdir()+"/.git/")
	} else { // gitdir + commondir name controller paths: the run script writes the host's (Y4)
		wt := "/worktrees/" + t.Worktree + "/"
		dotgit = append(dotgit, "--filter=+ /worktrees/", "--filter=- "+wt+"gitdir", "--filter=- "+wt+"commondir",
			"--filter=+ "+wt+"***", "--filter=- /worktrees/*", t.CommonDir+"/", target.Address+":"+t.Meta()+"/repo/")
	}
	return rsync(ctx, nil, dotgit...)
}

// UploadPlain mirrors a non-git tree under rsync's .gitignore filter (Y7).
func UploadPlain(ctx context.Context, t Tree, target remote.Target) error {
	if err := t.check(); err != nil {
		return err
	}
	args := append(base(target, t.Meta()), "--mkpath", "--delete", "--filter=:- .gitignore",
		t.Root+"/", target.Address+":"+t.Workdir()+"/")
	return rsync(ctx, nil, args...)
}

// DownloadGit fetches the remote run's output list (Y6).
func DownloadGit(ctx context.Context, t Tree, target remote.Target) error {
	if err := t.check(); err != nil {
		return err
	}
	args := append(base(target, t.Meta()), "--from0", "--files-from=:"+t.Meta()+"/down", "--update",
		"--ignore-missing-args", target.Address+":"+t.Workdir()+"/", t.Root+"/")
	return rsync(ctx, nil, args...)
}

// DownloadPlain fetches a non-git tree's outputs under the .gitignore filter (Y7).
func DownloadPlain(ctx context.Context, t Tree, target remote.Target) error {
	if err := t.check(); err != nil {
		return err
	}
	args := append(base(target, t.Meta()), "--update", "--filter=:- .gitignore", "--exclude=/.git/",
		target.Address+":"+t.Workdir()+"/", t.Root+"/")
	return rsync(ctx, nil, args...)
}

// Get copies one root-relative path back whatever its ignore status (Y8); a trailing "/" = contents.
// A path without the slash lands in its parent directory, so fetching a directory twice never nests
// it inside itself.
func Get(ctx context.Context, t Tree, target remote.Target, rel string) error {
	if err := t.check(); err != nil {
		return err
	}
	dst := filepath.Join(t.Root, rel) + "/"
	if !strings.HasSuffix(rel, "/") {
		dst = filepath.Join(t.Root, filepath.Dir(rel)) + "/"
	}
	args := append(base(target, t.Meta()), "--update", "--mkpath", target.Address+":"+t.Workdir()+"/"+rel, dst)
	return rsync(ctx, nil, args...)
}

func trailingSlash(p string) string {
	if strings.HasSuffix(p, "/") {
		return "/"
	}
	return ""
}

// GetPath resolves a --get argument (invocation-cwd-relative) to a root-relative path, refusing
// absolute paths, escapes and .git (Y8).
func (t Tree) GetPath(arg string) (string, error) {
	if arg == "" || filepath.IsAbs(arg) {
		return "", fmt.Errorf("--get %q: give a path relative to the current directory", arg)
	}
	rel := filepath.Clean(filepath.Join(t.Sub, arg))
	if rel == "." || rel == ".." || strings.HasPrefix(rel, "../") {
		return "", fmt.Errorf("--get %q: path leaves the tree at %s", arg, t.Root)
	}
	if rel == ".git" || strings.HasPrefix(rel, ".git/") {
		return "", fmt.Errorf("--get %q: .git never comes back", arg)
	}
	return rel + trailingSlash(arg), nil
}
