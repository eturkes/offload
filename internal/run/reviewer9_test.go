package run

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// X8: a directory at the run-script name → 125, nothing runs.
func TestReviewer9BootDestinationDirectoryFailsClosed(t *testing.T) {
	t.Parallel()
	for _, symlink := range []bool{false, true} {
		name := "directory"
		if symlink {
			name = "symlink-to-directory"
		}
		t.Run(name, func(t *testing.T) {
			h := newHost(t)
			dest := h.runFile()
			if symlink {
				if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(h.path("elsewhere"), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(h.path("elsewhere"), dest); err != nil {
					t.Fatal(err)
				}
			} else if err := os.MkdirAll(dest, 0o700); err != nil {
				t.Fatal(err)
			}
			rc, stderr := h.run("nosync", "review9", 0, nil, "touch ran", false)
			directory := dest
			if symlink {
				directory = h.path("elsewhere")
			}
			left, err := os.ReadDir(directory)
			if err != nil {
				t.Fatal(err)
			}
			if symlink && rc == 0 && h.exists("offload/w/ran") && len(left) == 0 {
				b, readErr := os.ReadFile(dest)
				if readErr == nil && string(b) == runScript {
					return // Replacing the symlink itself with the verified file is safe.
				}
			}
			if rc != 125 || h.exists("offload/w/ran") || len(left) != 0 {
				t.Fatalf("X8 setup failure: rc=%d (want 125 or verified replacement), CMD ran=%v, misplaced temp count=%d, stderr=%q", rc, h.exists("offload/w/ran"), len(left), stderr)
			}
		})
	}
}

// X1 + X6: TERM/HUP to the watcher = EOF; the status record still lands.
func TestReviewer9WatcherSignalsRecordStatus(t *testing.T) {
	t.Parallel()
	for _, sig := range []syscall.Signal{syscall.SIGTERM, syscall.SIGHUP} {
		t.Run(sig.String(), func(t *testing.T) {
			h := newHost(t)
			r := Remote{
				Workdir: "offload/w", Meta: "offload/.meta/w", Sub: ".", Mode: "nosync",
				Argv: []string{"bash", "-c", "echo $$ > pid; exec sleep 300"},
			}
			line, preamble := r.command("review9-status")
			c := exec.Command("bash", "-c", line)
			c.Dir, c.Env = h.home, h.env()
			var stderr bytes.Buffer
			c.Stderr = &stderr
			in, err := c.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			if err = c.Start(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = in.Close(); _ = c.Process.Kill() })
			if _, err = in.Write(preamble); err != nil {
				t.Fatal(err)
			}
			pid := readPids(t, h.path("offload/w/pid"), 1)[0]
			t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
			if err = c.Process.Signal(sig); err != nil {
				t.Fatal(err)
			}
			done := make(chan struct{})
			go func() { _ = c.Wait(); close(done) }()
			select {
			case <-done:
			case <-time.After(15 * time.Second):
				t.Fatal("watcher did not finish after signal")
			}
			status, err := os.ReadFile(h.path("offload/.meta/w/status"))
			if c.ProcessState.ExitCode() != 143 || err != nil || string(status) != "review9-status 143\n" || alive(pid) {
				t.Fatalf("signal=%v rc=%d status=%q err=%v leader alive=%v stderr=%q", sig, c.ProcessState.ExitCode(), status, err, alive(pid), stderr.String())
			}
		})
	}
}

// X2 + X8: caller variables named like script variables reach CMD unchanged.
func TestReviewer9RemoteKeepsCallerVariables(t *testing.T) {
	h := newHost(t)
	names := []string{"f", "d", "rest", "comp", "wd", "meta", "sub", "mode", "nonce", "count", "excludes", "cmd", "n", "first", "rc"}
	var want strings.Builder
	for _, name := range names {
		t.Setenv(name, "caller-"+name)
		want.WriteString(name + "=caller-" + name + "\n")
	}
	cmd := `for name in ` + strings.Join(names, " ") + `; do printf '%s=%s\n' "$name" "$(printenv "$name")"; done > vars`
	rc, stderr := h.run("nosync", "review9", 0, nil, cmd, false)
	b, err := os.ReadFile(h.path("offload/w/vars"))
	if err != nil || rc != 0 || string(b) != want.String() {
		t.Fatalf("rc=%d err=%v stderr=%q env=%q want=%q", rc, err, stderr, b, want.String())
	}
}
