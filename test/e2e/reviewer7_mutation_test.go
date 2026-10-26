package e2e_test

import (
	"syscall"
	"testing"
	"time"
)

// C3 + X6 + J2: an unambiguous returned status needs no second working SSH connection.
func TestR7ReturnedStatusSurvivesLaterTransportLoss(t *testing.T) {
	s := newSandbox(t, false)
	r := s.run("run", "--no-sync", "--", "sh", "-c", `touch "$HOME/.transport-fail"; exit 7`)
	r.requireCode(t, 7)
	_, end := s.runPair(time.Time{})
	requireField(t, end, "cmd_rc", 7.0)
}

// X1/X3 + J2: a terminal-style process-group INT must not bypass watcher teardown.
func TestR7GroupInterruptKeepsWatcherTeardown(t *testing.T) {
	s := newSandbox(t, false)
	s.markPIDs(s.root)
	p := s.start("run", "--local", "--", "bash", "-c", `printf '%s' $$ > command.pid; trap 'exit 17' TERM; printf ready; while :; do sleep 1; done`)
	p.await("command ready", func() bool { return p.output().stdout == "ready" })
	must(t, syscall.Kill(-p.cmd.Process.Pid, syscall.SIGINT))
	p.wait(25*time.Second).requireCode(t, 130)
	_, end := s.runPair(time.Time{})
	requireField(t, end, "cmd_rc", 17.0)
}
