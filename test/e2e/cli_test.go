package e2e_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestC1FlagBoundary(t *testing.T) {
	for _, boundary := range []string{"separator", "first-word"} {
		t.Run(boundary, func(t *testing.T) {
			s := newSandbox(t, false)
			args := []string{"run", "--local"}
			if boundary == "separator" {
				args = append(args, "--")
			}
			args = append(args, "printf", "%s\n", "--pool", "--host", "not-a-host", "--gpu", "--", "$HOME", "two words")
			r := s.run(args...)
			r.requireCode(t, 0)
			if want := "--pool\n--host\nnot-a-host\n--gpu\n--\n$HOME\ntwo words\n"; r.stdout != want {
				t.Fatalf("argv boundary: got %q, want %q", r.stdout, want)
			}
		})
	}
	for _, verb := range []string{"run", "status"} {
		for _, flag := range []string{"-h", "--help"} {
			t.Run(verb+flag, func(t *testing.T) {
				s := newSandbox(t, false)
				r := s.run(verb, flag)
				r.requireCode(t, 0)
				if r.stdout+r.stderr == "" {
					t.Error("help is empty")
				}
			})
		}
	}
}

func TestC2ConflictsAndPoolImplications(t *testing.T) {
	flags := [][]string{{"--pool"}, {"--host", "alpha"}, {"--gpu"}, {"--no-sync"}, {"--get", "artifact"}}
	for _, flag := range flags {
		t.Run("conflict-"+flag[0], func(t *testing.T) {
			s := newSandbox(t, true)
			args := append([]string{"run", "--local"}, flag...)
			r := s.run(append(args, "--", "bash", "-c", "touch must-not-run")...)
			r.requireCode(t, 2)
			assertAbsent(t, filepath.Join(s.root, "must-not-run"), s.workdir("alpha"))
		})
		t.Run("implies-pool-"+flag[0], func(t *testing.T) {
			s := newSandbox(t, true, host{name: "alpha", cpus: 8, load: 2, gpus: 1})
			args := append([]string{"run"}, flag...)
			r := s.run(append(args, "--", "bash", "-c", "printf artifact > artifact; printf '%s' \"$HOME\"")...)
			r.requireCode(t, 0)
			if r.stdout != filepath.Join(s.hosts, "alpha") {
				t.Fatalf("flag did not imply pool: HOME=%q", r.stdout)
			}
		})
	}
}

func TestC3ExitPrecedence(t *testing.T) {
	for _, place := range []string{"--local", "--pool"} {
		for _, code := range []int{0, 2, 42, 125, 255} {
			t.Run(fmt.Sprintf("%s/%d", place, code), func(t *testing.T) {
				s := newSandbox(t, false)
				r := s.run("run", place, "--", "bash", "-c", fmt.Sprintf("printf command-ran; exit %d", code))
				r.requireCode(t, code)
				if r.stdout != "command-ran" {
					t.Errorf("command did not execute: stdout=%q stderr=%q", r.stdout, r.stderr)
				}
			})
		}
		for _, signal := range []struct {
			name string
			code int
		}{{"INT", 130}, {"TERM", 143}, {"HUP", 129}} {
			t.Run(place+"/command-"+signal.name, func(t *testing.T) {
				s := newSandbox(t, false)
				s.run("run", place, "--", "bash", "-c", "kill -"+signal.name+" $$").requireCode(t, signal.code)
			})
		}
		t.Run(place+"/not-found", func(t *testing.T) {
			s := newSandbox(t, false)
			s.run("run", place, "--", "offload-e2e-command-that-does-not-exist").requireCode(t, 127)
			if place == "--local" {
				assertAbsent(t, filepath.Join(s.runtime, "heavy", "slot1"), filepath.Join(s.runtime, "heavy", "slot2"))
			}
		})
		t.Run(place+"/not-executable", func(t *testing.T) {
			s := newSandbox(t, false)
			s.write("noexec", "#!/bin/sh\nexit 0\n")
			s.run("run", place, "--", "./noexec").requireCode(t, 126)
		})
	}
	t.Run("bad-pool-before-command", func(t *testing.T) {
		s := newSandbox(t, false)
		s.write(s.config, "not toml")
		s.run("run", "--", "touch", "must-not-run").requireCode(t, 2)
		assertAbsent(t, filepath.Join(s.root, "must-not-run"), filepath.Join(s.runtime, "heavy", "slot1"))
	})
	t.Run("unknown-host-before-command", func(t *testing.T) {
		s := newSandbox(t, false)
		s.run("run", "--host", "unknown", "--", "touch", "must-not-run").requireCode(t, 2)
		assertAbsent(t, s.workdir("alpha"), filepath.Join(s.root, "must-not-run"))
	})
}

func TestC4OutputChannels(t *testing.T) {
	for _, place := range []string{"--local", "--pool"} {
		t.Run(place, func(t *testing.T) {
			s := newSandbox(t, true)
			r := s.run("run", place, "--", "bash", "-c", "printf 'command-out\n'; printf 'command-err\n' >&2")
			r.requireCode(t, 0)
			if r.stdout != "command-out\n" {
				t.Errorf("stdout=%q", r.stdout)
			}
			if strings.Count(r.stderr, "command-err\n") != 1 {
				t.Errorf("command stderr=%q", r.stderr)
			}
			for _, line := range strings.Split(strings.TrimSuffix(strings.ReplaceAll(r.stderr, "command-err\n", ""), "\n"), "\n") {
				if !strings.HasPrefix(line, "offload: ") {
					t.Errorf("offload message lacks prefix: %q", line)
				}
			}
		})
	}
}

func TestC5PlacementLine(t *testing.T) {
	for _, place := range []string{"--local", "--pool"} {
		t.Run(place, func(t *testing.T) {
			s := newSandbox(t, false)
			s.holdSlot("", 1)
			s.holdSlot("alpha", 1)
			r := s.run("run", place, "--", "bash", "-c", "printf 'COMMAND-START\n' >&2")
			r.requireCode(t, 0)
			want := "offload: local slot 2/2\n"
			if place == "--pool" {
				want = "offload: alpha slot 2/2 (8 CPUs, 25% busy) → ~/offload/" + filepath.Base(s.workdir("alpha")) + "\n"
			}
			placement, started := strings.Index(r.stderr, want), strings.Index(r.stderr, "COMMAND-START\n")
			if placement < 0 || started < placement+len(want) {
				t.Fatalf("placement before command: got %q, want %q before COMMAND-START", r.stderr, want)
			}
		})
	}
}

func TestP3MissingPool(t *testing.T) {
	for _, mode := range []string{"run", "pool", "status"} {
		t.Run(mode, func(t *testing.T) {
			s := newSandbox(t, false)
			must(t, os.Remove(s.config))
			args := []string{"run", "--", "printf", "local"}
			want := 0
			switch mode {
			case "pool":
				args, want = []string{"run", "--pool", "--", "true"}, 2
			case "status":
				args, want = []string{"status"}, 2
			}
			r := s.run(args...)
			r.requireCode(t, want)
			requireContains(t, r.stderr, s.config)
			if mode == "run" {
				if r.stdout != "local" {
					t.Errorf("stdout=%q", r.stdout)
				}
				if strings.Count(r.stderr, s.config) != 1 {
					t.Errorf("missing-file note must occur once: %q", r.stderr)
				}
				requireContains(t, r.stderr, "offload: local slot")
			}
		})
	}
}

func TestR2Status(t *testing.T) {
	s := newSandbox(t, true, host{name: "alpha", cpus: 8, load: 2, gpus: 1}, host{name: "beta", down: true})
	s.holdSlot("", 2)
	s.holdSlot("alpha", 1)
	s.write(filepath.Join(s.runtime, "heavy", "slot1.owner"), "99999 since 01:02:03: stale-owner\n")
	s.write(filepath.Join(s.runtime, "heavy", "slot2.owner"), "12345 since 01:02:03: held-local\n")
	s.write(filepath.Join(s.runtime, "offload", "hosts", "alpha", "slot1.owner"), "12345 since 01:02:03: alpha: root held-remote\n")
	r := s.run("status")
	r.requireCode(t, 0)
	requireContains(t, r.stdout, "1/2", "held-local", "alpha", "beta", "down", "Connection refused")
	if strings.Contains(r.stdout, "stale-owner") {
		t.Error("unlocked stale owner reported as held")
	}
	if strings.Index(r.stdout, "alpha") > strings.Index(r.stdout, "beta") {
		t.Errorf("pool-file order lost: %q", r.stdout)
	}
	var alpha string
	for _, line := range strings.Split(r.stdout, "\n") {
		if strings.Contains(line, "alpha") {
			alpha = line
			break
		}
	}
	for _, pattern := range []string{`1/2`, `\b8\b`, `\b25(?:\.0+)?%?`, `\b1\b`} {
		if !regexp.MustCompile(pattern).MatchString(alpha) {
			t.Errorf("status lacks CPUs/busy/GPU/slots (%s): %q", pattern, alpha)
		}
	}
}

func TestL1Eligibility(t *testing.T) {
	for _, mode := range []string{"default", "local", "pool"} {
		t.Run(mode, func(t *testing.T) {
			s := newSandbox(t, true)
			args := []string{"run"}
			if mode != "default" {
				args = append(args, "--"+mode)
			}
			if mode == "local" {
				s.write(s.config, "invalid pool must not be read")
			}
			r := s.run(append(args, "--", "pwd")...)
			r.requireCode(t, 0)
			want := s.root
			if mode == "pool" {
				want = s.workdir("alpha")
			}
			if strings.TrimSpace(r.stdout) != want {
				t.Fatalf("cwd=%q, want %q", r.stdout, want)
			}
		})
	}
	t.Run("local-full-falls-back", func(t *testing.T) {
		s := newSandbox(t, true)
		s.holdSlot("", 1)
		s.holdSlot("", 2)
		r := s.run("run", "--", "pwd")
		r.requireCode(t, 0)
		if strings.TrimSpace(r.stdout) != s.workdir("alpha") {
			t.Fatalf("did not fall back to pool: %q", r.stdout)
		}
	})
}

func TestL2Placement(t *testing.T) {
	for _, tc := range []struct {
		name     string
		alpha    float64
		beta     float64
		busy     bool
		explicit bool
		want     string
	}{
		{name: "lowest-ratio", alpha: 3, beta: 4, want: "beta"},
		{name: "tie-pool-order", alpha: 2, beta: 8, want: "alpha"},
		{name: "busy-best-host", alpha: 0, beta: 8, busy: true, want: "beta"},
		{name: "explicit-host", alpha: 0, beta: 8, explicit: true, want: "beta"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newSandbox(t, false, host{name: "alpha", cpus: 2, load: tc.alpha}, host{name: "beta", cpus: 8, load: tc.beta})
			if tc.busy {
				s.holdSlot("alpha", 1)
				s.holdSlot("alpha", 2)
			}
			args := []string{"run", "--pool"}
			if tc.explicit {
				args = append(args, "--host", "beta")
			}
			r := s.run(append(args, "--", "pwd")...)
			r.requireCode(t, 0)
			if strings.TrimSpace(r.stdout) != s.workdir(tc.want) {
				t.Errorf("placement=%q, want %q", r.stdout, s.workdir(tc.want))
			}
		})
	}
}

func TestL3WaitingSignals(t *testing.T) {
	for _, signal := range []struct {
		name string
		sig  syscall.Signal
		code int
	}{{"INT", syscall.SIGINT, 130}, {"TERM", syscall.SIGTERM, 143}, {"HUP", syscall.SIGHUP, 129}} {
		t.Run(signal.name, func(t *testing.T) {
			s := newSandbox(t, false)
			s.holdSlot("", 1)
			s.holdSlot("", 2)
			s.holdSlot("alpha", 1)
			s.holdSlot("alpha", 2)
			p := s.start("run", "--", "touch", "must-not-run")
			p.await("waiting message", func() bool { return strings.Contains(p.output().stderr, "offload: waiting") })
			p.signal(signal.sig)
			r := p.wait(3 * time.Second)
			r.requireCode(t, signal.code)
			if strings.Count(r.stderr, "offload: waiting") != 1 {
				t.Errorf("waiting message count: %q", r.stderr)
			}
			requireContains(t, r.stderr, "2/2", "alpha")
			assertAbsent(t, filepath.Join(s.root, "must-not-run"), s.workdir("alpha"))
		})
	}
	t.Run("release-retries", func(t *testing.T) {
		s := newSandbox(t, false)
		s.holdSlot("", 1)
		release := s.holdSlot("", 2)
		p := s.start("run", "--local", "--", "printf", "admitted")
		p.await("waiting message", func() bool { return strings.Contains(p.output().stderr, "offload: waiting") })
		release()
		r := p.wait(5 * time.Second)
		r.requireCode(t, 0)
		if r.stdout != "admitted" || strings.Count(r.stderr, "offload: waiting") != 1 {
			t.Errorf("retry output: %+v", r)
		}
	})
}

func TestL4NoGPU(t *testing.T) {
	s := newSandbox(t, false, host{name: "alpha", cpus: 8}, host{name: "beta", cpus: 4})
	r := s.run("run", "--gpu", "--", "touch", "must-not-run")
	r.requireCode(t, 125)
	assertAbsent(t, filepath.Join(s.root, "must-not-run"), filepath.Join(s.workdir("alpha"), "must-not-run"), filepath.Join(s.workdir("beta"), "must-not-run"))
}
