// Package testenv holds prerequisites shared by test packages (imported from tests alone).
package testenv

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
)

// RequireScope skips where no user systemd manager starts a transient scope, the manager found as the
// watcher finds it ($XDG_RUNTIME_DIR, else /run/user/<uid>); OFFLOAD_REQUIRE_SCOPE=1 (CI with a
// lingering user manager) turns that absence into a failure, so scope tests never skip silently.
func RequireScope(t testing.TB) {
	t.Helper()
	if scopeWorks() {
		return
	}
	if os.Getenv("OFFLOAD_REQUIRE_SCOPE") != "" {
		t.Fatal("OFFLOAD_REQUIRE_SCOPE is set, but systemd-run --user --scope fails")
	}
	t.Skip("no user systemd manager: scope teardown is untestable here")
}

func scopeWorks() bool {
	for _, d := range []string{os.Getenv("XDG_RUNTIME_DIR"), "/run/user/" + strconv.Itoa(os.Getuid())} {
		if fi, err := os.Stat(filepath.Join(d, "systemd", "private")); d == "" || err != nil || fi.Mode()&os.ModeSocket == 0 { //nolint:gosec // G703: the test process's own runtime dirs
			continue
		}
		c := exec.Command("systemd-run", "--user", "--scope", "--quiet", "--collect", "--expand-environment=no", "--", "true")
		c.Env = append(os.Environ(), "XDG_RUNTIME_DIR="+d)
		return c.Run() == nil
	}
	return false
}
