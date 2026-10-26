package nixbox_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCRR2EFIReadFailureStopsBeforeChanges(t *testing.T) {
	for _, fail := range []bool{false, true} {
		name := "positive-control"
		if fail {
			name = "listing-fails"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			fake, log := filepath.Join(dir, "efibootmgr"), filepath.Join(dir, "changes")
			body := "#!/bin/sh\nif [ $# = 0 ]; then\n"
			if fail {
				body += "  printf 'injected EFI listing failure\\n' >&2\n  exit 23\n"
			} else {
				body += "  printf 'Boot0001* Linux Boot Manager\\tHD(1,GPT,old)\\n'\n  exit 0\n"
			}
			body += "fi\nprintf '%s\\n' \"$*\" >>\"$EFI_CHANGE_LOG\"\n"
			if err := os.WriteFile(fake, []byte(body), 0o700); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("sh", "./efi-entry", "/dev/fake-disk")
			cmd.Env = append(os.Environ(), "EFIBOOTMGR="+fake, "EFI_CHANGE_LOG="+log, "DRY_RUN=")
			out, err := cmd.CombinedOutput()
			changes, readErr := os.ReadFile(log)
			if fail {
				if !strings.Contains(string(out), "injected EFI listing failure") {
					t.Fatalf("failure not reached: %q, %v", out, err)
				}
				if err == nil || len(changes) != 0 {
					t.Fatalf("EFI discovery failed but install mutated entries: err=%v changes=%q output=%q", err, changes, out)
				}
				if !errors.Is(readErr, os.ErrNotExist) {
					t.Fatalf("mutation log must not exist: %v", readErr)
				}
			} else if err != nil || readErr != nil || !strings.Contains(string(changes), "-q -b 0001 -B\n") || !strings.Contains(string(changes), "-q -c -d /dev/fake-disk -p 1") {
				t.Fatalf("positive control: stale deletion + new entry missing: err=%v readErr=%v changes=%q output=%q", err, readErr, changes, out)
			}
		})
	}
}
