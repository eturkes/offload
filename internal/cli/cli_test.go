package cli

import (
	"bytes"
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
