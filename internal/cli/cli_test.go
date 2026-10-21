package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestHelpPrintsUsageOnStdout(t *testing.T) {
	var out, errOut bytes.Buffer
	if rc := Main([]string{"--help"}, &out, &errOut); rc != 0 {
		t.Fatalf("rc = %d, want 0", rc)
	}
	if !strings.HasPrefix(out.String(), "Usage:") || errOut.Len() != 0 {
		t.Fatalf("stdout %q, stderr %q", out.String(), errOut.String())
	}
}

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
