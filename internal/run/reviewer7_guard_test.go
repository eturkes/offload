package run

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// X2 + X8 / R9: a newly named variable must fail the reservation guard, whatever its assignment form.
func TestR7ReservationGuardFindsDeclarationsAndAppends(t *testing.T) {
	if seed := os.Getenv("OFFLOAD_R7_GUARD_SEED"); seed != "" {
		watchScript += "\n" + seed + "\n"
		TestScriptVariablesAreReserved(t)
		return
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, seed := range []string{"plain=1", "declare plain=1", "plain+=suffix"} {
		t.Run(seed, func(t *testing.T) {
			cmd := exec.Command(self, "-test.run=^TestR7ReservationGuardFindsDeclarationsAndAppends$")
			cmd.Env = append(os.Environ(), "OFFLOAD_R7_GUARD_SEED="+seed)
			out, err := cmd.CombinedOutput()
			var ee *exec.ExitError
			if !errors.As(err, &ee) || ee.ExitCode() != 1 || !strings.Contains(string(out), `watch.bash sets "plain"`) {
				t.Fatalf("reservation guard accepted %q or failed elsewhere: err=%v output=%q; want rejection naming plain", seed, err, out)
			}
		})
	}
}
