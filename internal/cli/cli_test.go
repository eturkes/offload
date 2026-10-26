package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// C1 + C4: help on stdout, rc 0.
func TestHelpPrintsUsageOnStdout(t *testing.T) {
	var out, errOut bytes.Buffer
	if rc := Main([]string{"--help"}, &out, &errOut); rc != 0 {
		t.Fatalf("rc = %d, want 0", rc)
	}
	if !strings.HasPrefix(out.String(), "Usage:") || errOut.Len() != 0 {
		t.Fatalf("stdout %q, stderr %q", out.String(), errOut.String())
	}
}

// C3 + C4: no command = a usage error (2) on stderr.
func TestNoArgumentsIsAUsageError(t *testing.T) {
	var out, errOut bytes.Buffer
	if rc := Main(nil, &out, &errOut); rc != ExitUsage {
		t.Fatalf("rc = %d, want %d", rc, ExitUsage)
	}
	if out.Len() != 0 || !strings.Contains(errOut.String(), "offload run") {
		t.Fatalf("stdout %q, stderr %q", out.String(), errOut.String())
	}
}

func TestParseRun(t *testing.T) { // C1 + C2
	ok := map[string]options{
		"-- make -j8":             {argv: []string{"make", "-j8"}},
		"make -j8":                {argv: []string{"make", "-j8"}},
		"echo --pool":             {argv: []string{"echo", "--pool"}},
		"--local -- x":            {local: true, argv: []string{"x"}},
		"--host nixbox x":         {pool: true, host: "nixbox", argv: []string{"x"}},
		"--gpu x":                 {pool: true, gpu: true, argv: []string{"x"}},
		"--no-sync x":             {pool: true, noSync: true, argv: []string{"x"}},
		"--get a --get b/ -- x y": {pool: true, gets: []string{"a", "b/"}, argv: []string{"x", "y"}},
		"--pool -- --weird":       {pool: true, argv: []string{"--weird"}},
	}
	for args, want := range ok {
		got, err := parseRun(strings.Fields(args))
		if err != nil || fmt.Sprint(got) != fmt.Sprint(want) {
			t.Errorf("parseRun(%q) = %+v, %v; want %+v", args, got, err, want)
		}
	}
	for _, args := range []string{
		"", "--", "--local --pool x", "--local --host h x", "--local --gpu x",
		"--local --no-sync x", "--local --get p x", "--bogus x", "--host",
	} {
		if got, err := parseRun(strings.Fields(args)); err == nil {
			t.Errorf("parseRun(%q) = %+v, want an error", args, got)
		}
	}
}

func TestParseCheckTakesOnlyNonceFramedLines(t *testing.T) { // K10
	out := "layer sha256:banner\noffload-check other 11\noffload-check n1 abc\nlayer n1 sha256:a\nnoise\nlayer n2 sha256:x\nlayer n1 sha256:b\n"
	sum, layers, ok := parseCheck(out, "n1")
	if !ok || sum != "abc" || fmt.Sprint(layers) != "[sha256:a sha256:b]" {
		t.Fatalf("parseCheck = %q %q %v", sum, layers, ok)
	}
	if _, _, ok := parseCheck("layer n1 sha256:a\n", "n1"); ok {
		t.Fatal("no offload-check line, yet ok")
	}
}

// C1 + C3 + C4: `status` takes `-h`/`--help` (usage on stdout, rc 0) and no other argument (rc 2 +
// the reason); an unknown command or flag = rc 2 with an `offload: ` reason, never the flag package's
// own unprefixed lines.
func TestUsageErrorsOfStatusCommandsAndFlags(t *testing.T) {
	t.Setenv("OFFLOAD_CONFIG", "/nonexistent/pool.toml") // a status that runs would fail with rc 2 too: the reason tells
	for _, c := range []struct {
		args           []string
		rc             int
		stdout, stderr string
	}{
		{[]string{"status", "-h"}, 0, "Usage:", ""},
		{[]string{"status", "--help"}, 0, "Usage:", ""},
		{[]string{"status", "extra"}, ExitUsage, "", "offload: status takes no arguments\n"},
		{[]string{"bogus"}, ExitUsage, "", "offload: unknown command \"bogus\"\n"},
		{[]string{"run", "--bogus", "--", "true"}, ExitUsage, "", "offload: flag provided but not defined: -bogus\n\nUsage:"},
	} {
		var out, errOut bytes.Buffer
		rc := Main(c.args, &out, &errOut)
		if rc != c.rc || !strings.HasPrefix(out.String(), c.stdout) || !strings.HasPrefix(errOut.String(), c.stderr) ||
			(c.stdout != "" && errOut.Len() != 0) || strings.Contains(errOut.String(), "Usage of") {
			t.Errorf("%q: rc %d stdout %.40q stderr %.120q", c.args, rc, out.String(), errOut.String())
		}
	}
}

// K8: `__ctr-init` reaches the container init (outside a container it refuses: rc 125 + why), not the
// unknown-command path.
func TestCtrInitIsDispatched(t *testing.T) {
	var out, errOut bytes.Buffer
	if rc := Main([]string{"__ctr-init", "true"}, &out, &errOut); rc != 125 || !strings.Contains(errOut.String(), "__ctr-init runs inside") {
		t.Fatalf("rc %d stderr %q", rc, errOut.String())
	}
}
