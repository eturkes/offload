package ctr

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"
)

func reviewPodman(t *testing.T, script string) string {
	t.Helper()
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "podman"), []byte("#!/bin/sh\n"+script), 0o755))
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	return dir
}

func reviewProject(t *testing.T) Project {
	t.Helper()
	root := tree(t, map[string]string{".devcontainer/Containerfile": "FROM scratch\n"})
	binary, err := Executable()
	must(t, err)
	return Project{Root: root, Cwd: root, Context: filepath.Join(root, ".devcontainer"), Definition: filepath.Join(root, Definition), Binary: binary}
}

// fixedProject = run's lookup returning p as is (fields the tests set by hand).
func fixedProject(p Project) func(context.Context) (Project, error) {
	return func(context.Context) (Project, error) { return p, nil }
}

func reviewEnv(t *testing.T) func(string) string {
	t.Helper()
	root := t.TempDir()
	return func(k string) string {
		switch k {
		case "XDG_RUNTIME_DIR":
			return filepath.Join(root, "runtime")
		case "XDG_CACHE_HOME":
			return filepath.Join(root, "cache")
		default:
			return ""
		}
	}
}

func TestK3ReviewFind(t *testing.T) {
	p := reviewProject(t)
	git := exec.Command("git", "init", "-q", p.Root)
	if out, err := git.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	sub := filepath.Join(p.Root, "sub")
	must(t, os.Mkdir(sub, 0o700))
	link := filepath.Join(t.TempDir(), "link")
	must(t, os.Symlink(sub, link))
	got, err := Find(context.Background(), link)
	must(t, err)
	if got.Root != p.Root || got.Cwd != sub || got.Context != p.Context || got.Definition != p.Definition || got.Binary != p.Binary || got.CommonDir != "" {
		t.Fatalf("Find = %+v; root=%s cwd=%s context=%s binary=%s, no external git dir", got, p.Root, sub, p.Context, p.Binary)
	}
	must(t, os.Remove(p.Definition))
	_, err = Find(context.Background(), sub)
	var usage UsageError
	if !errors.As(err, &usage) || !strings.Contains(err.Error(), p.Definition) {
		t.Fatalf("missing definition = %v; want usage error naming %s", err, p.Definition)
	}
}

func TestK6ReviewExternalGitDir(t *testing.T) {
	p := reviewProject(t)
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", p.Root, "-c", "user.name=review", "-c", "user.email=review@example.invalid"}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	git("init", "-q")
	git("add", ".")
	git("commit", "-qm", "fixture")
	wt := filepath.Join(t.TempDir(), "linked")
	git("worktree", "add", "-q", "-b", "linked", wt)
	got, err := Find(context.Background(), wt)
	must(t, err)
	if got.Root != wt || got.CommonDir != filepath.Join(p.Root, ".git") {
		t.Fatalf("linked project=%+v; external common dir must be mounted", got)
	}
	for _, tc := range []struct {
		path string
		want bool
	}{{"/r", true}, {"/r/.git", true}, {"/r-other/.git", false}, {"/outside", false}} {
		if got := within("/r", tc.path); got != tc.want {
			t.Errorf("within /r %s = %t want %t", tc.path, got, tc.want)
		}
	}
}

func TestK5ReviewEnsure(t *testing.T) {
	for _, mode := range []string{"hit", "hit-runtime-blocked", "build", "built-by-other", "exists-error", "exists-error-once", "recheck-error", "build-error", "tag-error", "lock-error"} {
		t.Run(mode, func(t *testing.T) {
			p := reviewProject(t)
			state := t.TempDir()
			t.Setenv("REVIEW_STATE", state)
			t.Setenv("REVIEW_MODE", mode)
			reviewPodman(t, `case "$1 $2" in
"image exists")
  if [ -f "$REVIEW_STATE/queried" ]; then second=1; else second=0; fi
  touch "$REVIEW_STATE/queried"
  case "$REVIEW_MODE" in hit*) exit 0 ;; esac
  [ "$REVIEW_MODE" = built-by-other ] && [ "$second" = 1 ] && exit 0
  if [ "$REVIEW_MODE" = exists-error ] || { [ "$REVIEW_MODE" = exists-error-once ] && [ "$second" = 0 ]; } || { [ "$REVIEW_MODE" = recheck-error ] && [ "$second" = 1 ]; }; then
    echo storage-broken >&2; exit 42
  fi
  exit 1;;
"build --timestamp")
  printf '%s\n' "$@" > "$REVIEW_STATE/args"
  echo build-out; echo build-err >&2
  [ "$REVIEW_MODE" = build-error ] && exit 41
  exit 0;;
esac
exit 43
`)
			getenv := reviewEnv(t)
			if mode == "tag-error" {
				must(t, syscall.Mkfifo(filepath.Join(p.Context, "fifo"), 0o600))
			}
			if mode == "lock-error" || mode == "hit-runtime-blocked" {
				must(t, os.MkdirAll(filepath.Dir(getenv("XDG_RUNTIME_DIR")), 0o700))
				must(t, os.WriteFile(getenv("XDG_RUNTIME_DIR"), nil, 0o600))
			}
			var stderr bytes.Buffer
			got, err := Ensure(context.Background(), p, getenv, &stderr)
			failure := strings.Contains(mode, "error")
			if failure != (err != nil) {
				t.Fatalf("Ensure = %q, %v; mode=%s stderr=%q", got, err, mode, stderr.String())
			}
			args, readErr := os.ReadFile(filepath.Join(state, "args"))
			built := readErr == nil
			if want := mode == "build" || mode == "build-error"; built != want {
				t.Fatalf("build executed=%t want=%t; err=%v stderr=%q", built, want, err, stderr.String())
			}
			if !failure {
				if !strings.HasPrefix(got, "localhost/offload-ctr:") || len(got) != len("localhost/offload-ctr:")+16 {
					t.Fatalf("Ensure tag = %q", got)
				}
			}
			if mode == "build" {
				want := "build\n--timestamp\n0\n-t\n" + got + "\n-f\n" + p.Definition + "\n" + p.Context + "\n"
				if string(args) != want || stderr.String() != "offload: building "+got+" from "+p.Definition+"\n" {
					t.Fatalf("build args=%q stderr=%q", args, stderr.String())
				}
			}
			if mode == "build-error" && (!strings.Contains(stderr.String(), "build-out\n") || !strings.Contains(stderr.String(), "build-err\n")) {
				t.Fatalf("failed build output lost: %q", stderr.String())
			}
			if (strings.HasPrefix(mode, "exists-error") || mode == "recheck-error") && !strings.Contains(err.Error(), "storage-broken") {
				t.Fatalf("exists diagnostic lost: %v", err)
			}
		})
	}
}

func TestK5ReviewWaitCancellation(t *testing.T) {
	p := reviewProject(t)
	tag, err := Tag(p.Context)
	must(t, err)
	getenv := reviewEnv(t)
	path := filepath.Join(getenv("XDG_RUNTIME_DIR"), "offload", "ctr", strings.TrimPrefix(tag, Repo+":")+".lock")
	must(t, os.MkdirAll(filepath.Dir(path), 0o700))
	lock, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	must(t, err)
	defer func() { _ = lock.Close() }()
	must(t, syscall.Flock(int(lock.Fd()), syscall.LOCK_EX))
	reviewPodman(t, "[ \"$1 $2\" = 'image exists' ] && exit 1\nexit 42\n")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				done <- fmt.Errorf("panic: %v", p)
			}
		}()
		_, err := Ensure(ctx, p, getenv, new(bytes.Buffer))
		done <- err
	}()
	deadline := time.Now().Add(time.Second)
	for {
		fds, err := filepath.Glob("/proc/self/fd/*")
		must(t, err)
		n := 0
		for _, fd := range fds {
			if target, _ := os.Readlink(fd); target == path {
				n++
			}
		}
		if n >= 2 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("Ensure never opened the contended lock")
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("contended cancellation = %v, want context cancellation", err)
		}
	case <-time.After(time.Second):
		t.Fatal("contended lock ignored cancellation")
	}
}

func TestK6ReviewHomeFallback(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	got, err := Home(func(string) string { return "" }, "/project")
	must(t, err)
	if !strings.HasPrefix(got, filepath.Join(home, ".cache", "offload", "ctr")+"/") || !strings.HasSuffix(got, "/home") {
		t.Fatalf("fallback HOME=%q", got)
	}
	t.Setenv("HOME", "")
	if _, err := Home(func(string) string { return "" }, "/project"); err == nil {
		t.Fatal("missing HOME + XDG_CACHE_HOME must not create a relative cache")
	}
}

func TestK9ReviewLayers(t *testing.T) {
	for _, mode := range []string{"ordered", "empty", "failed"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("REVIEW_MODE", mode)
			reviewPodman(t, `case "$REVIEW_MODE" in
ordered) printf 'sha256:aaa\nsha256:bbb\n';;
empty) :;;
failed) printf 'sha256:partial\n'; exit 42;;
esac
`)
			got, err := Layers(context.Background(), "localhost/offload-ctr:0123456789abcdef")
			if mode == "failed" {
				if err == nil || len(got) != 0 {
					t.Fatalf("failed inspect returned %q, %v", got, err)
				}
				return
			}
			must(t, err)
			var want []string
			if mode == "ordered" {
				want = []string{"sha256:aaa", "sha256:bbb"}
			}
			if !slices.Equal(got, want) {
				t.Fatalf("layers=%q want=%q", got, want)
			}
		})
	}
}

func reviewRun(t *testing.T, p Project, getenv func(string) string, sigs <-chan os.Signal, stderr *bytes.Buffer) int {
	t.Helper()
	done := make(chan int, 1)
	go func() { done <- run(fixedProject(p), []string{"cmd", "a b", "--literal"}, getenv, sigs, stderr) }()
	select {
	case rc := <-done:
		return rc
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not return within 3s")
		return -1
	}
}

func TestK7ReviewRunBoundary(t *testing.T) {
	for _, mode := range []string{"record", "no-record-success", "no-record-failure", "ensure-error", "home-error", "colon", "binary-replaced", "launch-error"} {
		t.Run(mode, func(t *testing.T) {
			p := reviewProject(t)
			state := t.TempDir()
			t.Setenv("REVIEW_STATE", state)
			t.Setenv("REVIEW_MODE", mode)
			bin := reviewPodman(t, `case "$1 $2" in
"image exists") [ "$REVIEW_MODE" = ensure-error ] && exit 42; exit 0;;
run*)
  touch "$REVIEW_STATE/ran"
  case "$REVIEW_MODE" in
  no-record-success) exit 0;;
  no-record-failure) exit 41;;
  esac
  printf '7\n' >&3
  exit 42;;
esac
exit 43
`)
			getenv := reviewEnv(t)
			switch mode {
			case "home-error":
				must(t, os.MkdirAll(filepath.Dir(getenv("XDG_CACHE_HOME")), 0o700))
				must(t, os.WriteFile(getenv("XDG_CACHE_HOME"), nil, 0o600))
			case "colon":
				p.Root += ":x"
			case "binary-replaced":
				data, err := os.ReadFile(p.Binary)
				must(t, err)
				p.Binary = filepath.Join(state, "copy")
				must(t, os.WriteFile(p.Binary, data, 0o755))
			case "launch-error":
				base := getenv
				getenv = func(k string) string {
					if k == "XDG_CACHE_HOME" {
						must(t, os.Remove(filepath.Join(bin, "podman")))
						t.Setenv("PATH", bin)
					}
					return base(k)
				}
			}
			sigs := make(chan os.Signal)
			defer close(sigs)
			var stderr bytes.Buffer
			rc := reviewRun(t, p, getenv, sigs, &stderr)
			want := 125
			if mode == "record" {
				want = 7
				home, err := Home(getenv, p.Root)
				must(t, err)
				fi, err := os.Stat(home)
				must(t, err)
				if !fi.IsDir() || fi.Mode().Perm() != 0o700 {
					t.Fatalf("persistent HOME mode = %v; want directory 0700", fi.Mode())
				}
			}
			if rc != want {
				t.Fatalf("Run rc=%d want=%d stderr=%q", rc, want, stderr.String())
			}
			if want == 125 && !strings.Contains(stderr.String(), "offload:") {
				t.Fatalf("missing failure diagnostic: %q", stderr.String())
			}
			if mode == "ensure-error" || mode == "home-error" || mode == "colon" || mode == "binary-replaced" || mode == "launch-error" {
				if _, err := os.Stat(filepath.Join(state, "ran")); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("podman run started after setup failure: %v", err)
				}
			}
		})
	}
}

func TestK5ReviewSignalAfterEnsure(t *testing.T) {
	p := reviewProject(t)
	state := t.TempDir()
	t.Setenv("REVIEW_STATE", state)
	reviewPodman(t, "[ \"$1 $2\" = 'image exists' ] && exit 0\ntouch \"$REVIEW_STATE/ran\"\nprintf '7\\n' >&3\n")
	base := reviewEnv(t)
	sigs := make(chan os.Signal)
	defer close(sigs)
	getenv := func(k string) string {
		if k == "XDG_CACHE_HOME" {
			// Receipt of the second signal proves the first iteration set got + canceled setup.
			for range 2 {
				select {
				case sigs <- syscall.SIGTERM:
				case <-time.After(time.Second):
					t.Error("setup signal was not consumed")
				}
			}
		}
		return base(k)
	}
	var stderr bytes.Buffer
	if rc := reviewRun(t, p, getenv, sigs, &stderr); rc != 143 {
		t.Fatalf("setup signal rc=%d want=143 stderr=%q", rc, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(state, "ran")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("podman run started after setup signal: %v", err)
	}
}

func TestK8ReviewNotify(t *testing.T) {
	ch := Notify()
	must(t, syscall.Kill(os.Getpid(), syscall.SIGUSR1))
	select {
	case got := <-ch:
		if got != syscall.SIGUSR1 {
			t.Fatalf("notification=%v want USR1", got)
		}
	case <-time.After(time.Second):
		t.Fatal("USR1 notification missing")
	}
}

func TestK4ReviewRootModeAndUnreadable(t *testing.T) {
	p := tree(t, map[string]string{"a": "x"})
	before := tag(t, p)
	fi, err := os.Stat(p)
	must(t, err)
	must(t, os.Chmod(p, fi.Mode().Perm()^0o010))
	if got := tag(t, p); got != before {
		t.Fatal("context directory's own mode entered the below-context hash")
	}
	must(t, os.Chmod(filepath.Join(p, "a"), 0))
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(p, "a"), 0o644) })
	if _, err := Tag(p); err == nil {
		t.Fatal("unreadable regular file accepted")
	}
	must(t, os.Chmod(filepath.Join(p, "a"), 0o644))
	must(t, os.Mkdir(filepath.Join(p, "closed"), 0))
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(p, "closed"), 0o700) })
	if _, err := Tag(p); err == nil {
		t.Fatal("unreadable directory accepted")
	}
	must(t, os.Chmod(filepath.Join(p, "closed"), 0o700))
	must(t, os.WriteFile(filepath.Join(p, "closed", "file"), []byte("x"), 0o644))
	must(t, os.Chmod(filepath.Join(p, "closed"), 0o400))
	if _, err := Tag(p); err == nil {
		t.Fatal("entry below a non-searchable directory accepted")
	}
}

func TestK4ReviewSymlinkModesAreIgnored(t *testing.T) {
	file := filepath.Join(tree(t, map[string]string{"file": "x"}), "file")
	// procfs supplies real symlinks whose access bits can change without changing their targets.
	// The other process keeps its fd tree quiescent during each hash, unlike /proc/self/fd.
	cmd := exec.Command("bash", "--noprofile", "--norc", "-c", `set -eu
exec 3<"$1"
printf 'ready\n'
read -r _
exec 3<>"$1"
printf 'ready\n'
read -r _
`, "review", file)
	cmd.Env = append(os.Environ(), "BASH_ENV=")
	input, err := cmd.StdinPipe()
	must(t, err)
	output, err := cmd.StdoutPipe()
	must(t, err)
	must(t, cmd.Start())
	defer func() { _ = input.Close(); _ = cmd.Process.Kill(); _ = cmd.Wait() }()
	reader := bufio.NewReader(output)
	ready := func() {
		t.Helper()
		line, readErr := reader.ReadString('\n')
		must(t, readErr)
		if line != "ready\n" {
			t.Fatalf("child readiness = %q", line)
		}
	}
	ready()
	dir := fmt.Sprintf("/proc/%d/fd", cmd.Process.Pid)
	beforeMode, err := os.Lstat(filepath.Join(dir, "3"))
	must(t, err)
	before := tag(t, dir)
	_, err = input.Write([]byte("next\n"))
	must(t, err)
	ready()
	afterMode, err := os.Lstat(filepath.Join(dir, "3"))
	must(t, err)
	if beforeMode.Mode().Perm() == afterMode.Mode().Perm() {
		t.Fatal("positive control: symlink permission bits did not change")
	}
	if got := tag(t, dir); got != before {
		t.Fatalf("symlink mode alone changed the tag (%v → %v)", beforeMode.Mode(), afterMode.Mode())
	}
}

func TestK4ReviewEveryPermissionBit(t *testing.T) {
	p := tree(t, map[string]string{"a": "x"})
	base := tag(t, p)
	for _, bit := range []os.FileMode{1, 2, 4, 8, 16, 32, 64, 128, os.ModeSetuid, os.ModeSetgid, os.ModeSticky} {
		must(t, os.Chmod(filepath.Join(p, "a"), 0o644^bit))
		if got := tag(t, p); got == base {
			t.Errorf("permission bit %s omitted from tag", fmt.Sprint(bit))
		}
	}
}
