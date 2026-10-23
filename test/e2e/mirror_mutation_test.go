package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMirrorY10FileUploadFailure(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	// An existing mirror keeps the later prune from masking a lost upload error.
	s.run("run", "--pool", "--", "true").requireCode(t, 0)
	realRsync, err := exec.LookPath("rsync")
	must(t, err)
	shim := filepath.Join(s.base, "bin", "rsync")
	s.write(shim, fmt.Sprintf("#!/bin/bash\nfor arg do if [[ $arg == --files-from=- ]]; then printf 'injected file-upload failure\\n' >&2; exit 23; fi; done\nexec %q \"$@\"\n", realRsync))
	must(t, os.Chmod(shim, 0o700))
	r := s.run("run", "--pool", "--", "touch", "must-not-run")
	r.requireCode(t, 125)
	requireContains(t, r.stderr, "injected file-upload failure")
	assertAbsent(t, filepath.Join(s.root, "must-not-run"), filepath.Join(s.workdir("alpha"), "must-not-run"))
}
