package e2e_test

import (
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// J2/X3: a status returned by the remote watcher remains known when offload itself was signaled.
func TestReviewR1J2CancellationKeepsReturnedCommandStatus(t *testing.T) {
	for _, place := range []string{"--local", "--pool"} {
		t.Run(place, func(t *testing.T) {
			s := newSandbox(t, true)
			p := s.start("run", place, "--", "bash", "-c", "trap 'exit 17' TERM; printf ready; while :; do sleep 1; done")
			p.await("command ready", func() bool { return p.output().stdout == "ready" })
			p.signal(syscall.SIGTERM)
			p.wait(15*time.Second).requireCode(t, 143)
			if place == "--pool" {
				status := s.read(filepath.Join(s.hosts, "alpha", "offload", ".meta", filepath.Base(s.workdir("alpha")), "status"))
				if !strings.HasSuffix(status, " 17\n") {
					t.Fatalf("remote watcher did not return the fixture's 17: %q", status)
				}
			}
			_, end := s.runPair(time.Time{})
			if end["cmd_rc"] != 17.0 {
				t.Fatalf("%s returned command status was discarded: cmd_rc=%v, rc=%v, signal=%v; want cmd_rc=17", place, end["cmd_rc"], end["rc"], end["signal"])
			}
		})
	}
}
